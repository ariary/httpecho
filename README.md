# httpecho

The same thing as [`jmalloc/echo-server`](https://github.com/jmalloc/echo-server) **but malformed HTTP request are echoed as they are received**.

*Particularly useful for http request smuggling vulnerability research*

```shell
httpecho
# On another shell
curl -d "param1=value1&param2=value2" -X POST http://localhost:8888/ -H "Transfer-Encoding: chunked" -H "Content-Length: 8"
POST / HTTP/1.1
Host: localhost:8888
User-Agent: curl/7.58.0
Accept: */*
Transfer-Encoding: chunked
Content-Length: 8
Content-Type: application/x-www-form-urlencoded

1b
param1=value1&param2=value2
0

```

## Usage
```shell
Usage of httpecho: echo server accepting malformed HTTP request
  -s --serve serve continuously (default: wait for 1 request)
  -t, --timeout timeout to close connection. Needed for closing http request. (default: 1)
  -d, --dump dump incoming request to a file (default: only print to stdout)
  -p, --port listening on specific port (default: 8888)
  -h, --help dump incoming request to a file (default: only print to stdout)
```

## Install

```shell
curl -lO -L https://github.com/ariary/httpecho/releases/latest/download/httpecho
#or
go install github.com/ariary/httpecho
```
