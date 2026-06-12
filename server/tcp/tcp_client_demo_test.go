package tcp

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/user/can-server/internal/can"
)

const defaultTCPDemoAddr = "192.168.1.253:8000"

// RUN_TCP_CLIENT_DEMO=1 GOTOOLCHAIN=local GOCACHE=/private/tmp/go-build-cache-can-server \
// go test ./tcp -run TestTCPClientDemo -v
func TestTCPClientDemo(t *testing.T) {
	if os.Getenv("RUN_TCP_CLIENT_DEMO") != "1" {
		t.Skip("set RUN_TCP_CLIENT_DEMO=1 to connect to the TCP module and print frames")
	}

	addr := strings.TrimSpace(os.Getenv("TCP_DEMO_ADDR"))
	if addr == "" {
		addr = defaultTCPDemoAddr
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("connect to %s: %v", addr, err)
	}
	defer conn.Close()

	fmt.Printf("connected to TCP module: %s\n", addr)
	fmt.Println("reading 13-byte frames: 08 + 4-byte CAN ID + 8-byte data")
	fmt.Println("press Ctrl+C to stop.")

	frame := make([]byte, 13)
	for {
		if _, err := io.ReadFull(conn, frame); err != nil {
			t.Fatalf("read tcp frame from %s failed: %v", addr, err)
		}

		canID, data, err := can.ParseTCPPayload(frame)
		now := time.Now().Format("2006-01-02 15:04:05.000")
		if err != nil {
			fmt.Printf("[%s] invalid raw=%s err=%v\n", now, strings.ToUpper(hex.EncodeToString(frame)), err)
			continue
		}

		fmt.Printf(
			"[%s] raw=%s can_id=0x%02X data=%s\n",
			now,
			strings.ToUpper(hex.EncodeToString(frame)),
			canID,
			strings.ToUpper(hex.EncodeToString(data[:])),
		)
	}
}
