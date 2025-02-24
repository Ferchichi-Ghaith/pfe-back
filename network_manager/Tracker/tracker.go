package tracker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader with origin check
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allows all origins
	},
}

var (
	clients = make(map[*websocket.Conn]bool) // WebSocket clients
	mutex   = sync.Mutex{}                   // Protects 'clients' map
)

// Packet structure for incoming data
type Packet struct {
	Index  string `json:"_index"`
	Type   string `json:"_type"`
	Score  *int   `json:"_score"`
	Source struct {
		Layers struct {
			FrameTime       []string `json:"frame.time"`
			IPSrc           []string `json:"ip.src"`
			IPDst           []string `json:"ip.dst"`
			WSColProto      []string `json:"_ws.col.Protocol"`
			GeoipDstSummary []string `json:"ip.geoip.dst_summary"`
			IPProto         []string `json:"ip.proto"`
			TCPSrcPort      []string `json:"tcp.srcport"`
			TCPDstPort      []string `json:"tcp.dstport"`
			HTTPHost        []string `json:"http.host"`
			HTTPUserAgent   []string `json:"http.user_agent"`
			DNSQryName      []string `json:"dns.qry.name"`
			DNSA            []string `json:"dns.a"`
			WSColInfo       []string `json:"_ws.col.Info"`
			WSExpertMessage []string `json:"_ws.expert.message"`
		} `json:"layers"`
	} `json:"_source"`
}

// Transformed packet structure
type TransformedPacket struct {
	Timestamp       string `json:"timestamp"`
	Src             string `json:"src"`
	Dst             string `json:"dst"`
	Protocol        string `json:"protocol"`
	Info            string `json:"info"`
	GeoLocation     string `json:"geo_location"`
	IPProto         string `json:"ip_protocol"`
	TCPSrcPort      string `json:"tcp_srcport"`
	TCPDstPort      string `json:"tcp_dstport"`
	HTTPHost        string `json:"http_host"`
	HTTPUserAgent   string `json:"http_user_agent"`
	DNSQryName      string `json:"dns_qry_name"`
	DNSA            string `json:"dns_a"`
	WSExpertMessage string `json:"ws_expert_message"`
}

// Handles WebSocket connections
func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
		fmt.Println("Client disconnected")
	}()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	fmt.Println("Client connected")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Unexpected WebSocket closure:", err)
			}
			break
		}

		var packet Packet
		if err := json.Unmarshal(message, &packet); err != nil {
			log.Println("JSON Parsing Error:", err)
			continue
		}

		// Transform the packet
		transformedData := TransformedPacket{
			Timestamp:       getFirstElement(packet.Source.Layers.FrameTime, "N/A"),
			Src:             getFirstElement(packet.Source.Layers.IPSrc, "N/A"),
			Dst:             getFirstElement(packet.Source.Layers.IPDst, "N/A"),
			Protocol:        getFirstElement(packet.Source.Layers.WSColProto, "N/A"),
			Info:            getFirstElement(packet.Source.Layers.WSColInfo, "N/A"),
			GeoLocation:     getFirstElement(packet.Source.Layers.GeoipDstSummary, "N/A"),
			IPProto:         getFirstElement(packet.Source.Layers.IPProto, "N/A"),
			TCPSrcPort:      getFirstElement(packet.Source.Layers.TCPSrcPort, "N/A"),
			TCPDstPort:      getFirstElement(packet.Source.Layers.TCPDstPort, "N/A"),
			HTTPHost:        getFirstElement(packet.Source.Layers.HTTPHost, "N/A"),
			HTTPUserAgent:   getFirstElement(packet.Source.Layers.HTTPUserAgent, "N/A"),
			DNSQryName:      getFirstElement(packet.Source.Layers.DNSQryName, "N/A"),
			DNSA:            getFirstElement(packet.Source.Layers.DNSA, "N/A"),
			WSExpertMessage: getFirstElement(packet.Source.Layers.WSExpertMessage, "N/A"),
		}

		// Send to all WebSocket clients
		broadcastToClients(transformedData)
	}
}

// Sends a packet to all WebSocket clients
func broadcastToClients(packet TransformedPacket) {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := json.Marshal(packet)
	if err != nil {
		log.Println("Error encoding JSON for WebSocket:", err)
		return
	}

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("Error sending WebSocket message:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// Gets first element from a slice or returns default value
func getFirstElement(field []string, defaultValue string) string {
	if len(field) > 0 {
		return field[0]
	}
	return defaultValue
}

// Starts the WebSocket server
func StartWebSocketServer(hashid string) {
	// Gracefully handle shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Create a dynamic route for WebSocket using hashid
	wsRoute := "/" + hashid
	http.HandleFunc(wsRoute, handleConnection)

	server := &http.Server{Addr: ":4000", Handler: nil}

	go func() {
		log.Printf("WebSocket Server started on :4000 at route %s\n", wsRoute)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket Server Error: %v", err)
		}
	}()

	<-stop
	fmt.Println("Shutting down WebSocket server...")
	server.Close()
}
