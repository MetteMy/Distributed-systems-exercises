package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	pb "project-root/grpc"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChitChatServiceServer
	clients map[string]chan *pb.ChatMessage //maps key "username" to a channel
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
	log.Printf("server listening at %v", lis.Addr())

	//server instance:
	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, &server{
		clients: make(map[string]chan *pb.ChatMessage),
	})

	//listen and serve
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) Join(req *pb.JoinRequest, stream pb.ChitChatService_JoinServer) error {
	msgChan := make(chan *pb.ChatMessage, 10)
	s.mu.Lock()
	s.clients[req.Username] = msgChan
	s.mu.Unlock()

	s.clock++
	joinMsg := &pb.ChatMessage{
		Sender:      "Server",
		Body:        fmt.Sprintf("Participant %s joined Chit Chat at logical time %d", req.Username, s.clock),
		LogicalTime: s.clock,
	}
	s.broadcast(joinMsg)

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

	// Broadcast the leave message
	s.clock++
	leaveMsg := &pb.ChatMessage{
		Sender:      "Server",
		Body:        fmt.Sprintf("Participant %s left Chit Chat at logical time %d", username, s.clock),
		LogicalTime: s.clock,
	}
	s.broadcast(leaveMsg)
}

func (s *server) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.Empty, error) {
	s.removeClient(req.Username)
	return &pb.Empty{}, nil
}
func (s *server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.Empty, error) {
	s.mu.Lock()
	s.clock++
	msg := &pb.ChatMessage{
		Sender:      req.Sender,
		Body:        req.Body,
		LogicalTime: s.clock,
	}
	s.mu.Unlock()

	s.broadcast(msg)
	return &pb.Empty{}, nil
}
