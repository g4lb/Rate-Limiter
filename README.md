# Rate-Limiter
A web service that acts as a third party rate limiter, implemented in Go language

# Project Description
The server implements a limiter that maintains a count of requests to each URL within a given time period. If the count exceeds a certain threshold, the server responds with block true/false - depends on if the number of times the URL was reported reached the threshold

The rate limit and time period are configurable via command-line arguments.

The server exposes a single HTTP endpoint (/report) that accepts a JSON payload containing a URL. The server checks if the URL has been reported more than the configured rate limit within the configured time period. If it has, the server responds with a JSON payload indicating that the URL has been reported. If it has not, the server responds with a JSON payload indicating that the URL has not been reported. to replace http://localhost:8080 with the actual server address and port if you are running the server on a different machine or port.

# Usage

* git clone https://github.com/g4lb/Rate-Limiter.git
    * cd rate-limiter
  * go build
  * go run rateLimiter.go 5 10s