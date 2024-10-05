package main

import (
	"context"
	"io"
	"net"

	"github.com/Dudeiebot/dlog"
)

var logger = dlog.NewLog(dlog.LevelTrace)

func main() {
	listener, err := net.Listen("tcp", ":44400")
	if err != nil {
		logger.Log(context.Background(), dlog.LevelFatal, "", "Error", err)
		return
	}
	defer listener.Close()

	logger.Info("Echo server listening on :44400")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection")
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logger.Error("Error reading from connection", "Error", err)
			}
			return
		}

		received := string(buffer[:n])
		logger.Info("", "Received", received)

		_, err = conn.Write(buffer[:n])
		if err != nil {
			logger.Error("Error waiting for connection")
			return
		}
	}
}
