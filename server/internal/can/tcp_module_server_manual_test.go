package can

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const manualModuleAddr = "127.0.0.1:9000"

//RUN_TCP_MODULE_SERVER=1 BACKEND_URL="http://127.0.0.1:8080" \
//GOTOOLCHAIN=local GOCACHE=/private/tmp/go-build-cache-can-server \
//go test ./internal/can -run TestManualTCPModuleServer -v

func TestManualTCPModuleServer(t *testing.T) {
	if os.Getenv("RUN_TCP_MODULE_SERVER") != "1" {
		t.Skip("set RUN_TCP_MODULE_SERVER=1 to start the long-running TCP module simulator")
	}

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://127.0.0.1:8080"
	}

	listener, err := net.Listen("tcp", manualModuleAddr)
	if err != nil {
		t.Fatalf("listen on %s: %v", manualModuleAddr, err)
	}
	defer listener.Close()

	fmt.Printf("TCP module simulator listening on %s\n", manualModuleAddr)
	fmt.Printf("Backend API URL: %s\n", backendURL)
	fmt.Println("Commands:")
	fmt.Println("  send <CAN_ID_HEX> <8_BYTE_DATA_HEX>")
	fmt.Println("  example: send A1 0001020304050607")
	fmt.Println("  example: send B1 000A000000000000")
	fmt.Println("Press Ctrl+C to stop.")

	go acceptModuleConnections(listener)
	readManualCommands(backendURL)
}

func acceptModuleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handleModuleConnection(conn)
	}
}

func handleModuleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		payload := make([]byte, 9)
		if _, err := io.ReadFull(conn, payload); err != nil {
			if err != io.EOF {
				fmt.Printf("read tcp frame failed: %v\n", err)
			}
			return
		}
		fmt.Printf(
			"received from backend: can_id=0x%02X data=% X raw=% X\n",
			payload[0],
			payload[1:],
			payload,
		)
	}
}

func readManualCommands(backendURL string) {
	input := io.Reader(os.Stdin)
	tty, err := os.Open("/dev/tty")
	if err == nil {
		defer tty.Close()
		input = tty
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 || strings.ToLower(fields[0]) != "send" {
			fmt.Println("invalid command, use: send <CAN_ID_HEX> <8_BYTE_DATA_HEX>")
			continue
		}
		if err := postFrameToBackend(backendURL, fields[1], fields[2]); err != nil {
			fmt.Printf("send to backend failed: %v\n", err)
			continue
		}
		fmt.Printf("sent to backend: can_id=0x%s data=%s\n", strings.ToUpper(fields[1]), strings.ToUpper(fields[2]))
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("read command failed: %v\n", err)
	}
	fmt.Println("command input closed; TCP module simulator is still running. Press Ctrl+C to stop.")
	select {}
}

func postFrameToBackend(backendURL, canID, data string) error {
	body, err := json.Marshal(map[string]string{
		"can_id": canID,
		"data":   data,
	})
	if err != nil {
		return err
	}

	url := strings.TrimRight(backendURL, "/") + "/api/v1/can/frames"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	return nil
}
