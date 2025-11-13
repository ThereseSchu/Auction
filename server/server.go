package main

import (
	proto "ITUserver/grpc"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ITU_databaseServer struct {
	proto.UnimplementedITUDatabaseServer
	messages []string
}

func (s *ITU_databaseServer) GetMessages(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
	return &proto.Message{Message: s.messages}, nil
}

func main() {
	// we input args when running the server, like nodes in last assignment
	// args determines what port server is started on

	// every time server gets a new bid, increment logical clock and update replica server
	// physical time needs to be transfered to replica at certain pysical time intervals, alongside a logical timestamp update
	server := &ITU_databaseServer{messages: []string{}}
	server.start_server()
}

func (s *ITU_databaseServer) start_server() {
	grpcserver := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8000")

	if err != nil {
		log.Fatalf("SERVER WONT WORK")
	}

	log.Println("Server Started")

	proto.RegisterITUDatabaseServer(grpcserver, s)

	err = grpcserver.Serve(listener)
}
