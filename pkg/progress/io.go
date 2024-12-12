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
	"io"
)

type barr struct {
	fragment *BarFragment
	rc       io.ReadSeekCloser
	haserr   bool
	b        *Bar
}

func (r *barr) Seek(offset int64, whence int) (int64, error) {
	r.fragment.Processed = 0 // reset processed
	n, err := r.rc.Seek(offset, whence)
	if err != nil {
		r.b.Status = "failed"
		r.b.Done = true
	}
	switch whence {
	case io.SeekStart:
		r.fragment.Offset = n
	case io.SeekCurrent:
		r.fragment.Offset += n
	case io.SeekEnd:
		r.fragment.Offset = r.b.Total - n
	}
	r.b.mp.haschange = true
	return n, err
}

func (r *barr) Read(p []byte) (int, error) {
	n, err := r.rc.Read(p)
	if err != nil && err != io.EOF {
		r.b.Status = "failed"
		r.b.Done = true
		r.haserr = true
	}
	r.fragment.Processed += int64(n)
	r.b.mp.haschange = true
	return n, err
}

func (r *barr) Close() error {
	if r.haserr {
		r.b.mu.Lock()
		defer r.b.mu.Unlock()
		delete(r.b.Fragments, r.fragment.uid)
	}

	return r.rc.Close()
}

type barw struct {
	fragment *BarFragment
	wc       io.WriteCloser
	b        *Bar
	haserr   bool
}

func (r *barw) Write(p []byte) (int, error) {
	n, err := r.wc.Write(p)
	if err != nil && err != io.EOF {
		r.b.Done = true
		r.b.Status = "failed"
		r.haserr = true
	}
	r.fragment.Processed += int64(n)
	r.b.mp.haschange = true
	return n, err
}

func (r *barw) Close() error {
	if r.haserr {
		r.b.mu.Lock()
		defer r.b.mu.Unlock()
		delete(r.b.Fragments, r.fragment.uid)
	}
	return r.wc.Close()
}

type barwa struct {
	*barw
}

func (r barwa) WriteAt(p []byte, off int64) (int, error) {
	wat, ok := r.wc.(io.WriterAt)
	if !ok {
		return 0, io.ErrUnexpectedEOF
	}
	n, err := wat.WriteAt(p, off)
	if err != nil {
		r.b.Done = true
		r.b.Status = "failed"
		r.b.mp.haschange = true
		return n, err
	}
	r.fragment.Processed += int64(n)
	r.b.mp.haschange = true
	return n, nil
}
