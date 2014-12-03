package main

import (
	"fmt"
	"net/http"
	"ratelimiter"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You passed!")
}

func main() {
	ratelimiter.Initialize(":6379", 100)
	globalLimit := ratelimiter.Options{MaxCalls: 100, Period: 60}

	limitByIp := ratelimiter.Options{MaxCalls: 2, Period: 3, IpAddress: true}
	http.HandleFunc("/block_by_ip", ratelimiter.Limiter(globalLimit, ratelimiter.Limiter(limitByIp, echoHandler)))

	limitByUser := ratelimiter.Options{MaxCalls: 10, Period: 10, Username: true}
	http.HandleFunc("/block_by_user", ratelimiter.Limiter(globalLimit, ratelimiter.Limiter(limitByUser, echoHandler)))

	limitByPath := ratelimiter.Options{MaxCalls: 20, Period: 30, Path: true}
	http.HandleFunc("/block_by_path", ratelimiter.Limiter(globalLimit, ratelimiter.Limiter(limitByPath, echoHandler)))
	http.ListenAndServe(":8080", nil)
}
