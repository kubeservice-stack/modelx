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
	"net/url"
	"sort"
	"strconv"
)

func Keys(maps map[string]interface{}) []string {
	var keys []string
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

func Values(maps map[string]interface{}) []interface{} {
	var values []interface{}
	for _, v := range maps {
		values = append(values, v)
	}
	return values
}

func Sort(maps map[string]string) map[string]string {
	ret := make(map[string]string)

	priorities := make([]string, 0)
	for key := range maps {
		priorities = append(priorities, key)
	}
	sort.Strings(priorities)
	for _, key := range priorities {
		ret[key] = maps[key]
	}

	return ret
}

func SortKey(maps []string) []string {
	priorities := make([]int, 0)
	for _, key := range maps {
		keyi, err := strconv.Atoi(key)
		if err != nil {
			panic(err)
		}
		priorities = append(priorities, keyi)
	}
	sort.Ints(priorities)

	ret := make([]string, 0)
	for _, ki := range priorities {
		ret = append(ret, strconv.Itoa(ki))
	}

	return ret
}

func Merge(mapsA map[string]string, mapsB map[string]string) (map[string]string, error) {
	if mapsA == nil && mapsB == nil {
		return nil, nil
	}
	ret := mapsA
	if mapsB == nil {
		return ret, nil
	}
	for key, val := range mapsB {
		ret[key] = val
	}
	return ret, nil
}

func ToParam(param url.Values) map[string]string {
	ret := make(map[string]string)
	if param == nil {
		return ret
	}

	for key, value := range param {
		if len(value) <= 0 {
			ret[key] = ""
		} else {
			ret[key] = value[0]
		}
	}
	return ret
}

func ToValues(param map[string]string) url.Values {
	ret := make(url.Values)
	if param == nil {
		return ret
	}

	for key, value := range param {
		ret[key] = []string{value}
	}
	return ret
}

func ToMapStrings(param map[string]interface{}) map[string]string {
	ret := make(map[string]string)
	if param == nil {
		return ret
	}

	for key, value := range param {
		ret[key] = value.(string)
	}
	return ret
}

func Strings(param []interface{}) map[string]bool {
	ret := make(map[string]bool)
	if param == nil {
		return ret
	}

	for _, value := range param {
		if value.(string) != "" {
			ret[value.(string)] = true
		}
	}

	return ret
}

func SStrings(param []string) map[string]bool {
	ret := make(map[string]bool)
	if param == nil {
		return ret
	}

	for _, value := range param {
		ret[value] = true
	}

	return ret
}
