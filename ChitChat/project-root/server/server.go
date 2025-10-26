package main

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "project-root/grpc"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChitChatServiceServer
	clients map[string]chan *pb.ChatMessage
	clock   int64
	mu      sync.Mutex
}

func main() {

	lis, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, &server{
		clients: make(map[string]chan *pb.ChatMessage),
	})

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) Join(req *pb.JoinRequest, stream pb.ChitChatService_JoinServer) error {
	s.mu.Lock()
	msgChan := make(chan *pb.ChatMessage, 10)
	s.clients[req.Username] = msgChan

	s.clock = max(s.clock, req.LogicalTime) + 1 //Inkrementér server's clock her, da serveren her modtager besked om at en client vil joine

	s.mu.Unlock()

	//s.clock++ //inkrementér igen her, da vi sender en besked ud til alle clients.
	joinMsg := &pb.ChatMessage{
		Sender: "Server",
		Body:   fmt.Sprintf("Participant %s joined Chit Chat at logical time %d", req.Username, s.clock),
		//print det egentlige tidspunkt, hvor client anmoder om at joine. Ikke tidspunktet, serveren sender join-beskeden
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
	//s.clock++ //inkrementér da vi sender en besked ud til alle clients
	leaveMsg := &pb.ChatMessage{
		Sender:      "Server",
		Body:        fmt.Sprintf("Participant %s left Chit Chat at logical time %d", username, s.clock),
		LogicalTime: s.clock,
	}
	s.broadcast(leaveMsg)
}

func (s *server) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.Empty, error) {
	s.clock = max(s.clock, req.LogicalTime) + 1 //Inkrementér server clock da vi modtager besked om, at en client smutter
	s.removeClient(req.Username)
	return &pb.Empty{}, nil
}

func (s *server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.Empty, error) {
	s.mu.Lock()
	s.clock = max(s.clock, req.LogicalTime) + 1

	msg := &pb.ChatMessage{
		Sender:      req.Sender,
		Body:        req.Body,
		LogicalTime: s.clock,
	}
	s.mu.Unlock()

	//s.clock++
	s.broadcast(msg)

	return &pb.Empty{}, nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
