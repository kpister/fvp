package main

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *node) errorHandler(err error, task string, nodeID string) string {
	errStatus := status.Convert(err)
	switch errStatus.Code() {
	case codes.OK:
		return "conn"
	case codes.Canceled:
		Log(task, "msg to "+nodeID+" was dropped (Canceled)")
		return "dropped"
	case codes.DeadlineExceeded:
		Log(task, "msg to "+nodeID+" was dropped (DeadlineExceeded)")
		return "dropped"
	case codes.Unavailable:
		Log(task, "conn to "+nodeID+" failed")
		return "conn_failed"
	default:
		Log(task, "conn to "+nodeID+" failed for unknown reasons")
		return "failed"
	}
}
