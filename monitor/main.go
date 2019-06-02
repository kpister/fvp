package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path"
	"strings"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"github.com/kpister/fvp/server/proto/fvp"
	mon "github.com/kpister/fvp/server/proto/monitor"
	"google.golang.org/grpc"
)

type node struct {
	Address          string
	ServerAddrs      []string
	FailedNodesAddrs []string
	NodesFvpClients  map[string]fvp.ServerClient
}

func (n *node) connectServers() {
	// grpc will retry in 20 ms at most 5 times when failed
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithPerRetryTimeout(500 * time.Millisecond),
	}

	for _, addr := range n.ServerAddrs {

		Log("low", "connection", "Connecting to "+addr)

		conn, err := grpc.Dial(addr, grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))
		if err != nil {
			Log("low", "connection", fmt.Sprintf("Failed to connect to %s. %v\n", addr, err))
		}

		n.NodesFvpClients[addr] = fvp.NewServerClient(conn)
	}
}

func (n *node) MonitorSend(ctx context.Context, in *mon.MonitorSendMsg) (*mon.EmptyMessage, error) {

	// for good nodes just relay the message
	if !inArray(n.FailedNodesAddrs, in.From) {
		// send to each server in to list
		for _, addr := range in.To {

			ctx, cancel := context.WithTimeout(
				context.Background(),
				time.Duration(100)*time.Millisecond)
			defer cancel()

			_, err := n.NodesFvpClients[addr].Send(ctx, in.Msg)
			if err != nil {
				n.errorHandler(err, "send", addr)
			}
		}
	} else { // do some wierd shit for bad nodes

	}

	return &mon.EmptyMessage{}, nil
}

func (n *node) GetState(ctx context.Context, in *mon.EmptyMessage) (*mon.ServerState, error) {

	return &mon.ServerState{}, nil
}

func (n *node) createNode() {
	n.NodesFvpClients = make(map[string]fvp.ServerClient)
}

var (
	n          *node
	configFile = flag.String("config", "mon_config.json", "the file to read the monitor configuration from")
	help       = flag.Bool("h", false, "for usage")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	readConfig(*configFile)
}

func readConfig(configFile string) {
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(configData, &n)
	n.createNode()

	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	setupLog(path.Join(os.Getenv("LOCAL"), "monitor", "log.txt"))

	grpcServer := grpc.NewServer()
	mon.RegisterMonitorServer(grpcServer, n)
	n.connectServers()

	lis, err := net.Listen("tcp", ":"+strings.Split(n.Address, ":")[1])
	if err != nil {
		Log("low", "connection", fmt.Sprintf("Failed to listen on the port. %v", err))
	}

	Log("low", "connection", "Listening on "+n.Address)
	grpcServer.Serve(lis)

}
