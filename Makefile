
all:	proto

proto:
	protoc -I client/proto -I${GOPATH}/src --go_out=plugins=grpc:client/proto --go_out=server/proto/kvstore client/proto/kvstore.proto
	protoc -I monitor/proto -I${GOPATH}/src --go_out=plugins=grpc:monitor/proto --go_out=server/proto/monitor monitor/proto/monitor.proto
	protoc -I server/proto/fvp -I${GOPATH}/src --go_out=plugins=grpc:server/proto/fvp server/proto/fvp/fvp.proto

build:
	go build -o monitor/monitor github.com/kpister/fvp/monitor
	go build -o server/server github.com/kpister/fvp/server
	go build -o client/client github.com/kpister/fvp/client

