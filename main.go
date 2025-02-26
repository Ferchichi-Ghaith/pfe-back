package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	hashid "T7-SERVER/network_manager/Hash"
	Tshark "T7-SERVER/network_manager/Tshark" // Fixed import alias
)

func main() {
	// Get the hashed UUID
	hashuuid := hashid.GetHashedUUID()

	// Prepare the JSON payload
	payload := map[string]string{"uuid": hashuuid}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Define the Node.js server URL
	url := "https://ws-t7-production.up.railway.app/uuid"

	// Send the POST request
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Print response
	fmt.Println("route:", hashuuid)
	if resp.StatusCode == 200 {
		fmt.Println("Success route created")
	} else {
		fmt.Println("Failed")
	}

	// Start Tshark
	Tshark.StartTshark(hashuuid) // Call the exported function ✅
}
