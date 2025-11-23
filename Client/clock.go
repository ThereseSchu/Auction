package main

import "sync"

type Clock struct {
	Value int64
	Mutex sync.Mutex
}

func (clock *Clock) UpdateClock(newValue int64) {
	clock.Mutex.Lock()
	defer clock.Mutex.Unlock()
	if clock.Value < newValue {
		clock.Value = newValue
	}
}

func (clock *Clock) Increment() {
	clock.Mutex.Lock()
	defer clock.Mutex.Unlock()
	clock.Value++
}

func (clock *Clock) GetTime() int64 {
	clock.Mutex.Lock()
	defer clock.Mutex.Unlock()
	return clock.Value
}

// Constructor
func NewClock() *Clock {
	return &Clock{
		Value: 0,
	}
}