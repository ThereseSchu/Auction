package main

import (
	proto "ITUserver/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
		_, err = client.GetMessages(context.Background(), &proto.Empty{})
		if err != nil {
			log.Printf("Failed to GetMessages from %s: %v", port, err)
			currentID = switchID(currentID)
			time.Sleep(5 * time.Second)
			continue
		}

		var username = createUser()

		for {
			handleUserInput(username)
		}
	}
}

func switchID(id int) int {
	if id == 1 {
		return 2
	}
	return 1
}

func handleUserInput(username string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var input = strings.Split(scanner.Text(), " ")
		if input[0] == "Bid" {
			bid(username, input[1])
		} else if input[0] == "Status" {
			status(username)
		} else {
			log.Printf("Unknown command: %s", os.Args[0])
		}
	}
}

func bid(username string, amount string) {
	// Bid

	// Wait for acknowledgement	z
}

func status(username string) {
	// if auction not finished send status

	// if auction finished send result
}

func createUser() string {
	fmt.Println("Enter Username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	return username
}
