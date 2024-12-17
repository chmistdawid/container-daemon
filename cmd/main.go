package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const socketPath = "/var/run/cont.sock"

func main() {
	// Detach process if needed (not covered here for brevity)
	setupLogging()

	// Clean up the socket file if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket file: %v", err)
	}

	// Create and start the Unix socket server
	go startUnixSocketServer()

	// Handle signals for clean shutdown
	go handleSignals()

	// Block main to keep the daemon alive
	select {}
}

// startUnixSocketServer starts a Unix domain socket server.
func startUnixSocketServer() {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to start Unix socket listener: %v", err)
	}
	defer listener.Close()

	// Ensure the socket file has appropriate permissions
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Failed to set permissions on socket file: %v", err)
	}

	log.Printf("Unix socket server started at %s", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection handles incoming connections to the Unix socket.
func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Connection established: %v", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Printf("Received message: %s", msg)

		// Echo the message back to the client (optional)
		_, err := conn.Write([]byte("Received: " + msg + "\n"))
		if err != nil {
			log.Printf("Failed to write response: %v", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}

	log.Printf("Connection closed: %v", conn.RemoteAddr())
}

// setupLogging configures logging for the daemon.
func setupLogging() {
	file, err := os.OpenFile("daemon.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(file)
}

// handleSignals listens for system signals to cleanly shutdown.
func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	log.Println("Daemon shutting down")
	cleanup()
	os.Exit(0)
}

// cleanup performs any necessary cleanup before exiting.
func cleanup() {
	log.Println("Cleaning up resources...")
	if err := os.Remove(socketPath); err != nil {
		log.Printf("Failed to remove socket file: %v", err)
	}
}
