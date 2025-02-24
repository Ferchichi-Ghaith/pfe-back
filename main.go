package main

import (
	hashid "T7-SERVER/network_manager/Hash"
	tracker "T7-SERVER/network_manager/Tracker"
	"fmt"
	"log"
)

func main() {
	// Get the hashed UUID using the hashid package
	hashedUUID := hashid.GetHashedUUID()

	// Print the hashed UUID
	fmt.Println("Hashed UUID:", hashedUUID)

	// Start the WebSocket server with the hashed UUID as the route
	tracker.StartWebSocketServer(hashedUUID)

	log.Println("WebSocket server is running...")
}
