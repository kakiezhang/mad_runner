package main

import (
	// "fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultTokenPoolSize   = 200
	DefaultTokenExpireSecs = 3
)

var (
	Tp *TokenPool
)

type TokenPool struct {
	queue     chan Token
	WorkQueue sync.Map
	WorkCount int32
	size      int
	release   int32
}

func NewTokenPool(size, expirySecs int) {
	if size <= 0 {
		size = DefaultTokenPoolSize
	}
	if expirySecs <= 0 {
		expirySecs = DefaultTokenExpireSecs
	}

	Tp = &TokenPool{
		queue: make(chan Token, size),
		size:  size,
	}

	for i := 1; i <= Tp.size; i++ {
		t := Token{
			QueueNum: i,
		}
		Tp.queue <- t
	}

	go Tp.PriodicExpire(expirySecs)
}

func (tp *TokenPool) PriodicExpire(secs int) {
	expiry := time.Duration(secs) * time.Second
	heartbeat := time.NewTicker(expiry)
	defer heartbeat.Stop()

	for range heartbeat.C {
		currentTime := time.Now().Unix()
		// fmt.Println(currentTime)

		if atomic.LoadInt32(&(tp.release)) == 1 {
			break
		}

		tp.WorkQueue.Range(func(k, v interface{}) bool {
			// fmt.Println(k)
			// fmt.Println(v)
			t, ok := v.(Token)
			if !ok {
				return true
			}

			// fmt.Printf("takeTime: %d\n", t.takeTime)

			if t.takeTime <= 0 {
				return true
			}
			if currentTime-t.takeTime <= int64(secs) {
				return true
			}

			tp.Back(t)

			return true
		})
	}
}

func (tp *TokenPool) Back(t Token) {
	tp.WorkQueue.Delete(t.QueueNum)
	t.clearTakeTime()
	tp.queue <- t
	atomic.AddInt32(&(tp.WorkCount), -1)

	if tp.WorkCount == 0 {
		tp.Release()
	}
}

func (tp *TokenPool) Borrow() Token {
	t := <-tp.queue
	t.setTakeTime()
	tp.WorkQueue.Store(t.QueueNum, t)
	atomic.AddInt32(&(tp.WorkCount), 1)
	return t
}

func (tp *TokenPool) FreeCount() int {
	return len(tp.queue)
}

func (tp *TokenPool) Release() {
	atomic.StoreInt32(&(tp.release), 1)
}
