# Go Rate Limiter

This is a simple rate limiter which uses Redis to keep track of call frequency

## Usage Overview

Every rate limit request has three components:

* A time period, in seconds
* A maximum number of allowed calls
* A configurable uniqueness constraint

For example, you can say that you want to limit an endpoint to 10 reqs per second per ip address by saying

```
ratelimiter.Limiter(10, 1, ratelimiter.ByIpAddress, myNormalHandlerFunction)
```

For simplicity, I just implemented this as a wrapper over the `http.HandlerFunc` function type in the `net/http`
Go package. You can stack Limiters on top of each other, if desired, and you can write your own `ratelimiter.KeyPartialFunc`
functions which, given a request as input, output a string of your choosing (e.g. IP address \& username)

## Implementation Overview

Essentially, I construct a Redis key based off of the uniqueness constraint string and the 'time bucket' that
the request fits into. When a request comes in, an appropriate time bucket is computed, meaning that if you had a 5 minute
bucket that started at 11:00, all requests from 11:00-11:05 would resolve to the same key in Redis. A counter is incremented
for that key and, when the count exceeds the maximum, all further requests cause an HTTP 429 response. This continues until
the request is mapped to a new bucket.

## Quirks

* Since the limiters can be stacked, limiters that are deeper inside won't be incremented once an outer limiter is
  throttling the user.
* Since the time windows are fixed, a user could time a flood of requests to happen near the boundaries of the time window 
  such that the number of requests that happen in a short time period could be higher than you expect.

## Running the code

I have an example server which has a few rate-limited endpoints. It expects that you have Redis running locally.

```
cd /the/code
source .envrc
go run server.go
```

then

```
# To see rate limiting by user & IP, with global rate limit
curl -u case.taintor: localhost:8080/block_by_user_and_ip
# To see rate limiting by user
curl -u case.taintor: localhost:8080/block_by_user
# To see rate limiting by path
curl -u case.taintor: localhost:8080/block_by_path
```

## Scaling and High Availability

It's hard to say what your scaling strategy should be when it comes to rate limiting without knowing what sort of
problems you are having. As in, it's highly likely that if you are experiencing serious scaling issues related
to rate limiting, then you are probably experiencing other issues such as network saturation. A few concepts:

* Utilize Redis HA capabilities - such as RedisLabs or Sentinel (I should say that I have not investigated whether these truly 
  work...)
* It could be a good idea to separate your Redis usage to mitigate issues where rate-limiting infrastructure affects normal-
  usage infrastructure
* Ensure that the rate limiting happens as far up the request chain as possible - potentially in a service separate from your
  business logic.
* Apply limits that are easy to generically apply and design your APIs in a way to enable this. For instance, I have seen a few
  APIs that require payload parsing in order to determine who the user is, which means that user-level rate limiting is 
  ineffective.