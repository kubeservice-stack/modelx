/*
Copyright 2024 The KubeService-Stack Authors.

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

package progress

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"kubegems.io/modelx/pkg/util"

	"github.com/satori/go.uuid"
)

var SpinnerDefault = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

type Bar struct {
	Name       string
	MaxNameLen int    // max name length
	Total      int64  // total bytes, -1 for indeterminate
	Width      int    // width of the bar
	Status     string // status text
	Done       bool   // if the bar is done
	Fragments  map[string]*BarFragment

	nameindex    int // scroll name index
	refreshcount int // refresh count for scroll name
	mu           sync.RWMutex
	mp           *MultiBar
}

type BarFragment struct {
	Offset       int64  // offset of the fragment
	Processed    int64  // processed bytes
	uid          string // uid of the fragment, for delete
	nototalindex int    // index when no total
}

func (b *Bar) SetNameStatus(name, status string, done bool) {
	b.Name, b.Status, b.Done = name, status, done
	b.Notify()
}

func (b *Bar) SetStatus(status string, done bool) {
	b.Status = status
	b.Done = done
	b.Notify()
}

func (b *Bar) SetDone() {
	b.Done = true
	b.Notify()
}

func (r *Bar) Notify() {
	if r.mp != nil {
		r.mp.haschange = true
	}
}

func (b *Bar) WrapWriter(wc io.WriteCloser, name string, total int64, initStatus string) io.WriteCloser {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Name = name
	b.Total = total
	b.Status = initStatus
	b.Notify()

	if b.Fragments == nil {
		b.Fragments = make(map[string]*BarFragment)
	}
	uid := uuid.NewV4().String()
	thisfragment := &BarFragment{
		uid: uid,
	}
	b.Fragments[uid] = thisfragment

	w := &barw{fragment: thisfragment, wc: wc, b: b}
	if _, ok := wc.(io.WriterAt); ok {
		return barwa{barw: w}
	}
	return w
}

func (b *Bar) Print(w io.Writer) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	processwidth := b.Width

	buff := make([]byte, processwidth)
	status := ""
	if b.Done {
		for i := range buff {
			buff[i] = '+'
		}
		status = b.Status
	} else {
		for i := range buff {
			buff[i] = '-'
		}
		var totalProcessed int64
		for _, f := range b.Fragments {
			totalProcessed += f.Processed
			if b.Total > 0 {
				start := int(float64(processwidth) * float64(f.Offset) / float64(b.Total))
				end := int(float64(processwidth) * float64(f.Offset+f.Processed) / float64(b.Total))
				if end > processwidth {
					end = processwidth
				}
				if start < 0 {
					start = 0
				}
				for i := start; i < end; i++ {
					buff[i] = '+'
				}
			} else {
				buff[f.nototalindex%processwidth] = '+'
				f.nototalindex++
			}
		}
		if totalProcessed > 0 {
			if b.Total <= 0 {
				status = util.HumanSize(float64(totalProcessed))
			} else {
				status = util.HumanSize(float64(totalProcessed)) + "/" + util.HumanSize(float64(b.Total))
			}
		} else {
			status = b.Status
		}
	}
	showname := b.Name
	if len(b.Name) > b.MaxNameLen {
		b.mp.haschange = true // force print
		fullname := b.Name + "  "
		lowptr := b.nameindex % len(fullname)
		maxptr := lowptr + b.MaxNameLen
		if maxptr < len(fullname) {
			showname = fullname[lowptr:maxptr]
		} else {
			showname = fullname[lowptr:] + fullname[:maxptr-len(fullname)]
		}
		// 3x speed low than fps
		if b.refreshcount%3 == 0 {
			b.nameindex++
		}
		b.refreshcount++
	} else if len(showname) < b.MaxNameLen {
		showname += strings.Repeat(" ", b.MaxNameLen-len(showname))
	}
	fmt.Fprintf(w, "%s [%s] %s\n", showname, string(buff), status)
}

func (b *Bar) WrapReader(rc io.ReadSeekCloser, name string, total int64, initStatus string) io.ReadSeekCloser {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Total = total
	b.Name = name
	b.Status = initStatus
	b.Done = false // reset done
	defer b.Notify()

	if b.Fragments == nil {
		b.Fragments = make(map[string]*BarFragment)
	}
	uid := uuid.NewV4().String()
	thisfragment := &BarFragment{
		uid: uid,
	}
	b.Fragments[uid] = thisfragment

	return &barr{fragment: thisfragment, rc: rc, b: b}
}

func Percent(total, completed int64) int {
	if total <= 0 {
		return 0
	}
	if completed >= total {
		return 100
	}
	round := math.Round(float64(completed) / float64(total) * 100)
	if round > 100 {
		return 100
	}
	if round < 0 {
		return 0
	}
	return int(round)
}
