package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// Define structures to match the XML output
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

type PrettyPort struct {
	Port    string `json:"port"`
	State   string `json:"state"`
	Service string `json:"service"`
}

func main() {
	cmd := exec.Command("nmap", "127.0.0.1", "-sV", "-oX", "data.xml")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing nmap:", err)
		return
	}

	// Read the generated XML file
	xmlFile, err := os.Open("data.xml")
	if err != nil {
		fmt.Println("Error opening XML file:", err)
		return
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// Parse the XML
	var nmapData NmapRun
	err = xml.Unmarshal(byteValue, &nmapData)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	// Extract and format Ports array
	var formattedPorts []PrettyPort
	for _, host := range nmapData.Hosts {
		for _, port := range host.Ports {
			service := port.Service.Name
			if port.Service.Product != "" {
				service += " (" + port.Service.Product + ")"
			}
			formattedPorts = append(formattedPorts, PrettyPort{
				Port:    port.PortID,
				State:   port.State.State,
				Service: service,
			})
		}
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(formattedPorts, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return
	}

	// Print the JSON output
	fmt.Println(string(jsonData))
}
