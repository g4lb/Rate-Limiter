package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type reportRequest struct {
	URL string `json:"url"`
}

type reportResponse struct {
	Reported bool `json:"blocked"`
}

type limiter struct {
	mu         sync.Mutex
	lastAccess map[string]time.Time
	counts     map[string]int
}

func (l *limiter) shouldLimit(url string, threshold int, ttl time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Remove old entries from lastAccess map
	for k, v := range l.lastAccess {
		if time.Since(v) > ttl {
			log.Printf("Expiring URL: %s\n", k)
			delete(l.lastAccess, k)
			delete(l.counts, k)
		}
	}

	// Check if we have exceeded the threshold for this URL
	count, ok := l.counts[url]
	if ok && count >= threshold {
		log.Printf("Rate limit exceeded for URL: %s\n", url)
		return true
	}

	// Update counts for this URL and return false
	l.lastAccess[url] = time.Now()
	l.counts[url]++
	log.Printf("Count incremented for URL: %s\n", url)
	return false
}

func main() {
	// Parse command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./rate_limiter <threshold> <ttl>")
		return
	}

	threshold, err := strconv.Atoi(os.Args[1])

	if err != nil {
		fmt.Println("Invalid threshold parameter")
		return
	}

	ttl, err := time.ParseDuration(os.Args[2])
	if err != nil {
		fmt.Println("Invalid TTL parameter")
		return
	}

	ratePtr := flag.Int("rate", threshold, "Rate limit in requests per second")
	ttlPtr := flag.Duration("ttl", ttl, "Time period in which URL visits will be counted")
	flag.Parse()

	// Create a new limiter with the specified rate limit and TTL
	lim := &limiter{
		lastAccess: make(map[string]time.Time),
		counts:     make(map[string]int),
	}

	// Schedule a function to clear the counts for each URL after the TTL has over
	ticker := time.NewTicker(*ttlPtr)
	go func() {
		for {
			<-ticker.C
			log.Println("Clearing counts for all URLs")
			lim.mu.Lock()
			for k := range lim.counts {
				delete(lim.counts, k)
			}
			lim.mu.Unlock()
		}
	}()

	// Define the HTTP handlers
	http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var req reportRequest
		log.Println("Received report request")
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if we should limit this request
		if lim.shouldLimit(req.URL, *ratePtr, *ttlPtr) {
			// Do nothing
		}

		// Return true if the URL has been reported enough times to reach the threshold
		if lim.counts[req.URL] >= *ratePtr {
			log.Printf("URL has been reported %d times, reporting as true: %s\n", lim.counts[req.URL], req.URL)
			json.NewEncoder(w).Encode(reportResponse{Reported: true})
			return
		}

		// Return false if the URL has not been reported enough times to reach the threshold
		log.Printf("URL has been reported %d times, reporting as false: %s\n", lim.counts[req.URL], req.URL)
		json.NewEncoder(w).Encode(reportResponse{Reported: false})
	})

	// Start the HTTP server
	addr := fmt.Sprintf(":%d", 8080)
	log.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
