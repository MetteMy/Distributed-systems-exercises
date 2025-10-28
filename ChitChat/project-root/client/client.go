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
	"google.golang.org/grpc/credentials/insecure"
)

type client struct {
	username string
	clock    int64
	conn     pb.ChitChatServiceClient
}

func main() {
	//setup logFile
	logFile, err := os.OpenFile("../chitchat.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// handle server address
	serverAddress := "localhost"
	if len(os.Args) > 2 { // if the client specifies an ip address
		serverAddress = os.Args[2]
	}

	//connect
	var conn *grpc.ClientConn
	conn, err = grpc.NewClient((serverAddress + ":9000"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewChitChatServiceClient(conn)

	client := &client{
		username: os.Args[1],
		clock:    0,
		conn:     c,
	}

	//LEAVING:
	_, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nLeaving Chit Chat...")

		client.clock++
		_, err := c.Leave(context.Background(), &pb.LeaveRequest{
			Username:    os.Args[1],
			LogicalTime: client.clock,
		})

		if err != nil {
			log.Printf("Error sending leave request: %v", err)
		}
		err = conn.Close()
		if err != nil {
			return
		}
		cancel()
		os.Exit(0)
	}()

	//JOINING:
	client.clock++
	msgStream, err := c.Join(context.Background(), &pb.JoinRequest{
		Username:    os.Args[1],
		LogicalTime: client.clock,
	})
	if err != nil {
		log.Fatalf("could not join: %v", err)
	}

	// Request messages
	go func() {
		for {
			//stream for receiving messages
			msg, err := msgStream.Recv()
			if msg != nil {
				//client's internal clock, is updated everytime it receives a message
				client.clock = max(msg.LogicalTime, client.clock) + 1

			}
			if err != nil {
				log.Printf("Stream closed: %v", err)
				cancel()
				os.Exit(0)
				return
			}
			fmt.Printf("[%s @ logical time %d]: %s", msg.Sender, msg.LogicalTime, msg.Body)
		}
	}()

	// PUBLISHING:
	client.clock++
	for {
		fmt.Print("> ")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if len(text) > 128 {
			fmt.Printf("\nYour message exceeds 128 characters: %s\nPlease write a message of max 128 characters", text)
		} else {
			_, err := c.Publish(context.Background(), &pb.PublishRequest{
				Sender:      os.Args[1],
				Body:        text,
				LogicalTime: client.clock,
			})
			if err != nil {
				log.Printf("Error publishing: %v", err)
			}
		}
	}

}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
