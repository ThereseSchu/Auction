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
	"sync"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ITU_databaseServer struct {
	proto.UnimplementedITUDatabaseServer
	replicaClient proto.ITUDatabaseClient
	messages      []string
	auction       Auction
	mu            sync.Mutex
	globalTick    int64
}

type Auction struct {
	ongoing       bool
	highestBid    int64
	timestamp     int64
	highestBidder string
	endTime       int64
}

func main() {
	ID, _ := strconv.ParseInt(os.Args[1], 10, 32)

	var mainServerClient proto.ITUDatabaseClient

	if ID == 1 {
		mainServerClient = connectReplica()
	}

	// every time server gets a new bid, increment logical clock and update replica server
	// physical time needs to be transfered to replica at certain pysical time intervals, alongside a logical timestamp update
	server := &ITU_databaseServer{
		messages:      []string{},
		auction:       Auction{},
		replicaClient: mainServerClient}

	server.start_server(ID)
}

func (s *ITU_databaseServer) PlaceBid(ctx context.Context, bid *proto.Bid) (*proto.Ack, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    id := bid.Id
    timestamp := bid.Timestamp
    bidAmount := bid.Bid

    if timestamp > s.globalTick {
        s.globalTick = timestamp
    }
    s.globalTick++

    if s.auction.highestBid == 0 {
        s.startAuction(id, timestamp, bidAmount)
        log.Println("Auction started")
        return &proto.Ack{
            Status:    proto.BidStatus_SUCCESS,
            Timestamp: int64(s.globalTick),
        }, nil
    }

    if s.auction.endTime != 0 && s.globalTick > s.auction.endTime {
        s.auction.ongoing = false
        log.Println("Auctionen er forbi brormand ):")
        return &proto.Ack{
            Status:    proto.BidStatus_FAIL,
            Timestamp: int64(s.globalTick),
        }, nil
    }

    if s.auction.highestBid >= bidAmount {
        log.Println("Du er for fattig Silas, prøv noget højere")
        return &proto.Ack{
            Status:    proto.BidStatus_FAIL,
            Timestamp: int64(s.globalTick),
        }, nil
    } 

	s.updateAuction(id, timestamp, bidAmount)

return &proto.Ack{
    Status:    proto.BidStatus_SUCCESS,
    Timestamp: int64(s.globalTick),
}, nil
}



func (s *ITU_databaseServer) start_server(ID int64) {
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
func (s *ITU_databaseServer) startAuction(name string, timestamp int64, bidAmount int64) {
	s.auction = Auction{
		ongoing:       true,
		highestBid:    bidAmount,
		timestamp:     timestamp,
		highestBidder: name,
		endTime:       timestamp + 100,
	}
}

func (s *ITU_databaseServer) updateAuction(name string, timestamp int64, bidAmount int64) {
	s.auction = Auction{
		highestBid:    bidAmount,
		timestamp:     timestamp,
		highestBidder: name,
	}
}

//
