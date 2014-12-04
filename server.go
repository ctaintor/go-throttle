package main

import (
	"fmt"
	"net/http"
	"ratelimiter"
)

func echoHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "You passed!")
}

func globalStr(req *http.Request) string {
	return "global"
}

func main() {
	ratelimiter.Initialize(":6379", 100)
	var globalMaxCalls = 100
	var globalPeriod int64 = 60

	http.HandleFunc("/block_by_ip",
		ratelimiter.Limiter(globalMaxCalls, globalPeriod, globalStr,
			ratelimiter.Limiter(5, 10, ratelimiter.ByIpAddress, echoHandler)))

	http.HandleFunc("/block_by_user",
		ratelimiter.Limiter(globalMaxCalls, globalPeriod, globalStr,
			ratelimiter.Limiter(60, 30, ratelimiter.ByUsername, echoHandler)))

	http.HandleFunc("/block_by_path",
		ratelimiter.Limiter(globalMaxCalls, globalPeriod, globalStr,
			ratelimiter.Limiter(1, 1, ratelimiter.ByPath, echoHandler)))

	http.HandleFunc("/block_by_user_and_ip",
		ratelimiter.Limiter(globalMaxCalls, globalPeriod, globalStr,
			ratelimiter.Limiter(20, 10, func(req *http.Request) string {
				return ratelimiter.ByUsername(req) + ":" + ratelimiter.ByPath(req)
			}, echoHandler)))

	http.ListenAndServe(":8080", nil)
}
