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

package orders

type Order interface {
	Size() int
	Empty() bool
	Values() []interface{}
	String() string
}

type OrderMap interface {
	Put(key interface{}, value interface{})
	Get(key interface{}) (value interface{}, found bool)
	Remove(key interface{})
	Keys() []interface{}

	Order
}

type OrderSet interface {
	Add(items ...interface{})
	Contains(items ...interface{}) bool
	Remove(items ...interface{})

	Order
}
