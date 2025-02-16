package network_manager

import (
	"fmt"
	"os/exec"
	"syscall"
)

var tsharkCmd *exec.Cmd // Global variable to keep track of the process

// StartTshark starts a TShark capture session in the background
func StartTshark(interfaceName string) string {
	filename := "capture.pcap"

	// Prepare the TShark command
	tsharkCmd = exec.Command("tshark", "-i", interfaceName, "-w", filename, "-q")
	tsharkCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Run in a separate process group

	// Start the process
	err := tsharkCmd.Start()
	if err != nil {
		fmt.Printf("Error starting TShark: %v\n", err)
		return ""
	}

	fmt.Println("TShark started. Capturing packets to", filename, "...")
	return filename
}

// StopTshark stops the running TShark process
func StopTshark() {
	if tsharkCmd != nil && tsharkCmd.Process != nil {
		fmt.Println("Stopping TShark capture...")

		// Kill the TShark process
		err := syscall.Kill(-tsharkCmd.Process.Pid, syscall.SIGTERM)
		if err != nil {
			fmt.Printf("Error stopping TShark: %v\n", err)
		} else {
			fmt.Println("TShark stopped successfully.")
		}
	}
}
