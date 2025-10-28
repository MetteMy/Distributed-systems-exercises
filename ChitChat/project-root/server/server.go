package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	pb "project-root/grpc"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChitChatServiceServer
	clients map[string]chan *pb.ChatMessage
	clock   int64
	mu      sync.Mutex
}

func main() {
	//setup logFile
	logFile, err := os.OpenFile("../chitchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//instantiate listener
	lis, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Chit Chat is up and running, the server is listening at %v", lis.Addr())
	fmt.Printf("The server is up and running, listening at %v\n", lis.Addr())

	//server instance:
	grpcServer := grpc.NewServer()
	s := &server{
		clients: make(map[string]chan *pb.ChatMessage),
	}
	pb.RegisterChitChatServiceServer(grpcServer, s)

	// run grpc server in a separate go routine
	//listen and serve
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	//graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// blocks code below, till signal is received
	<-signalChan
	log.Println("received termination signal, shutting down gracefully ...")
	fmt.Println("received termination signal, shutting down gracefully ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		err := s.RequestLeave()
		if err != nil {
			log.Fatalf("failed to request leave: %v", err)
		}
		grpcServer.GracefulStop()
		cancel()
	}()
	<-ctx.Done()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		log.Fatalf("timed out waiting for server to shutdown gracefully, forcing exit")
		grpcServer.Stop()
	}
	log.Println("server shutdown gracefully completed")
	fmt.Println("server shutdown gracefully completed")
}

func (s *server) Join(req *pb.JoinRequest, stream pb.ChitChatService_JoinServer) error {
	s.mu.Lock()
	msgChan := make(chan *pb.ChatMessage, 10)
	s.clients[req.Username] = msgChan

	s.clock = max(s.clock, req.LogicalTime) + 1 //increment the server's clock, when the server receives a join request
	eventTime := s.clock
	s.mu.Unlock()

	s.clock++
	joinMsg := &pb.ChatMessage{
		Sender:      "Server",
		Body:        fmt.Sprintf("Participant %s joined Chit Chat\n", req.Username),
		LogicalTime: eventTime,
	}

	s.broadcast(joinMsg)
	log.Printf("Participant %s joined Chit Chat at logical time: %d", req.Username, joinMsg.LogicalTime)

	for msg := range msgChan {
		if err := stream.Send(msg); err != nil {
			s.removeClient(req.Username)
			break
		}
	}
	return nil
}

func (s *server) broadcast(msg *pb.ChatMessage) {
	for username := range s.clients {
		s.clients[username] <- msg
	}
}

func (s *server) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.Empty, error) {
	s.clock = max(s.clock, req.LogicalTime) + 1 //increment the server's clock, when receiving a leave request
	eventTime := s.clock

	s.removeClient(req.Username)

	leaveMsg := &pb.ChatMessage{
		Sender:      "Server",
		Body:        fmt.Sprintf("Participant %s left the chat\n", req.Username),
		LogicalTime: eventTime,
	}
	s.broadcast(leaveMsg)
	log.Printf("User %s has left ChitChat at logical time: %d", req.Username, s.clock)

	return &pb.Empty{}, nil
}

// method for clean up when a client leaves chit chat:
func (s *server) removeClient(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove the client from map
	if _, exists := s.clients[username]; !exists {
		return
	}
	close(s.clients[username])
	delete(s.clients, username)
}

func (s *server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.Empty, error) {
	s.mu.Lock()
	s.clock = max(s.clock, req.LogicalTime) + 1
	eventTime := s.clock

	msg := &pb.ChatMessage{
		Sender:      req.Sender,
		Body:        req.Body,
		LogicalTime: eventTime,
	}
	s.mu.Unlock()

	s.broadcast(msg)
	log.Printf("[%s @ logical time %d]: %s", msg.Sender, msg.LogicalTime, msg.Body)
	return &pb.Empty{}, nil
}

func (s *server) RequestLeave() error {
	for username := range s.clients {
		s.clock++
		_, err := s.Leave(context.Background(), &pb.LeaveRequest{Username: username})
		if err != nil {
			return err
		}
		//msg := &pb.ChatMessage{
		//	Sender:      "Server",
		//	Body:        fmt.Sprintf("The server is shutting down, and user %s is being requested to leave", username),
		//	LogicalTime: s.clock,
		//}
		//s.clients[username] <- msg
		log.Printf("the server is shutting down and the user %s was requested to leave", username)
	}
	return nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
