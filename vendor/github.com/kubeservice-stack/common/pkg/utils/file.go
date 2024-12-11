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
	"os"
	"path/filepath"
)

var (
	mkdirAllFunc  = os.MkdirAll
	removeAllFunc = os.RemoveAll
	removeFunc    = os.Remove
)

func MkDirIfNotExist(path string) error {
	if !Exist(path) {
		if e := mkdirAllFunc(path, os.ModePerm); e != nil {
			return e
		}
	}
	return nil
}

func RemoveDir(path string) error {
	if Exist(path) {
		if e := removeAllFunc(path); e != nil {
			return e
		}
	}
	return nil
}

func RemoveFile(file string) error {
	if Exist(file) {
		if e := removeFunc(file); e != nil {
			return e
		}
	}
	return nil
}

func MkDir(path string) error {
	if e := mkdirAllFunc(path, os.ModePerm); e != nil {
		return e
	}
	return nil
}

func ListDir(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}

func Exist(file string) bool {
	if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func GetExistPath(path string) string {
	if Exist(path) {
		return path
	}
	dir, _ := filepath.Split(path)
	length := len(dir)
	if length > 0 && os.IsPathSeparator(dir[length-1]) {
		dir = dir[:length-1]
	}
	return GetExistPath(dir)
}

func Path() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

func Pwd() string {
	str, _ := os.Getwd()
	return str
}

func Dir() string {
	return filepath.Dir(Path())
}

func SearchFile(filename string, paths ...string) (fullpath string, err error) {
	for _, path := range paths {
		if fullpath = filepath.Join(path, filename); Exist(fullpath) {
			return
		}
	}
	err = errors.New(fullpath + " not found in paths")
	return
}
