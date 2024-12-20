package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chmistdawid/container-daemon/cont_image"
)

const socketPath = "/var/run/cont.sock"

func main() {
	setupLogging()

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket file: %v", err)
	}

	go startUnixSocketServer()

	go handleSignals()

	cont_image.DownloadImage()

	select {}
}

func startUnixSocketServer() {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to start Unix socket listener: %v", err)
	}
	defer listener.Close()

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

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Connection established: %v", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Printf("Received message: %s", msg)
		parseMessage(msg, conn)
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

func setupLogging() {
	file, err := os.OpenFile("daemon.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(file)
}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	log.Println("Daemon shutting down")
	cleanup()
	os.Exit(0)
}

func cleanup() {
	log.Println("Cleaning up resources...")
	if err := os.Remove(socketPath); err != nil {
		log.Printf("Failed to remove socket file: %v", err)
	}
}

func parseMessage(msg string, conn net.Conn) {
	if strings.HasPrefix(msg, "start") {
		command := strings.Split(msg, " ")
		if len(strings.Split(msg, " ")) != 2 {
			log.Printf("Invalid message: %s", msg)
			return
		}
		_, err := conn.Write([]byte("starting container " + command[1] + "\n"))
		if err != nil {
			log.Printf("Failed to write response: %v", err)
			return
		}
	}
	if strings.HasPrefix(msg, "stop") {
		command := strings.Split(msg, " ")
		if len(strings.Split(msg, " ")) != 2 {
			log.Printf("Invalid message: %s", msg)
			return
		}
		_, err := conn.Write([]byte("stopping container " + command[1] + "\n"))
		if err != nil {
			log.Printf("Failed to write response: %v", err)
			return
		}
	}
}

func runContainer(containerID string) {
	log.Printf("Running container %s", containerID)
	image_dir := cont_image.DownloadImage()
	if image_dir == "" {
		log.Printf("Failed to download image")
		return
	}
	log.Printf("Image directory: %s", image_dir)
}
