package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func measurePut(n int, c *client, seq int32, logFp *os.File) int32 {
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("val%d", i)
		t := time.Now()
		c.MessagePut(0, key, val, seq)
		timeTrack(t, "PUT", logFp)
		seq++
	}
	return seq
}

func measureGet(n int, c *client, logFp *os.File) {
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key%d", i)
		t := time.Now()
		c.MessageGet(0, key)
		timeTrack(t, "GET", logFp)
	}
}

func timeTrack(start time.Time, task string, logFp *os.File) {
	elapsed := time.Since(start)
	if logFp != nil {
		logFp.WriteString(fmt.Sprintf("%s duration:%d\n", task, elapsed))
	}
}

func main() {
	path := path.Join(os.Getenv("LOCAL"), "client_benchmark.txt")
	logFp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Open log file error: %v\n", err)
	}

	inputReader := bufio.NewReader(os.Stdin)
	client := createClient("client")
	seqNumber := (int32)(rand.Int31n(1000000000))
	for {
		in, _ := inputReader.ReadString('\n')
		in = strings.TrimSpace(in)

		splits := strings.Split(in, " ")
		if len(splits) <= 1 {
			fmt.Println("bad input")
			continue
		}

		switch splits[0] {
		// CONN host:port
		case "CONN":
			if len(splits) != 2 {
				fmt.Println("bad input")
				continue
			}

			if client.Connect(splits[1], 1) {
				fmt.Println("connected to " + splits[1])
			} else {
				fmt.Println("failed to connect to " + splits[1])
			}

		// PUT key val
		case "PUT":
			if len(splits) != 3 {
				fmt.Println("bad input")
				continue
			}

			key, val := splits[1], splits[2]
			ret := client.MessagePut(0, key, val, seqNumber)
			fmt.Printf("return code: %v\n", ret)
			seqNumber++

		// GET key
		case "GET":
			if len(splits) != 2 {
				fmt.Println("bad input")
				continue
			}

			key := splits[1]
			val, ret := client.MessageGet(0, key)
			fmt.Printf("return value: %s\n", val)
			fmt.Printf("return code: %v\n", ret)

		// set client ID
		// ID clientID
		case "ID":
			if len(splits) != 2 {
				fmt.Println("bad input")
				continue
			}

			client.Id = splits[1]

		case "PUT_N":
			n, _ := strconv.Atoi(splits[1])
			seqNumber = measurePut(n, client, seqNumber, logFp)

		case "GET_N":
			n, _ := strconv.Atoi(splits[1])
			measureGet(n, client, logFp)
		}
	}
}
