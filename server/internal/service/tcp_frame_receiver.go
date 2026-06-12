package service

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/can"
)

type TCPFrameReceiver struct {
	tcpConfigs  TCPConfigProvider
	frames      *CANFrameService
	dialTimeout time.Duration
	retryDelay  time.Duration
}

func NewTCPFrameReceiver(cfg *config.Config, tcpConfigs TCPConfigProvider, frames *CANFrameService) *TCPFrameReceiver {
	return &TCPFrameReceiver{
		tcpConfigs:  tcpConfigs,
		frames:      frames,
		dialTimeout: cfg.CAN.TCPTimeout(),
		retryDelay:  2 * time.Second,
	}
}

func (r *TCPFrameReceiver) Start() {
	go func() {
		for {
			tcpCfg, err := r.tcpConfigs.GetActive()
			if err != nil {
				log.Printf("can tcp receiver load active config failed: %v", err)
				time.Sleep(r.retryDelay)
				continue
			}

			addr := net.JoinHostPort(tcpCfg.Host, fmt.Sprintf("%d", tcpCfg.Port))
			log.Printf("can tcp receiver connecting active tcp config: id=%d name=%s addr=%s", tcpCfg.ID, tcpCfg.Name, addr)
			conn, err := net.DialTimeout("tcp", addr, r.dialTimeout)
			if err != nil {
				log.Printf("can tcp receiver connect failed: addr=%s err=%v", addr, err)
				time.Sleep(r.retryDelay)
				continue
			}

			log.Printf("can tcp receiver connected: addr=%s", addr)
			r.handleConn(addr, conn)
			time.Sleep(r.retryDelay)
		}
	}()
}

func (r *TCPFrameReceiver) handleConn(addr string, conn net.Conn) {
	defer conn.Close()
	for {
		payload := make([]byte, 13)
		if _, err := io.ReadFull(conn, payload); err != nil {
			if err != io.EOF {
				log.Printf("can tcp receiver read failed from %s: %v", addr, err)
			}
			return
		}

		canID, data, err := can.ParseTCPPayload(payload)
		if err != nil {
			log.Printf("can tcp receiver invalid frame from %s: %v raw=% X", addr, err, payload)
			continue
		}

		start := time.Now()
		if err := r.frames.RecordReceivedAndBroadcast(canID, data); err != nil {
			log.Printf("can tcp receiver persist failed: can_id=0x%02X data=% X err=%v", canID, data, err)
			continue
		}
		log.Printf("can tcp receiver stored and broadcast: can_id=0x%02X data=% X elapsed=%s", canID, data, time.Since(start))
	}
}

func FormatTCPListenAddr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
