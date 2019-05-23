package main

import (
	"io"
	"log"
	"os"
)

var (
	valid_tags   []string
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
// tag \in {"low", "medium", "high"}, defines severity
// event \in {"vote", "accept", "confirm"}, is the occuring event
// msg is general text, which helps debugging
func Log(tag string, event string, msg string) {
	if !check(valid_tags, tag) {
		panic("INVALID TAG!")
	}
	if !check(valid_events, event) {
		panic("INVALID EVENTS!")
	}

	// use log package to log the data
	// timestamp (usec):tag:event:msg
	log.Printf(tag + ":" + event + ":" + msg + "\n")
}

// Setup the logger to write to a specific file
func setupLog(name string) {
	valid_tags = []string{"low", "medium", "high"}
	valid_events = []string{"vote", "accept", "confirm"}

	// open the file
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Open log file error: %v\n", err)
	}
	// set up the logging package to write to stderr and the file
	mw := io.MultiWriter(os.Stderr, f)
	log.SetOutput(mw)
}
