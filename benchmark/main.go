package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func measurePut(n int, c *client, seq int32) int32 {
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("val%d", i)
		c.MessagePut(0, key, val, seq)
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

var (
	cfgFile  = flag.String("cfgFile", "", "config file")
	nRequest = flag.Int("nRequest", 5, "number of puts per term")
	nTerm    = flag.Int("nTerm", 100, "number of terms")
	delay    = flag.Int("delay", 2, "delay between terms in seconds")
)

func readConfig(cfgPath string) []string {
	f, _ := os.Open(cfgPath)
	scanner := bufio.NewScanner(f)
	ips := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line, "~")
		ips = append(ips, splits[0])
	}
	return ips
}

func main() {
	flag.Parse()

	ips := readConfig(*cfgFile)

	clients := make([]*client, 0)
	for i, ip := range ips {
		clients = append(clients, createClient("client"))
		clients[i].Connect(ip, 1)
	}
	seqNumber := (int32)(rand.Int31n(1000000000))
	for term := 0; term < *nTerm; term++ {
		seqNumber = measurePut(*nRequest, clients[0], seqNumber)

		time.Sleep(time.Duration(*delay) * time.Second)
		for i := 0; i < len(ips); i++ {
			clients[i].MessageIncrementTerm(0)
		}
	}
}
