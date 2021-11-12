# Benchmarks
|   Server | Requests/sec |
|---------:|--------------|
|      Chi | 226k         |
|      Std | 612k         |
| Fasthttp | 1289k        |


## Update 13.11.2021

```console
GOMAXPROCS=20 techempower
```

```console
$ wrk -d 10s -t 6 -c 1000 http://localhost:8080/json
Running 10s test @ http://localhost:8080/json
  6 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.03ms    1.31ms  45.87ms   91.33%
    Req/Sec   132.58k     8.35k  164.28k    75.83%
  7915077 requests in 10.05s, 0.99GB read
Requests/sec: 787235.61
Transfer/sec:    100.60MB
```

```console
GOMAXPROCS=12 techempower
```
```console
Running 10s test @ http://localhost:8080/json
  6 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.37ms    1.21ms  42.10ms   87.44%
    Req/Sec   122.63k     7.32k  159.16k    71.33%
  7319976 requests in 10.05s, 0.91GB read
Requests/sec: 728012.89
Transfer/sec:     93.03MB
```
```console
$ wrk -d 10s -t 6 -c 50 http://localhost:8080/json
Running 10s test @ http://localhost:8080/json
  6 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   116.32us  141.81us   1.55ms   86.31%
    Req/Sec    92.37k     5.54k  115.54k    71.29%
  5569038 requests in 10.10s, 711.68MB read
Requests/sec: 551415.23
Transfer/sec:     70.47MB
```


```console
GOMAXPROCS=6 techempower
```
```console
$ wrk -d 10s -t 6 -c 50 http://localhost:8080/json
Running 10s test @ http://localhost:8080/json
  6 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   129.13us  113.42us   1.62ms   87.01%
    Req/Sec    70.09k     2.84k   77.37k    71.78%
  4226043 requests in 10.10s, 540.06MB read
Requests/sec: 418441.89
Transfer/sec:     53.47MB
```
```console
$ wrk -d 10s -t 6 -c 1000 http://localhost:8080/json
Running 10s test @ http://localhost:8080/json
  6 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.40ms    1.24ms  37.46ms   82.28%
    Req/Sec    70.37k     4.26k   78.03k    75.50%
  4199832 requests in 10.05s, 536.71MB read
Requests/sec: 417884.38
Transfer/sec:     53.40MB
```

### Results

65-72K RPS per core, concurrency increases tail latency.
