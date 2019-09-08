package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var (
	IPs = []string{}

	totalCount int

	cleanupDone = make(chan struct{})

	sig os.Signal

	wg sync.WaitGroup

	// parameters
	help      bool
	debugMode bool

	hostStr  string
	hostFile string

	userCommand string
	gPoolSize   int
)

func init() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&debugMode, "v", false, "verbose mode, show more info")

	flag.StringVar(&hostStr, "p", "", "host IP(s) string(split by comma(,))")
	flag.StringVar(&hostFile, "i", "", "host IP(s) file(one ip line by line)")

	flag.StringVar(&userCommand, "c", "", "a command required which could get a hook param $ip")

	flag.IntVar(&gPoolSize, "g", DefaultTokenPoolSize, "goroutine pool size")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Mad Runner: 
Usage: ./madrun -c command [-hv] [-g concurrent_num] [-p IP(s)] [-i IP(s) file]`+"\n"+`
Options:
`)
		flag.PrintDefaults()
		fmt.Println()
	}
}

func main() {
	flag.Parse()

	if help || userCommand == "" {
		flag.Usage()
		return
	}

	if hostStr != "" {
		IPs = strings.Split(hostStr, ",")

	} else if hostFile != "" {
		for _, ip := range readLines(hostFile) {
			if ip != "" && ip != "\n" {
				IPs = append(IPs, ip)
			}
		}

	} else {
		flag.Usage()
		return
	}

	setupCloseHandler()

	totalCount = len(IPs)

	beforeTime := time.Now()
	go runIPs()

	for {
		select {
		case <-cleanupDone:
			goto loopend
		}
	}
loopend:
	t1 := time.Since(beforeTime)

	beforeTime = time.Now()
	runStats()
	t2 := time.Since(beforeTime)

	fmt.Printf("elapsed: runIPs[ %s ] runStats[ %s ]\n", t1, t2)
	return
}

func setupCloseHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		sig = <-signalChan
		defer close(cleanupDone)

		fmt.Println("\nDetected Ctrl+C, please wait a sec...")

		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()

		for range ticker.C {
			fmt.Printf("SpinCheck: FreeCount[ %d ], WorkCount[ %d ]\n", Tp.FreeCount(), Tp.WorkCount)
			if Tp.FreeCount()+int(Tp.WorkCount) == gPoolSize {
				break
			}
		}

		Tp.ResetTokenPool(wg.Done)
	}()
}
