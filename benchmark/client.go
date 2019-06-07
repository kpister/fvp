package main

import (
	"context"
	"fmt"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"time"

	kv "github.com/kpister/fvp/client/proto"
)

type client struct {
	Id        string
	conn      *grpc.ClientConn
	kvClients []kv.KeyValueStoreClient
	servAddr  string
}

func createClient(clientId string) *client {
	return &client{
		Id:        clientId,
		kvClients: make([]kv.KeyValueStoreClient, 0),
	}
}

func (c *client) MessagePut(connId int, key string, value string, seqNumber int32) kv.ReturnCode {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req := &kv.PutRequest{
		Key:      key,
		Value:    value,
		ClientId: c.Id,
		SeqNo:    seqNumber,
	}

	res, err := c.kvClients[connId].Put(
		ctx,
		req,
		grpc_retry.WithMax(5),
		grpc_retry.WithPerRetryTimeout(500*time.Millisecond))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return kv.ReturnCode_FAILURE
	}

	return res.Ret
}

func (c *client) MessageGet(connId int, key string) (string, kv.ReturnCode) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req := &kv.GetRequest{
		Key: key,
	}

	res, err := c.kvClients[connId].Get(
		ctx,
		req,
		grpc_retry.WithMax(5),
		grpc_retry.WithPerRetryTimeout(500*time.Millisecond))
	if err != nil {
		return "", kv.ReturnCode_FAILURE
	}

	return res.Value, res.Ret
}

func (c *client) MessageIncrementTerm(connId int) {
	ctx := context.Background()
	c.kvClients[connId].IncrementTerm(ctx, &kv.EmptyMessage{})
}

func (c *client) SetupConnection(servAddr string) bool {
	if c.conn != nil {
		c.conn.Close()
	}

	c.servAddr = servAddr
	// grpc will retry in 15 ms at most 5 times when failed
	// TODO: put parameters into config
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithPerRetryTimeout(500 * time.Millisecond),
		grpc_retry.WithCodes(codes.DeadlineExceeded, codes.Unavailable, codes.Canceled),
	}
	conn, err := grpc.Dial(c.servAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))
	if err != nil {
		return false
	}

	c.conn = conn
	return true
}

func (c *client) Connect(servAddr string, N int) bool {
	if !c.SetupConnection(servAddr) {
		return false
	}

	c.kvClients = make([]kv.KeyValueStoreClient, 0)
	for i := 0; i < N; i++ {
		c.kvClients = append(c.kvClients, kv.NewKeyValueStoreClient(c.conn))
	}

	return true
}
