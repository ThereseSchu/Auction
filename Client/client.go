package main

import (
	proto "ITUserver/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	username  string
	timestamp int64
}

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
	
		var clock = NewClock()
		client := proto.NewITUDatabaseClient(conn)

		_, err = client.GetMessages(context.Background(), &proto.Empty{})
		if err != nil {
			log.Printf("Failed to GetMessages from %s: %v", port, err)
			currentID = switchID(currentID)
			time.Sleep(5 * time.Second)
			continue
		}

		var clientstruct = new(Client)

		clientstruct.username = createUser()

		for {
			handleUserInput(clientstruct, client)
		}
	}
}

func switchID(id int) int {
	if id == 1 {
		return 2
	}
	return 1
}

func handleUserInput(clientstruct *Client, client proto.ITUDatabaseClient) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var input = strings.Split(scanner.Text(), " ")
		if input[0] == "Bid" {
			bidamount, _ := strconv.ParseInt(input[1], 6, 12)
			bid(clientstruct, bidamount, client)
		} else if input[0] == "Status" {
			status(clientstruct, client)
		} else {
			log.Printf("Unknown command: %s", os.Args[0])
		}
	}
}

func bid(clientstruct *Client, amount int64, client proto.ITUDatabaseClient) {
	// Bid
	_, err := client.PlaceBid(context.Background(), &proto.Bid{
		Id:        clientstruct.username,
		Bid:       amount,
		Timestamp: clientstruct.timestamp,
	})
	if err != nil {
		return
	}
	// Wait for acknowledgement	z

	//clock.incriment
}

func status(clientstruct *Client, client proto.ITUDatabaseClient) {
	result, err := client.PrintStatus(context.Background(), &proto.Empty{})
	log.Println(result)
	if err != nil {
		return
	}
}

func createUser() string {
	fmt.Println("Enter Username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	return username
}
