package main

import (
	"context"
	"log"
	"net"

	// Import the generated protobuf code
	pb "github.com/marceloaguero/shipper/consignment-service/proto/consignment"
	vesselProto "github.com/marceloaguero/shipper/vessel-service/proto/vessel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port          = ":50051"
	vesselAddress = "127.0.0.1:50052"
)

type Repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

// Repository - Dummy repository, this simulates the use of a datastore
// of some kind. We'll replace this with a real implementation later on.
type ConsignmentRepository struct {
	consignments []*pb.Consignment
}

func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

func (repo *ConsignmentRepository) GetAll() []*pb.Consignment {
	return repo.consignments
}

// Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition. You can check the interface
// in the generated code itself for the exact method signatures etc
// to give you a better idea.
type service struct {
	repo         Repository
	vesselClient vesselProto.VesselServiceClient
}

// CreateConsignment - we created just one method on our service,
// which is a create method, which takes a context and a request as an
// argument, these are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment) (*pb.Response, error) {

	// Here we call a client instance of our vessel service with our consignment weight,
	// and the amount of containers as the capacity value
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})

	log.Printf("Found vessel: %s, \n", vesselResponse.Vessel.Name)
	if err != nil {
		return nil, err
	}

	// We set the VesselId as the vessel we got back
	// from our vessel service
	req.VesselId = vesselResponse.Vessel.Id

	// Save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return nil, err
	}

	// Return matching the `Response` message we created in our
	// protobuf definition.
	return &pb.Response{Created: true, Consignment: consignment}, nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest) (*pb.Response, error) {
	consignments := s.repo.GetAll()
	return &pb.Response{Consignments: consignments}, nil
}

func main() {

	repo := &ConsignmentRepository{}

	// Connect to vessel service
	conn, err := grpc.Dial(vesselAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect to vessel service: %v", err)
	}
	defer conn.Close()
	vesselClient := vesselProto.NewVesselServiceClient(conn)

	// Set-up our gRPC server.
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Register our service with the gRPC server, this will tie our
	// implementation into the auto-generated interface code for our
	// protobuf definition.
	pb.RegisterShippingServiceServer(s, &service{repo, vesselClient})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
