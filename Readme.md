# Write your own Rate Limiter

## Description
The challenge was to build a rate limiter with the following rate limiting algorithms -
1. Token Bucket
2. Fixed Window Counter
3. Sliding Window Log

---
## Usage
Steps to build the binary (NOTE: this would run a test server (using port 8080, it's defined in `limiter/test_server.go`) -
```
go build -o cc-rate-limiter .
``` 

---
## Flags
The following flags for the following options are available -
1. `-b - The rate limiting algorithm to be used. The options are -`
    1. `token_bucket` :
       1. Flags for token bucket
           1. `-capacity` : capacity of the bucket
           2. `-refill_rate` : rate at which the bucket is refilled (per sec)
    2. `fixed_window_counter`
       1. Flags for fixed window counter
           1. `-max_request_count` : max number of requests allowed in a window
           2. `-window_size` : size of the window (in sec)
    3. `sliding_window_log`
       1. Flags for sliding window log
           1. `-request_per_sec` : max number of requests allowed per sec
           2. `-window_size` : size of the window (in sec)
---
### Example run:

#### 1. running a test server (:8080) with `token bucket rate limiter` with capacity 10 and refill rate 1 token per second.
```
./rate-limiter -b token_bucket \
-capacity 10 \
-refill_rate 1
```

#### 2. running a test server (:8080) with `fixed window counter rate limiter` with max request count 10 and window size 1s
```
./rate-limiter -b fixed_window_counter \
-max_request_count 10 \
-window_size 1s
```

#### 3. running a test server (:8080) with `sliding window log rate limiter` with max request count 10 and window size 1s
```
./rate-limiter -b sliding_window_log \
-request_per_sec 10 \
-window_size 1s
```


#### Humble request to provide feedback on the code and the challenge. Thanks!
#### kindly raise a bug in case of any issues.