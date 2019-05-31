package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

func main() {
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
		}
	}
}
