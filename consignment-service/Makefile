build:
	protoc -I. --go_out=plugins=grpc:$(GOPATH)/src/github.com/marceloaguero/shipper/consignment-service \
		proto/consignment/consignment.proto
	docker build -t consignment-service .

run:
	docker run -p 50051:50051 consignment-service

