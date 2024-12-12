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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDSR(t *testing.T) {
	assert := assert.New(t)
	dsr := DSR()
	assert.Equal([]byte{0x1b, 0x5b, 0x36, 0x6e}, dsr)
}

func TestSGR256FG(t *testing.T) {
	assert := assert.New(t)
	sgr := SGR256_FG(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x33, 0x38, 0x3b, 0x35, 0x3b, 0x38, 0x6d}, sgr)
}

func TestSGR256BG(t *testing.T) {
	assert := assert.New(t)
	sgr := SGR256_BG(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x34, 0x38, 0x3b, 0x35, 0x3b, 0x38, 0x6d}, sgr)
}

func TestCursor(t *testing.T) {
	assert := assert.New(t)
	cuu := CUU(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x41}, cuu)

	cuf := CUF(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x43}, cuf)

	cud := CUD(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x42}, cud)

	cub := CUB(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x44}, cub)

	cnl := CNL(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x45}, cnl)

	cpl := CNL(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x45}, cpl)

	cha := CHA(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x47}, cha)

	cup := CUP(uint8(4), uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x4, 0x3b, 0x8, 0x48}, cup)

	cht := CHT(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x49}, cht)

	ed := ED(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x4a}, ed)

	el := EL(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x4b}, el)

	su := SU(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x53}, su)

	sd := SD(uint8(8))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x54}, sd)

	hvp := HVP(uint8(8), uint8(4))
	assert.Equal([]byte{0x1b, 0x5b, 0x38, 0x3b, 0x34, 0x66}, hvp)
}
