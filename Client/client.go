package main

import (
	proto "ITUserver/grpc"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// pseudo-code
	// we try connecting to this below, if error from Getmessage != nill,
	// we try to connect to 8001 instead
	// is try-catch a thing in go?
	// send message should be bid
	// get message should be result
	// we keep getmessage for error checking

	conn, err := grpc.NewClient("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Count not connect")
	}

	client := proto.NewITUDatabaseClient(conn)

	if err != nil {
		log.Fatalf("WE DID NOT RECIEVE OR FAILED TO SEND")
	}

	messages, err := client.GetMessages(context.Background(), &proto.Empty{})
	_ = messages
}
