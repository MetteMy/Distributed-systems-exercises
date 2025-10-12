package main

import (
	"context"
	"log"
	"net"
	pb "project-root/grpc"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChitChatServiceServer
}

func main() {
	//instantiate listener
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())

	//server instance:
	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, &server{})

	//listen and serve
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) SayHello(ctx context.Context, message *pb.Message) (*pb.Message, error) {
	log.Printf("Received: %v", message.Body)
	return &pb.Message{Body: "Hello from the server"}, nil
}
