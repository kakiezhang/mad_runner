package main

import (
	"time"
)

type Token struct {
	takeTime int64 // 被取走的时间
	QueueNum int
}

func (t *Token) setTakeTime() {
	t.takeTime = time.Now().Unix()
}

func (t *Token) clearTakeTime() {
	t.takeTime = 0
}
