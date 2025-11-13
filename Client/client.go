package main

import (
	proto "ITUserver/grpc"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	// pseudo-code
	// we try connecting to this below, if error from Getmessage != nill,
	// we try to connect to 8001 instead
	// is try-catch a thing in go?
	// send message should be bid
	// get message should be result
	// we keep getmessage for error checking

	currentID := 1

	for {
		// port := fmt.Sprintf("port:800%d", currentID)
		port := fmt.Sprintf("localhost:800%d", currentID)
		log.Printf("Attempting to connect to %s...", port)

		conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Printf("Failed to create client for %s: %v", port, err)
			currentID = switchID(currentID)
			time.Sleep(5 * time.Second)
			continue
		}

		client := proto.NewITUDatabaseClient(conn)
		messages, err := client.GetMessages(context.Background(), &proto.Empty{})

		if err != nil {
			log.Printf("Failed to GetMessages from %s: %v", port, err)
			currentID = switchID(currentID)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Successfully got messages from %s", port)
		_ = messages

		time.Sleep(10 * time.Second)
	}
}

func switchID(id int) int {
	if id == 1 {
		return 2
	}
	return 1
}
