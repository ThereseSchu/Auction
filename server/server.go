package main

import (
	proto "ITUserver/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

type ITU_databaseServer struct {
	proto.UnimplementedITUDatabaseServer
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
	// we input args when running the server, like nodes in last assignment
	// args determines what port server is started on

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

func (s *ITU_databaseServer) PlaceBid(ctx context.Context, bid *proto.Empty) (*proto.Bid, error) {

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
