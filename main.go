package main

import (
	"T7-SERVER/network_manager"
	"bufio"
	"fmt"
	"net"
	"os"
)

// Function to get the dynamic IP of wlan0
func getWlanIP(interfaceName string) (string, error) {
	intf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", fmt.Errorf("error getting interface %s: %v", interfaceName, err)
	}

	addrs, err := intf.Addrs()
	if err != nil {
		return "", fmt.Errorf("error getting addresses for %s: %v", interfaceName, err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil { // Only return IPv4 address
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IPv4 address found for %s", interfaceName)
}

func main() {
	interfaceName := "wlan0" // Change this based on your system

	// Get dynamic WLAN IP
	wlanIP, err := getWlanIP(interfaceName)
	if err != nil {
		fmt.Println("Error getting WLAN IP:", err)
		return
	}

	fmt.Println("WLAN IP detected:", wlanIP)

	// Start packet capture
	fmt.Println("Starting packet capture on", interfaceName, "...")
	file := network_manager.StartTshark(interfaceName)

	if file == "" {
		fmt.Println("Failed to start TShark")
		return
	}

	fmt.Println("Packet capture saved to:", file)

	// Wait for user input to stop the capture
	fmt.Println("Press ENTER to stop capturing...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// Stop the capture
	network_manager.StopTshark()
	fmt.Println("Packet capturing stopped.")

	// Start classification using the local IP and dynamic wlan0 IP
	localIP := "127.0.0.1"
	fmt.Println("Classifying captured packets...")
	network_manager.DisplayPcapFile(file, localIP, wlanIP)
	fmt.Println("Packet classification completed.")
}
