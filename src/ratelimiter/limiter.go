package ratelimiter

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
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

type Options struct {
	MaxCalls  int
	Period    int64
	Username  bool
	IpAddress bool
	Path      bool
}

func Limiter(options Options, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c := redisPool.Get()
		defer c.Close()

		fmt.Println("OK")
		keyPartial := generateKeyPartial(options, req)
		timeBucket := strconv.FormatInt(time.Now().Unix()/options.Period, 10)
		key := "ratelimiter:" + keyPartial + ":" + timeBucket

		expireIn := options.Period - (time.Now().Unix() % options.Period) + 1
		fmt.Printf("%s - %d\n", key, expireIn)

		currentCount, _ := redis.Int(c.Do("INCR", key))
		c.Do("EXPIRE", key, expireIn)

		fmt.Fprintf(w, "I wrap %d-%d/%d!", options.Period, currentCount, options.MaxCalls)
		if currentCount > options.MaxCalls {
			fmt.Fprintf(w, "FAIL!")
		} else {
			handlerFunc(w, req)
		}
	}
}

func generateKeyPartial(options Options, req *http.Request) string {
	limiters := make([]string, 0, 3)
	if options.IpAddress {
		ip, _, _ := net.SplitHostPort(req.RemoteAddr)
		limiters = append(limiters, "ip-"+ip)
	}
	if options.Username {
		//I assume that they are sending the username using HTTP Basic, with no password (e.g. "case.taintor:")
		authorization, ok := req.Header["Authorization"]
		if ok && len(authorization) > 0 {
			parts := strings.SplitN(authorization[0], " ", 2)
			if len(parts) == 2 {
				//This is actually the base64 username - less readable, but no real rason to decode for this use-case :)
				limiters = append(limiters, "username-"+parts[1])
			}
		}
	}
	if options.Path {
		limiters = append(limiters, "path-"+req.URL.Path)
	}

	keyPartial := strings.Join(limiters, ":")
	if len(keyPartial) == 0 {
		keyPartial = "global"
	}
	return keyPartial
}
