build:
	go build -race -o diskey ./cmd/main.go

run: build
	./diskey

protobuf:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	pkg/command/builtin.proto

bench:
	#make --directory ./pkg/batcher bench
	make --directory ./pkg/cluster bench