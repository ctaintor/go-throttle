package ratelimiter

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var redisPool *redis.Pool

func Initialize(redisAddress string, maxConnections int) {
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, maxConnections)
}

type KeyPartialFunc func(req *http.Request) string

func Limiter(maxCalls int, period int64, keyPartialFunc KeyPartialFunc, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c := redisPool.Get()
		defer c.Close()

		keyPartial := keyPartialFunc(req)
		timeBucket := strconv.FormatInt(time.Now().Unix()/period, 10)
		key := "ratelimiter:" + keyPartial + ":" + timeBucket

		expireIn := period - (time.Now().Unix() % period) + 1

		c.Send("MULTI")
		c.Send("INCR", key)
		c.Send("EXPIRE", key, expireIn)
		responses, ok := c.Do("EXEC")
		values, ok := redis.Values(responses, ok)
		currentCount, ok := redis.Int(values[0], ok)

		log.Printf("%d - every %ds (%d/%d)", key, period, currentCount, maxCalls)
		if currentCount > maxCalls {
			w.WriteHeader(429)
		} else {
			handlerFunc(w, req)
		}
	}
}

func ByIpAddress(req *http.Request) string {
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

func ByUsername(req *http.Request) string {
	//I assume that they are sending the username using HTTP Basic, with no password (e.g. "case.taintor:")
	authorization, ok := req.Header["Authorization"]
	if ok && len(authorization) > 0 {
		parts := strings.SplitN(authorization[0], " ", 2)
		if len(parts) == 2 {
			//This is actually the base64 username - less readable, but no real rason to decode for this use-case :)
			return "username-" + parts[1]
		}
	}
	return ""
}

func ByPath(req *http.Request) string {
	return "path-" + req.URL.Path
}
