
all:	proto

proto:
	protoc -I client/proto -I${GOPATH}/src --go_out=plugins=grpc:client/proto --go_out=server/proto/kvstore client/proto/kvstore.proto
	protoc -I monitor/proto -I${GOPATH}/src --go_out=plugins=grpc:monitor/proto --go_out=server/proto/monitor monitor/proto/monitor.proto
	protoc -I server/proto/fvp -I${GOPATH}/src --go_out=plugins=grpc:server/proto/fvp server/proto/fvp/fvp.proto

