package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

// Define structures for parsing Nmap XML output
type NmapRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts   []Host   `xml:"host"`
}

type Host struct {
	XMLName xml.Name `xml:"host"`
	Ports   []Port   `xml:"ports>port"`
}

type Port struct {
	PortID  string  `xml:"portid,attr"`
	State   State   `xml:"state"`
	Service Service `xml:"service"`
}

type State struct {
	State string `xml:"state,attr"`
}

type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

const (
	WebSocketURL = "ws://localhost:4000"
	ClientUUID   = "843cf35d79f927bac5c197614c8844a4c7420fb2fcdb1cda1cbba4259aac8199"
	TargetHost   = "127.0.0.1"
)

// Execute an Nmap scan and parse results
func startScan(target string) (string, error) {
	// Run Nmap
	cmd := exec.Command("nmap", target, "-sV", "-oX", "scan_output.xml")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing Nmap: %v", err)
	}

	// Read XML output
	xmlFile, err := os.Open("scan_output.xml")
	if err != nil {
		return "", fmt.Errorf("error opening XML file: %v", err)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// Parse XML data
	var nmapData NmapRun
	if err := xml.Unmarshal(byteValue, &nmapData); err != nil {
		return "", fmt.Errorf("error parsing XML: %v", err)
	}

	// Get current timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Process scan results as a formatted string
	var scanResults string
	for _, host := range nmapData.Hosts {
		for _, port := range host.Ports {
			service := port.Service.Name
			if port.Service.Product != "" {
				service += " (" + port.Service.Product + ")"
			}
			scanResults += fmt.Sprintf("Port: %s | State: %s | Service: %s\n",
				port.PortID, port.State.State, service)
		}
	}

	// Append timestamp to the result
	finalResult := fmt.Sprintf("Scan Time: %s\n%s", timestamp, scanResults)

	return finalResult, nil
}

// Handles incoming WebSocket messages
func handleMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error unmarshalling WebSocket message:", err)
			continue
		}

		// Check if a valid message is received
		if messageContent, ok := msg["message"].(string); ok {
			switch messageContent {
			case "basic":
				log.Println("Received scan request: Starting Nmap basic scan")

				// Perform the scan
				scanResults, err := startScan(TargetHost)
				if err != nil {
					log.Println("Scan failed:", err)
					continue
				}

				// Prepare the formatted message
				resultMessage := map[string]string{
					"target":  ClientUUID,
					"content": scanResults,
				}

				// Convert to JSON
				jsonData, err := json.Marshal(resultMessage)
				if err != nil {
					log.Println("Error encoding JSON:", err)
					continue
				}

				// Send scan results back to server
				err = conn.WriteMessage(websocket.TextMessage, jsonData)
				if err != nil {
					log.Println("Error sending WebSocket message:", err)
				} else {
					log.Println("Scan result sent successfully")
				}
			default:
				log.Println("Unknown command received:", messageContent)
			}
		}
	}
}

func main() {
	// WebSocket connection URL
	wsURL := fmt.Sprintf("%s/%s/doscan", WebSocketURL, ClientUUID)
	log.Println("Connecting to WebSocket:", wsURL)

	// Establish WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("WebSocket connection failed:", err)
	}
	defer conn.Close()

	log.Println("WebSocket connection established!")

	// Start message listener
	handleMessages(conn)
}
