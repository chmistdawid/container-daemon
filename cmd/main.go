package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Step 1: Detach the process (if necessary)
	if os.Getppid() != 1 {
		// Detach by forking the process
		if forkDaemon() {
			return
		}
	}

	// Step 2: Initialize daemon-specific setup
	setupLogging()

	// Step 3: Handle signals for clean shutdown
	go handleSignals()

	// Step 4: Main work loop
	log.Println("Daemon started")
	for {
		doWork()
		time.Sleep(1 * time.Second) // Simulate periodic work
	}
}

// forkDaemon forks the process to detach it from the terminal.
func forkDaemon() bool {
	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	proc, err := os.StartProcess(os.Args[0], os.Args, attr)
	if err != nil {
		log.Fatalf("Failed to fork daemon: %v", err)
	}
	log.Printf("Daemon process started with PID %d", proc.Pid)
	return true
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

// doWork represents the main task of the daemon.
func doWork() {
	log.Println("Daemon is working...")
	// Add your task here
}

// cleanup performs any necessary cleanup before exiting.
func cleanup() {
	log.Println("Cleaning up resources...")
	// Add cleanup code here
}
