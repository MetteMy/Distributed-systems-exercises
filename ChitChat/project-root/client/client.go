package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	pb "project-root/grpc"
	"syscall"

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

	//LEAVING:
	_, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nLeaving Chit Chat...")

		_, err := c.Leave(context.Background(), &pb.LeaveRequest{
			Username: os.Args[1],
		})
		if err != nil {
			log.Printf("Error sending leave request: %v", err)
		}
		conn.Close()
		cancel()
		os.Exit(0)
	}()

	//JOINING:
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
	// Publishing
	for {
		fmt.Print("> ")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		_, err := c.Publish(context.Background(), &pb.PublishRequest{
			Sender: os.Args[1],
			Body:   text,
		})
		if err != nil {
			log.Printf("Error publishing: %v", err)
		}
	}

}
