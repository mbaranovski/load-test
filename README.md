# load-test
Simple, single executable CLI load-test.

## Installation
```
go get github.com/mbaranovski/load-test
```

## Usage
```
load-test http://facebook.com
```

## Options
```
Usage: load-test [...options] <url>
      -n  Total number of requests. Default: 100
      -c  Concurrency level. Cannot be smaller than number of requests. Default: 50
      -rps  Number of requests per second. Default: 50
      -i Prints detailed info about each request. Default: false
      -disable-compression  Disable compression. Default: false
      -disable-keepalive    Prevents re-use of TCP connections between requests. Default: false
```