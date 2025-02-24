package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	// Create tshark command
	tsharkCmd := exec.Command("tshark",
		"-i", "wlan0",
		"-T", "json",
		"-l",
		"-e", "frame.time",
		"-e", "ip.src",
		"-e", "ip.dst",
		"-e", "_ws.col.Protocol",
		"-e", "ip.proto",
		"-e", "tcp.srcport",
		"-e", "tcp.dstport",
		"-e", "udp.srcport",
		"-e", "udp.dstport",
		"-e", "http.host",
		"-e", "http.user_agent",
		"-e", "dns.qry.name",
		"-e", "dns.a",
		"-e", "ip.geoip.dst_summary",
		"-e", "ip.geoip.src_summary",
		"-e", "_ws.col.Info",
		"-e", "ip.len",
		"-e", "ip.ttl",
		"-e", "frame.len",
		"-e", "ip.dsfield",
		"-e", "ip.flags",
		"-e", "ip.version",
		"-e", "tcp.flags.syn",
		"-e", "tcp.flags.ack",
		"-e", "tcp.flags.fin",
		"-e", "tcp.flags.urg",
		"-e", "frame.time_epoch",
		"-e", "ip.checksum",
		"-e", "tcp.analysis.retransmission",
	)

	// Create jq command
	jqCmd := exec.Command("jq",
		"--stream",
		"-c",
		"fromstream(1|truncate_stream(inputs))",
	)

	// Create wscat command
	wscatCmd := exec.Command("wscat",
		"-c",
		"ws://localhost:4000/843cf35d79f927bac5c197614c8844a4c7420fb2fcdb1cda1cbba4259aac8199",
	)

	// Connect pipes between commands
	jqCmd.Stdin, _ = tsharkCmd.StdoutPipe()
	wscatCmd.Stdin, _ = jqCmd.StdoutPipe()

	// Set output for debugging
	wscatCmd.Stdout = os.Stdout
	wscatCmd.Stderr = os.Stderr

	// Handle interrupts
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start commands in reverse order
	if err := wscatCmd.Start(); err != nil {
		log.Fatalf("Failed to start wscat: %v", err)
	}
	defer wscatCmd.Process.Kill()

	if err := jqCmd.Start(); err != nil {
		log.Fatalf("Failed to start jq: %v", err)
	}
	defer jqCmd.Process.Kill()

	if err := tsharkCmd.Start(); err != nil {
		log.Fatalf("Failed to start tshark: %v", err)
	}
	defer tsharkCmd.Process.Kill()

	// Wait for termination in background
	go func() {
		tsharkCmd.Wait()
		jqCmd.Wait()
		wscatCmd.Wait()
	}()

	// Wait for interrupt
	<-sigChan
	log.Println("Received interrupt, shutting down...")
}
