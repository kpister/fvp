
all:	proto

proto:
	protoc -I client/proto -I${GOPATH}/src --go_out=plugins=grpc:client/proto client/proto/kvstore.proto
	protoc -I server/proto -I${GOPATH}/src --go_out=plugins=grpc:server/proto server/proto/fvp.proto

