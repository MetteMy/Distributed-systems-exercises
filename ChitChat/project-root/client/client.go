package main

import (
	"context"
	"log"
	pb "project-root/grpc"

	"google.golang.org/grpc"
)

func main() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewChitChatServiceClient(conn)

	message := pb.Message{
		Body: "Hello from the client:)",
	}

	response, err := c.SayHello(context.Background(), &message)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", response.Body)
}
