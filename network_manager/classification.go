package network_manager

import (
	"fmt"
	"os"
	"os/exec"
)

// DisplayPcapFile processes the captured .pcap file and saves output to JSON files
func DisplayPcapFile(filename, localIP, wlanIP string) {
	fmt.Println("Analyzing file", filename)

	// Create output files
	filePaths := map[string]string{
		"localSend":    "local-send.json",
		"localReceive": "local-receive.json",
		"wlanSend":     "wlan0-send.json",
		"wlanReceive":  "wlan0-receive.json",
	}

	files := make(map[string]*os.File)
	for key, path := range filePaths {
		file, err := os.Create(path)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", path, err)
			return
		}
		defer file.Close()
		files[key] = file
	}

	// Define commands
	commands := []struct {
		cmd  *exec.Cmd
		file *os.File
		name string
	}{
		{exec.Command("tshark", "-r", filename, "-T", "json", "-Y", "ip.src == "+localIP), files["localSend"], "local send traffic"},
		{exec.Command("tshark", "-r", filename, "-T", "json", "-Y", "ip.dst == "+localIP), files["localReceive"], "local receive traffic"},
		{exec.Command("tshark", "-r", filename, "-T", "json", "-Y", "ip.src == "+wlanIP), files["wlanSend"], "wlan0 send traffic"},
		{exec.Command("tshark", "-r", filename, "-T", "json", "-Y", "ip.dst == "+wlanIP), files["wlanReceive"], "wlan0 receive traffic"},
	}

	// Start and wait for commands
	for _, c := range commands {
		c.cmd.Stdout = c.file
		err := c.cmd.Start()
		if err != nil {
			fmt.Printf("Error starting %s command: %v\n", c.name, err)
			return
		}
	}

	for _, c := range commands {
		err := c.cmd.Wait()
		if err != nil {
			fmt.Printf("Error capturing %s: %v\n", c.name, err)
			return
		}
	}

	// Output result
	fmt.Println("Traffic data captured:")
	for _, path := range filePaths {
		fmt.Println("-", path)
	}
}
