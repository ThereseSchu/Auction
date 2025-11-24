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
	username string
	clock    *Clock
}

func main() {
	// pseudo-code
	// we try connecting to this below, if error from Getmessage != nill,
	// we try to connect to 8001 instead
	// is try-catch a thing in go?
	// send message should be bid
	// get message should be result
	// we keep getmessage for error checking

	scanner := bufio.NewScanner(os.Stdin)
	currentID := 1

	var clientstruct = new(Client)
	clientstruct.clock = NewClock()
	clientstruct.username = createUser(scanner)
	fmt.Println("Welcome to the Auction\n   Post a bid with 'Bid 'Amount''\n   Get the status with 'Status'")

	for {
		// port := fmt.Sprintf("port:800%d", currentID)
		port := fmt.Sprintf("localhost:800%d", currentID)

		log.Printf("Connected to %s...", port)

		conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Printf("Failed to create client for %s: %v", port, err)
			currentID = switchID(currentID)
			time.Sleep(5 * time.Second)

			continue
		}

		client := proto.NewITUDatabaseClient(conn)

		_, err = client.TestConnection(context.Background(), &proto.Empty{})

		if err != nil {
			log.Printf("Failed to GetMessages from %s: %v", port, err)
			currentID = switchID(currentID)
			time.Sleep(5 * time.Second)
			continue
		}

		if !handleUserInput(clientstruct, client, scanner) {
			break
		}
	}
}

func switchID(id int) int {
	if id == 1 {
		return 2
	}
	return 1
}

func handleUserInput(clientstruct *Client, client proto.ITUDatabaseClient, scanner *bufio.Scanner) bool {
	for scanner.Scan() {
		var input = strings.Split(scanner.Text(), " ")
		if input[0] == "Bid" {
			bidamount, _ := strconv.ParseInt(input[1], 10, 64)
			bid(clientstruct, bidamount, client)
		} else if input[0] == "Status" {
			status(clientstruct, client)
		} else {
			log.Printf("Unknown command: %s", os.Args[0])
		}
		return true
	}
	return false
}

func bid(clientstruct *Client, amount int64, client proto.ITUDatabaseClient) bool {
	// Bid
	_, err := client.PlaceBid(context.Background(), &proto.Bid{
		Id:        clientstruct.username,
		Bid:       amount,
		Timestamp: int64(clientstruct.clock.GetTime()),
	})
	if err != nil {

		fmt.Println("------------------------------------------------")
		fmt.Println("    Connection to server lost. Switched to backup.")
		fmt.Println("    Your previous bid was NOT processed.")
		fmt.Println("    Please enter your bid again.")
		fmt.Println("------------------------------------------------")

		return false
	}
	clientstruct.clock.Increment()
	return true
}

func status(clientstruct *Client, client proto.ITUDatabaseClient) bool {
	result, err := client.PrintStatus(context.Background(), &proto.Empty{})
	if err != nil {
		log.Println("Failed to get auction status:", err)
		return false
	}

	if !result.AuctionIsOngoing {
		fmt.Println("Der er ingen igangværende auktion lige nu.")
	} else {
		fmt.Printf("Højeste bud: %d fra bidder: %s, tid tilbage: %d\n",
			result.HighestBid, result.HighestBidder, result.TimeLeft)
	}

	clientstruct.clock.Increment()
	return true
}

func createUser(scanner *bufio.Scanner) string {
	fmt.Println("Enter Username: ")
	scanner.Scan()
	username := scanner.Text()

	return username
}
