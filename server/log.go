package main

import (
	"io"
	"log"
	"os"
)

var (
	valid_events []string
)

func check(valid []string, el string) bool {
	for _, v := range valid {
		if v == el {
			return true
		}
	}
	return false
}

// Log requires a specific set of input strings
// event \in {"vote", "accept", "confirm", "broadcast", "connection", "send"}, is the occuring event
// msg is general text, which helps debugging
func Log(event string, msg string) {
	if !check(valid_events, event) {
		panic("INVALID EVENTS!")
	}

	// use log package to log the data
	// timestamp (usec):event:msg
	log.Printf(event + ":" + msg + "\n")
}

// Setup the logger to write to a specific file
func setupLog(name string) {
	valid_events = []string{"vote", "accept", "confirm", "broadcast", "connection", "send", "put"}

	// open the file
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Open log file error: %v\n", err)
	}

	// set up the logging package to write to stderr and the file
	mw := io.MultiWriter(os.Stderr, f)
	log.SetOutput(mw)
	log.SetFlags(log.Lmicroseconds)
}
