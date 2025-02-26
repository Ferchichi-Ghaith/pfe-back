package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// This is the handler that will respond to requests
	fmt.Fprintf(w, "Hello, you've reached %s!", r.URL.Path)
}

func main() {
	// Register the handler for the default route "/"
	http.HandleFunc("/", handler)

	// Start the server and listen on port 8080
	fmt.Println("Server is listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting the server: ", err)
	}
}
