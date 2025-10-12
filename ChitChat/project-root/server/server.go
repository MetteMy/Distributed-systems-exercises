package main

import (
	"fmt"
	"log"
	"net"
	pb "project-root/grpc"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChitChatServiceServer
	clients map[string]chan *pb.ChatMessage //maps key username to a channel
	clock   int64
	mu      sync.Mutex
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
			delete(s.clients, req.Username)
			s.Leave()
			break // client disconnected
		}
	}
	return nil
}

func (s *server) broadcast(msg *pb.ChatMessage) {
	for username := range s.clients {
		s.clients[username] <- msg
	}
}

func (s *server) Leave(req *pb.LeaveRequest, stream pb.ChitChatService_LeaveServer) error {
	s.clock++
	leaveMsg := &pb.ChatMessage{
		Sender:      "Server",
		Body:        fmt.Sprintf("Participant %s left Chit Chat at logical time %d", req.Username, s.clock),
		LogicalTime: s.clock,
	}
	s.broadcast(leaveMsg)
	return nil
}

/*func (s *server) SayHello(ctx context.Context, message *pb.Message) (*pb.Message, error) {
	log.Printf("Received: %v", message.Body)
	return &pb.Message{Body: "Hello from the server"}, nil
}*/
