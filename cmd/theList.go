package main

import (
	"io"
	"log"
	"os"
	"theListBot/internal/server"
)

/*
The list bot hooks into a discord server and is used to send gifs that are associated with 2 character codes
The bot will be able to send gifs to a channel based on the 2 character code that is sent to it from users
The bot will also be able to add new gifs to the list of gifs that it can send by allowing users to create code associations to gif links
*/

func main() {
	// Configure logging to be more selective
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create log file for persistent logging
	logFile, err := os.OpenFile("thelistbot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// Write logs to both stderr and the log file
		multiWriter := io.MultiWriter(os.Stderr, logFile)
		log.SetOutput(multiWriter)
	} else {
		log.Println("Failed to open log file, using console logging only:", err)
	}

	// Check for config directory override
	if configPath := os.Getenv("GIFLIST_CONFIG_PATH"); configPath == "" {
		// If not set, use the default location
		homeDir, err := os.UserHomeDir()
		if err == nil {
			defaultPath := homeDir + "/.thelistbot"
			log.Printf("Using default config path: %s", defaultPath)
		}
	} else {
		log.Printf("Using config path from environment: %s", configPath)
	}

	log.Println("Starting theListBot...")

	// Create a new server
	server := server.NewServer()

	// Start server (this will block until shutdown signal)
	server.Start()

	log.Println("Exiting theListBot")
}
