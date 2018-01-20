package main

import (
	"fmt"
	"time"
	"net/http"
	"flag"
	"os"
)

var maxGoRoutines int
var numberOfRequests int
var rps int
var disablekeepAlive bool
var disableCompression bool
var info bool
var url string

var downs = 0
var completed = 0
var startTime time.Time
var minTime = time.Duration(999 * time.Second)
var maxTime = time.Duration(0)

var netClient = &http.Client{}

var ch chan string
var done chan bool

var usage = `Usage: ./load-test [options...] <url>
Options:
  -n  Total number of requests. Default: 100
  -c  Concurrency level. Cannot be smaller than number of requests. Default: 50
  -rps  Number of requests per second. Default: 50
  -i Prints detailed info about each request. Default: false
  -disable-compression  Disable compression. Default: false
  -disable-keepalive    Prevents re-use of TCP connections between requests. Default: false
`

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.IntVar(&maxGoRoutines, "c", 50, "")
	flag.IntVar(&numberOfRequests, "n", 100, "")
	flag.IntVar(&rps, "rps", 50, "")
	flag.BoolVar(&info, "i", false, "")
	flag.BoolVar(&disablekeepAlive, "disable-compression", false, "")
	flag.BoolVar(&disableCompression, "disable-keepalive", false, "")
	flag.Parse()
	url = flag.Args()[0]
	ch = make(chan string, numberOfRequests)
	done = make(chan bool)

	tr := &http.Transport{
		MaxIdleConns:        maxGoRoutines,
		MaxIdleConnsPerHost: maxGoRoutines,
		DisableKeepAlives:   disablekeepAlive,
		DisableCompression:  disableCompression,
	}
	netClient = &http.Client{Transport: tr}
}

func main() {
	printIntro()

	// Fill the channel
	go fillChBuffer(ch)

	startTime = time.Now()

	// Spawn workers
	for i := 0; i < maxGoRoutines; i++ {
		go worker(ch, done)
	}

	// Wait for read from done channel maxGoRoutines times.
	for i := 0; i < maxGoRoutines; i++ {
		<-done
	}

	printSummary()
}

func printIntro() {
	fmt.Printf("Running load test on: '%s' with %d requests. Concurrency level of %d. Requests per second: %d ", url, numberOfRequests, maxGoRoutines, rps)
	if !disableCompression && !disablekeepAlive {
		fmt.Println("")
	}

	if disablekeepAlive {
		fmt.Printf("Keep-A-Live disabled. ")
	}

	if disableCompression {
		fmt.Printf("Compression disabled. \n")
	}
}

func printSummary() {
	fmt.Printf("\rSuccessful requests: %d | Failed requests: %d | Errors: %.2f%% \n", numberOfRequests-downs, downs, float64(downs)/float64(numberOfRequests)*100)
	fmt.Printf("Min.time: %s | Max.time: %s | Avg.time: %s | Total.time: %s \n", minTime, maxTime, (maxTime+minTime)/2, time.Since(startTime))
}

func fillChBuffer(ch chan string) {
	for k := 0; k < numberOfRequests; k++ {
		ch <- url
	}
	close(ch)
}

func checkLink(link string) {
	time.Sleep(time.Second / time.Duration(rps))
	start := time.Now()

	_, err := netClient.Get(link)

	fmt.Printf("\rProgress: %d / %d ", completed, numberOfRequests)

	if err != nil {
		downs++
		if info {
			fmt.Print(" ", link, "failed to respond", err)
		}
	}

	if err == nil {
		elapsed := time.Since(start)

		if elapsed >= maxTime {
			maxTime = elapsed
		}

		if elapsed <= minTime {
			minTime = elapsed
		}

		if info {
			fmt.Printf("took %s to respond. \n", elapsed)
		}
	}
}

func worker(c chan string, done chan bool) {
	for link := range c {
		checkLink(link)
		completed++
	}
	done <- true
}
