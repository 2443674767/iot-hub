package can

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const manualModuleAddr = "127.0.0.1:9000"

// RUN_TCP_MODULE_SERVER=1 GOTOOLCHAIN=local GOCACHE=/private/tmp/go-build-cache-can-server \
// go test ./internal/can -run TestManualTCPModuleServer -v
func TestManualTCPModuleServer(t *testing.T) {
	if os.Getenv("RUN_TCP_MODULE_SERVER") != "1" {
		t.Skip("set RUN_TCP_MODULE_SERVER=1 to start the long-running TCP module simulator")
	}

	listener, err := net.Listen("tcp", manualModuleAddr)
	if err != nil {
		t.Fatalf("listen on %s: %v", manualModuleAddr, err)
	}
	defer listener.Close()

	clients := newManualModuleClients()
	fmt.Printf("TCP module simulator listening on %s\n", manualModuleAddr)
	fmt.Println("Configure the backend active TCP target to this address.")
	fmt.Println("Commands:")
	fmt.Println("  send <CAN_ID_HEX> <8_BYTE_DATA_HEX>")
	fmt.Println("  example: send A1 0702580000000000")
	fmt.Println("  example: send A2 0702580000000000")
	fmt.Println("Press Ctrl+C to stop.")

	go acceptModuleConnections(listener, clients)
	readManualCommands(clients)
}

type manualModuleClients struct {
	mu    sync.Mutex
	conns map[net.Conn]struct{}
}

func newManualModuleClients() *manualModuleClients {
	return &manualModuleClients{conns: make(map[net.Conn]struct{})}
}

func (c *manualModuleClients) add(conn net.Conn) {
	c.mu.Lock()
	c.conns[conn] = struct{}{}
	c.mu.Unlock()
}

func (c *manualModuleClients) remove(conn net.Conn) {
	c.mu.Lock()
	delete(c.conns, conn)
	c.mu.Unlock()
}

func (c *manualModuleClients) write(payload []byte) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for conn := range c.conns {
		n, err := conn.Write(payload)
		if err != nil || n != len(payload) {
			_ = conn.Close()
			delete(c.conns, conn)
			continue
		}
		count++
	}
	return count
}

func acceptModuleConnections(listener net.Listener, clients *manualModuleClients) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		clients.add(conn)
		fmt.Printf("backend connected: %s\n", conn.RemoteAddr())
		go handleModuleConnection(conn, clients)
	}
}

func handleModuleConnection(conn net.Conn, clients *manualModuleClients) {
	defer func() {
		clients.remove(conn)
		_ = conn.Close()
		fmt.Printf("backend disconnected: %s\n", conn.RemoteAddr())
	}()

	for {
		payload := make([]byte, 13)
		if _, err := io.ReadFull(conn, payload); err != nil {
			if err != io.EOF {
				fmt.Printf("read tcp frame failed: %v\n", err)
			}
			return
		}
		canID, data, err := ParseTCPPayload(payload)
		if err != nil {
			fmt.Printf("invalid tcp frame: %v raw=% X\n", err, payload)
			continue
		}
		fmt.Printf(
			"received from backend: dlc=0x08 can_id=0x%02X data=% X raw=% X\n",
			canID,
			data,
			payload,
		)
	}
}

func readManualCommands(clients *manualModuleClients) {
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
		raw, err := buildManualRawFrame(fields[1], fields[2])
		if err != nil {
			fmt.Printf("invalid frame: %v\n", err)
			continue
		}
		payload, err := hex.DecodeString(raw)
		if err != nil {
			fmt.Printf("decode frame failed: %v\n", err)
			continue
		}
		count := clients.write(payload)
		fmt.Printf("sent to backend connections=%d raw=%s\n", count, strings.ToUpper(raw))
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("read command failed: %v\n", err)
	}
	fmt.Println("command input closed; TCP module simulator is still running. Press Ctrl+C to stop.")
	select {}
}

func buildManualRawFrame(canIDHex, dataHex string) (string, error) {
	canID, err := parseManualCANID(canIDHex)
	if err != nil {
		return "", err
	}
	data, err := parseManualCANData(dataHex)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(BuildTCPPayload(canID, data)), nil
}

func parseManualCANID(raw string) (byte, error) {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	n, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return 0, fmt.Errorf("CAN ID must be one byte hex")
	}
	return byte(n), nil
}

func parseManualCANData(raw string) ([8]byte, error) {
	var data [8]byte
	value := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return data, fmt.Errorf("data must be hex")
	}
	if len(decoded) != 8 {
		return data, fmt.Errorf("data must be 8 bytes")
	}
	copy(data[:], decoded)
	return data, nil
}
