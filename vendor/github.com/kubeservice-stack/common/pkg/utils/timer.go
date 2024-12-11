/*
Copyright 2023 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"sync"
	"time"
)

type TimerPool struct {
	pool sync.Pool
}

func NewTimerPool() *TimerPool {
	return &TimerPool{}
}

// Get returns a timer for the given duration d from the pool.
//
// Return back the timer to the pool with Put.
func (tp *TimerPool) Get(d time.Duration) *time.Timer {
	if v := tp.pool.Get(); v != nil {
		t := v.(*time.Timer)
		if t.Reset(d) {
			panic("active timer trapped to the pool!")
		}
		return t
	}
	return time.NewTimer(d)
}

// Put returns t to the pool.
//
// t cannot be accessed after returning to the pool.
func (tp *TimerPool) Put(t *time.Timer) {
	if !t.Stop() {
		// Drain t.C if it wasn't obtained by the caller yet.
		select {
		case <-t.C:
		default:
		}
	}
	tp.pool.Put(t)
}
