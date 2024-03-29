package main

import (
	"sync"
	"sync/atomic"
)

const (
	DefaultTokenPoolSize = 200
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

type Token struct {
	QueueNum int
}

func NewTokenPool(size int) {
	if size <= 0 {
		size = DefaultTokenPoolSize
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
}

func (tp *TokenPool) ResetTokenPool(f func()) {
	tp.WorkQueue.Range(func(k, v interface{}) bool {
		t, ok := v.(Token)
		if !ok {
			return true
		}

		tp.Back(t)
		f()

		return true
	})
}

func (tp *TokenPool) Back(t Token) {
	tp.WorkQueue.Delete(t.QueueNum)
	tp.queue <- t
	atomic.AddInt32(&(tp.WorkCount), -1)

	if tp.WorkCount == 0 {
		tp.Release()
	}
}

func (tp *TokenPool) Borrow() Token {
	t := <-tp.queue
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
