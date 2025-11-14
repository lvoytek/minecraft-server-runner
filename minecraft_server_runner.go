// Copyright 2025 Lena Voytek

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
)

func main() {
	app := kingpin.New("minecraft-server-runner", "Runs a Minecraft server and streams output over TCP")
	serverPath := app.Flag("server-path", "Path to the Minecraft server directory").Required().String()
	port := app.Flag("port", "TCP port to listen on for connections").Default("25566").Int()

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Start TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", *port, err)
	}
	defer listener.Close()

	log.Printf("Listening for connections on port %d", *port)

	// Channel to manage connected clients
	var (
		clients   []*net.Conn
		clientsMu sync.Mutex
		outputCh  = make(chan []byte, 100)
	)

	// Start the Minecraft server
	cmd := exec.Command(
		"java",
		"-Xmx8192M",
		"-Xms128M",
		"-jar", "server.jar",
		"nogui",
	)
	cmd.Dir = *serverPath

	// Get stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Start the server process
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Minecraft server started")

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Read stdout
	go readOutput(stdout, outputCh)
	// Read stderr
	go readOutput(stderr, outputCh)

	// Broadcast output to all connected clients
	go func() {
		for data := range outputCh {
			clientsMu.Lock()
			for _, conn := range clients {
				if conn != nil {
					_, _ = (*conn).Write(data)
				}
			}
			clientsMu.Unlock()
		}
	}()

	// Accept client connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}

			clientsMu.Lock()
			clients = append(clients, &conn)
			clientsMu.Unlock()

			log.Printf("Client connected: %s (total: %d)", conn.RemoteAddr(), len(clients))
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")

	// Terminate the server process
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		log.Printf("Failed to signal server: %v", err)
	}

	// Close all client connections
	clientsMu.Lock()
	for _, conn := range clients {
		if conn != nil {
			(*conn).Close()
		}
	}
	clientsMu.Unlock()

	close(outputCh)

	// Wait for server to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("Server exited with error: %v", err)
	}

	log.Println("Shutdown complete")
}

// readOutput reads from a pipe and sends lines to the output channel
func readOutput(reader io.Reader, outputCh chan []byte) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := append(scanner.Bytes(), '\n')
		outputCh <- line
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading output: %v", err)
	}
}
