package main

import (
	proto "ITUserver/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ITU_databaseServer struct {
	proto.UnimplementedITUDatabaseServer
	messages      []string
	replicaClient proto.ITUDatabaseClient
	messages []string
	auction Auction
}

type Auction struct {
	ongoing bool
	highestBid int64
	timestamp int64
	highestBidder string
	endTime int64

}

func (s *ITU_databaseServer) GetMessages(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
	return &proto.Message{Message: s.messages}, nil
}

func main() {
	ID, _ := strconv.ParseInt(os.Args[1], 10, 32)

	var mainServerClient proto.ITUDatabaseClient

	if ID == 1 {
		mainServerClient = connectReplica()
	}

	server := &ITU_databaseServer{replicaClient: mainServerClient}
	var auction = Auction{
} 

	// every time server gets a new bid, increment logical clock and update replica server
	// physical time needs to be transfered to replica at certain pysical time intervals, alongside a logical timestamp update
	server := &ITU_databaseServer{messages: []string{}}
	server.start_server(int32(ID))
}

func (s *ITU_databaseServer) placeBid(ctx context.Context, bid *proto.Bid) {
	var id = bid.Id
	var timestamp = bid.Timestamp
	var bidAmount = bid.Bid

	if s.auction.highestBid == 0 {
		s.startAuction(id, timestamp, bidAmount)
		log.Println("Auction startet")
	}
}

func (s *ITU_databaseServer) start_server(ID int32) {
	port := fmt.Sprintf(":800%d", ID)

	grpcserver := grpc.NewServer()
	listener, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("SERVER WONT WORK")
	}

	log.Println("Server Started on " + port)

	proto.RegisterITUDatabaseServer(grpcserver, s)

	err = grpcserver.Serve(listener)
}

func connectReplica() proto.ITUDatabaseClient {
	conn, err := grpc.NewClient("localhost:8002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := proto.NewITUDatabaseClient(conn)

	for {
		_, err = client.TestConnection(context.Background(), &proto.Empty{})

		if err != nil {
			log.Printf("Failed TestConnection to %s: %v", "ReplicaServer", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Successfully connected to ReplicaServer!")
		return client
	}
}

func (s *ITU_databaseServer) PlaceBid(ctx context.Context, in *proto.Bid) (*proto.Ack, error) {

	if s.replicaClient != nil {
		// example of how to send to other server
		s.replicaClient.TestConnection(ctx, &proto.Empty{})
	}

	return &proto.Ack{}, nil
}

func (s *ITU_databaseServer) PrintStatus(ctx context.Context, in *proto.Empty) (*proto.Result, error) {

	if s.replicaClient != nil {
		// example of how to send to other server
		s.replicaClient.TestConnection(ctx, &proto.Empty{})
	}

	return &proto.Result{}, nil
}

func (s *ITU_databaseServer) TestConnection(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {

	if s.replicaClient != nil {
		// example of how to send to other server
		s.replicaClient.TestConnection(ctx, &proto.Empty{})
	}

	return &proto.Empty{}, nil
}
func (s *ITU_databaseServer) startAuction(name string, timestamp int64, bidAmount int64){
	s.auction = Auction{
		ongoing: true,
		highestBid: bidAmount,
		timestamp: timestamp,
		highestBidder: name,
		endTime: timestamp + 100,
		
	}
}

//