package main

import (
	proto "ITUserver/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

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
		log.Printf("Auctionen er sat igang med et beløb på %d ---logical timer = (%d)", bidAmount, s.globalTick)
		return &proto.Ack{
			Status:    proto.BidStatus_SUCCESS,
			Timestamp: s.globalTick,
		}, nil
	}

	if s.auction.endTime != 0 && s.globalTick > s.auction.endTime {
		s.auction.ongoing = false
		log.Printf("Auctionen er forbi brormand ---logical timer = (%d)", s.globalTick)
		return &proto.Ack{
			Status:    proto.BidStatus_EXCEPTION,
			Timestamp: int64(s.globalTick),
		}, nil
	}

	if s.auction.highestBid >= bidAmount {
		log.Printf("Du er for fattig %s, det højeste bud er %d og du bød %d — prøv noget højere! ---logical timer = (%d)",
			id, s.auction.highestBid, bidAmount, s.globalTick)

		return &proto.Ack{
			Status:    proto.BidStatus_FAIL,
			Timestamp: s.globalTick,
		}, nil
	}

	if s.replicaClient != nil {
		err := s.doBackup(ctx)
		if err != nil {
			log.Printf("Backup failed, rejecting bid for safety.")
			return &proto.Ack{
				Status:    proto.BidStatus_EXCEPTION,
				Timestamp: s.globalTick,
			}, nil
		}
	}

	s.updateAuction(id, timestamp, bidAmount)
	log.Printf("Ny højeste bud: %d af %s ---logical timer = (%d)", bidAmount, id, s.globalTick)
	return &proto.Ack{
		Status:    proto.BidStatus_SUCCESS,
		Timestamp: int64(s.globalTick),
	}, nil
}

func (s *ITU_databaseServer) start_server(ID int64) {
	go s.startClock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

	timeLeft := int64(0)
	if s.auction.ongoing && s.auction.endTime > s.globalTick {
		timeLeft = s.auction.endTime - s.globalTick
	}

	return &proto.Result{
		HighestBidder:    s.auction.highestBidder,
		HighestBid:       s.auction.highestBid,
		AuctionIsOngoing: s.auction.ongoing,
		TimeLeft:         timeLeft,
	}, nil
}

func (s *ITU_databaseServer) TestConnection(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {

	return &proto.Empty{}, nil
}

func (s *ITU_databaseServer) SendBackup(ctx context.Context, in *proto.Backup) (*proto.Bid, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.auction.ongoing = in.Ongoing
	s.auction.highestBid = in.HigestBid
	s.auction.timestamp = in.Timestamp
	s.auction.highestBidder = in.HighestBidder
	s.auction.endTime = in.EndTime

	return &proto.Bid{Id: "Backup updateret"}, nil
}


func (s *ITU_databaseServer) doBackup(ctx context.Context) error {
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.replicaClient.SendBackup(ctx2, &proto.Backup{
		Ongoing:       s.auction.ongoing,
		HigestBid:     s.auction.highestBid,
		Timestamp:     s.auction.timestamp,
		HighestBidder: s.auction.highestBidder,
		EndTime:       s.auction.endTime,
	})
	return err
}

func (s *ITU_databaseServer) startClock() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		s.mu.Lock()
		s.globalTick++

		if s.auction.ongoing && s.globalTick > s.auction.endTime {
			s.auction.ongoing = false
			log.Println("Auktionen er forbi")
		}

		s.mu.Unlock()
	}
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
	s.auction.highestBid = bidAmount
	s.auction.timestamp = timestamp
	s.auction.highestBidder = name
}

//
