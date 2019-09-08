package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	STATSFINAL = `
Stats: 
total[ %d ], success[ %d ], losted[ %d ], failed[ %d ]
`
)

var (
	hookCommand = `export ip=%s; %s`

	lostedIPs sync.Map
	resultIPs sync.Map
	doneCount int32
)

type commandResult struct {
	ok      bool
	slimErr string
}

type mrStats struct {
	successNum  int
	successHost string

	totalFailedNum  int
	totalFailedHost string

	varyFailedNum  map[string]int
	varyFailedHost map[string]string

	lostedNum  int
	lostedHost string
}

func runIPs() {
	var wg sync.WaitGroup
	var tk Token

	NewTokenPool(gPoolSize, gExpirySecs)

	for _, ip := range IPs {
		wg.Add(1)

		lostedIPs.Store(ip, struct{}{})

		if debugMode {
			fmt.Printf("Borrow token: ip[ %s ], freeTokenCount[ %d ]\n", ip, Tp.FreeCount())
		}
		tk = Tp.Borrow()

		go doCommand(ip, tk, wg.Done)
	}

	wg.Wait()
}

func doCommand(ip string, tk Token, deferFunc func()) {
	defer func() {
		deferFunc()
	}()

	if debugMode {
		fmt.Printf(
			"Got token: ip[ %s ], tokenNum[ %d ], freeTokenCount[ %d ]\n",
			ip, tk.QueueNum, Tp.FreeCount())
	}

	cmd := fmt.Sprintf(hookCommand, ip, userCommand)
	// fmt.Println(cmd)

	var debugMsg string

	res, err := execShell(cmd)
	if err == "" {
		resultIPs.Store(ip, commandResult{
			ok:      true,
			slimErr: "",
		})
		debugMsg = fmt.Sprintf(
			"Execute ok: ip[ %s ], res[ %s ]\n", ip, res)

	} else {
		resultIPs.Store(ip, commandResult{
			ok:      false,
			slimErr: err[:30],
		})
		debugMsg = fmt.Sprintf(
			"Execute failed: ip[ %s ], err[ %s ]\n", ip, err)
	}

	if debugMode {
		fmt.Print(debugMsg)
	}

	lostedIPs.Delete(ip)

	Tp.Back(tk)
	if debugMode {
		fmt.Printf(
			"Back token: ip[ %s ], freeTokenCount[ %d ]\n",
			ip, Tp.FreeCount())
	}

	atomic.AddInt32(&doneCount, 1)
	fmt.Printf("Processed: ip[ %s ], [ %d / %d ]\n",
		ip, doneCount, totalCount)
}

func runStats() {
	mrs := &mrStats{
		successNum:  0,
		successHost: "",

		totalFailedNum:  0,
		totalFailedHost: "",

		varyFailedNum:  make(map[string]int),
		varyFailedHost: make(map[string]string),

		lostedNum:  0,
		lostedHost: "",
	}

	fmt.Printf(STATSFINAL,
		totalCount, mrs.successNum, mrs.lostedNum, mrs.totalFailedNum)

	retryFile := "./retry.ips"

	if mrs.totalFailedNum > 0 {
		writeFile(mrs.totalFailedHost, retryFile)
	}

	if mrs.lostedNum > 0 {
		appendFile(mrs.lostedHost, retryFile)
	}
}

func (mrs *mrStats) statNormal(ips sync.Map) {
	ips.Range(func(k, v interface{}) bool {
		ip, ok := k.(string)
		if !ok {
			return true
		}
		res, ok := v.(commandResult)
		if !ok {
			return true
		}

		if res.ok {
			mrs.successNum++
			mrs.successHost = mrs.successHost + fmt.Sprintf("%s\n", ip)
		} else {
			mrs.totalFailedNum++
			mrs.totalFailedHost = mrs.totalFailedHost + fmt.Sprintf("%s\n", ip)

			mrs.varyFailedNum[res.slimErr]++
			mrs.varyFailedHost[res.slimErr] = mrs.varyFailedHost[res.slimErr] + fmt.Sprintf("%s\n", ip)
		}

		return true
	})
}

func (mrs *mrStats) statLosted(ips sync.Map) {
	ips.Range(func(k, v interface{}) bool {
		ip, ok := k.(string)
		if !ok {
			return true
		}

		mrs.lostedNum++
		mrs.lostedHost = mrs.lostedHost + fmt.Sprintf("%s\n", ip)

		return true
	})
}
