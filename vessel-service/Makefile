build:
	protoc -I. --go_out=plugins=grpc:$(GOPATH)/src/github.com/marceloaguero/shipper/vessel-service \
	proto/vessel/vessel.proto
	docker build -t vessel-service .

run:
	docker run -p 50052:50051 vessel-service
