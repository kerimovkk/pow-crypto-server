package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/kerimovkk/pow-server/internal/client"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:8080", "Server address (host:port)")
	timeout := flag.Duration("timeout", 30*time.Second, "Connection timeout")
	count := flag.Int("count", 1, "Number of quotes to request")
	flag.Parse()

	log.Printf("Connecting to server at %s...", *serverAddr)

	// Create client
	c := client.NewClient(*serverAddr, *timeout)

	// Request quotes
	successCount := 0
	for i := 0; i < *count; i++ {
		if *count > 1 {
			log.Printf("\n=== Request %d/%d ===", i+1, *count)
		}

		quote, err := c.GetQuote()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		successCount++
		fmt.Println("\n" + quote + "\n")
	}

	log.Printf("Successfully received %d/%d quotes", successCount, *count)
}
