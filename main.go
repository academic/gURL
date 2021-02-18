package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"time"
)

// VERBOSE ...
var VERBOSE = 0

// Connect loop
func connect(start, done chan bool, url string) {

	for {
		<-start

		tr := &http.Transport{}
		client := &http.Client{Transport: tr}

		r, err := client.Get(url)

		if err != nil {
			log.Fatal(err)
			done <- true
			return
		}

		defer r.Body.Close()

		// Print the body
		if VERBOSE > 0 {

			body, err := ioutil.ReadAll(r.Body)

			if err != nil {
				log.Fatal(err)
				done <- true
				return
			}

			fmt.Printf(string(body))
		}

		done <- true
	}
}

// Goroutine to keep consuming <-done
func requestsDone(done, end chan bool, n, c int) {

	startTime := time.Now().UnixNano()
	lastNow := startTime

	for i := 0; i < n; i++ {
		<-done
		if i%c == 0 && i != 0 {
			now := time.Now().UnixNano()
			interval := (now - lastNow) / 1e6
			fmt.Printf("%d reqs done. + %d msecs\n", i, interval)
			lastNow = now
		}
	}

	interval := (time.Now().UnixNano() - lastNow) / 1e6
	fmt.Printf("%d reqs done. + %d msecs\n", n, interval)

	totalTime := time.Now().UnixNano() - startTime
	fmt.Printf("\ntotal time: %f secs\n", float64(totalTime)/1e9)

	end <- true
}

func main() {

	url := flag.String("url", "", "url to connect to: 'https://www.google.com'")
	n := flag.Int("n", 1, "total number of requests")
	c := flag.Int("c", 1, "number of parallel requests")

	verbose := flag.Int("v", 0, "verbosity")

	flag.Parse()

	VERBOSE = *verbose

	start := make(chan bool, *n)
	done := make(chan bool, *n)

	end := make(chan bool, *n)

	// just one request
	if *n == 1 && *c == 1 {

		if VERBOSE == 0 {
			VERBOSE = 1
		}

		go connect(start, done, *url)
		start <- true
		<-done

		return
	}

	// Multiple Requests

	// Setting Max Procs to the Number of CPU Cores
	fmt.Printf("Max procs %d\n", runtime.GOMAXPROCS(runtime.NumCPU()))
	fmt.Printf("Max procs %d\n\n", runtime.GOMAXPROCS(0))

	go requestsDone(done, end, *n, *c)

	for i := 0; i < *c; i++ {
		go connect(start, done, *url)
		// start some goroutines immediately
		start <- true
	}

	for i := *c; i < *n; i++ {
		// fill in the chan so everybody can work
		start <- true
	}

	// wait for all the requests to be terminated
	<-end
}
