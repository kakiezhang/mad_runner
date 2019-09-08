package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	IPs = []string{}

	totalCount int

	// parameters
	help      bool
	debugMode bool

	hostStr  string
	hostFile string

	userCommand string
	gPoolSize   int
	gExpirySecs int
)

func init() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&debugMode, "v", false, "verbose mode, show more info")

	flag.StringVar(&hostStr, "p", "", "host IP(s) string(split by comma(,))")
	flag.StringVar(&hostFile, "i", "", "host IP(s) file(one ip line by line)")

	flag.StringVar(&userCommand, "c", "", "a command required which could get a hook param $ip")

	flag.IntVar(&gPoolSize, "g", DefaultTokenPoolSize, "goroutine pool size")
	flag.IntVar(&gExpirySecs, "exps", DefaultTokenExpireSecs, "goroutine expire seconds")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Mad Runner: 
Usage: ./madrun -c command [-hv] [-exps expiry_secs] 
[-p host string] [-i host file] [-c command] [-g concurrent_num]`+"\n"+`
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

	totalCount = len(IPs)

	beforeTime := time.Now()
	runIPs()
	t1 := time.Since(beforeTime)

	beforeTime = time.Now()
	runStats()
	t2 := time.Since(beforeTime)

	fmt.Printf("elapsed: runIPs[ %s ] runStats[ %s ]\n", t1, t2)

	return
}
