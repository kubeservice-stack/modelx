/*
Copyright 2022 The KubeService-Stack Authors.

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
	"errors"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type (
	reducetype func(interface{}) interface{}
	filtertype func(interface{}) bool
)

func InSlice(v string, sl []string) bool {
	if len(sl) == 0 {
		return false
	}

	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InSliceIface(v interface{}, sl []interface{}) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InSliceIfaceToLower(v string, sl interface{}) (bool, error) {
	slArr, err := ToSlice(sl)
	if err != nil {
		return false, err
	}

	alSArr := ToStrings(slArr)

	for _, vv := range alSArr {
		if strings.EqualFold(v, vv) {
			return true, nil
		}
	}
	return false, nil
}

func SliceRandList(min, max int) []int {
	if max < min {
		min, max = max, min
	}
	length := max - min + 1
	t0 := time.Now()
	rand.Seed(int64(t0.Nanosecond()))
	list := rand.Perm(length)
	for index := range list {
		list[index] += min
	}
	return list
}

func SliceMerge(slice1, slice2 []interface{}) (c []interface{}) {
	c = append(slice1, slice2...)
	return
}

func SliceReduce(slice []interface{}, a reducetype) (dslice []interface{}) {
	for _, v := range slice {
		dslice = append(dslice, a(v))
	}
	return
}

func SliceRand(a []interface{}) (b interface{}) {
	randnum := rand.Intn(len(a))
	b = a[randnum]
	return
}

func SliceSum(intslice []int64) (sum int64) {
	for _, v := range intslice {
		sum += v
	}
	return
}

func SliceFilter(slice []interface{}, a filtertype) (ftslice []interface{}) {
	for _, v := range slice {
		if a(v) {
			ftslice = append(ftslice, v)
		}
	}
	return
}

func SliceDiff(slice1, slice2 []interface{}) (diffslice []interface{}) {
	for _, v := range slice1 {
		if !InSliceIface(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}

	for _, v1 := range slice2 {
		if !InSliceIface(v1, slice1) {
			diffslice = append(diffslice, v1)
		}
	}
	return
}

func SliceRange(start, end, step int64) (intslice []int64) {
	for i := start; i <= end; i += step {
		intslice = append(intslice, i)
	}
	return
}

func SliceShuffle(slice []interface{}) []interface{} {
	for i := 0; i < len(slice); i++ {
		a := rand.Intn(len(slice))
		b := rand.Intn(len(slice))
		slice[a], slice[b] = slice[b], slice[a]
	}
	return slice
}

func InterfacesToStrings(items []interface{}) (s []string) {
	for _, item := range items {
		s = append(s, item.(string))
	}
	return s
}

func ToStringDict(items []interface{}, key string) ([]string, error) {
	var ret []string
	for _, item := range items {
		it, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("interface{} to map[string]string err")
		}
		ret = append(ret, it[key].(string))
	}
	return ret, nil
}

func ToSlice(arr interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("toslice arr not slice")
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret, nil
}

func ToStrings(arr []interface{}) []string {
	var ret []string
	for _, value := range arr {
		if value != nil {
			ret = append(ret, value.(string))
		}
	}
	return ret
}

var (
	MININTSTR string = "0000000000000000000000"
	MAXINTSTR string = "9999999999999999999999"
)

func ReplayStr(i int, size int) string {
	tmpstr := strconv.Itoa(i)
	if len(tmpstr) < size {
		return MININTSTR[:(size-len(tmpstr))] + tmpstr
	} else {
		return tmpstr
	}
}

func ReplayMaxStr(size int) string {
	return MAXINTSTR[:size]
}
