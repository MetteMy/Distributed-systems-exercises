package main

import (
	"context"
	"log"
	"os"
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

	/*joining := pb.JoinRequest{
		Username: os.Args[1],
	}*/

	stream, err := c.Join(context.Background(), &pb.JoinRequest{
		Username: os.Args[1],
	})
	if err != nil {
		log.Fatalf("could not join: %v", err)
	}

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("Stream closed: %v", err)
				return
			}
			log.Printf("[%s @ %d]: %s", msg.Sender, msg.LogicalTime, msg.Body)
		}
	}()

	/*message := pb.Message{
		Body: "Hello from the client:)",
	}

	response, err := c.SayHello(context.Background(), &message)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", response.Body)*/
}
