package main

import (
	"fmt"
	"sync"
	"time"
)

/*
	问题：参考 Hystrix 实现一个滑动窗口计数器。
*/

func main() {
	counter := NewRollingWindowCounter(int64(time.Second*3), 10)
	for i := 0; i < 4; i++ {
		go func() {
			for {
				counter.Add()
			}
		}()
	}

	for {
		time.Sleep(time.Second)
		fmt.Println(counter.Tack())
	}
}

type rollingWindowCounter struct {
	sync.Mutex
	windowSize  int64
	bucketSize  int64
	bucketCount int64
	startTime   int64
	counter     []int64
	index       int64
}

func NewRollingWindowCounter(windowSize, bucketCount int64) *rollingWindowCounter {
	return &rollingWindowCounter{
		windowSize:  windowSize,
		bucketSize:  windowSize / bucketCount,
		bucketCount: bucketCount,
		startTime:   time.Now().UnixNano(),
		counter:     make([]int64, bucketCount),
	}
}

func (rwc *rollingWindowCounter) Add() {
	rwc.Lock()
	defer rwc.Unlock()

	rwc.rolling()
	rwc.counter[rwc.index]++
}

func (rwc *rollingWindowCounter) Tack() int64 {
	rwc.Lock()
	defer rwc.Unlock()

	rwc.rolling()
	var value int64
	for _, v := range rwc.counter {
		value += v
	}
	return value
}

//滚动去除过期数据
func (rwc *rollingWindowCounter) rolling() {
	now := time.Now().UnixNano()
	windowNum := max(now-rwc.startTime-rwc.bucketSize, 0) / (rwc.bucketSize)
	if windowNum == 0 {
		return
	}

	rolling := min(windowNum, rwc.bucketCount)

	for i := int64(0); i < rolling; i++ {
		rwc.index = (rwc.index + 1) % rwc.bucketCount
		rwc.counter[rwc.index] = 0
	}
	rwc.startTime += windowNum * (rwc.bucketSize)
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
