// Copyright 2016 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFull(t *testing.T) {
	assertions := assert.New(t)
	version := Full()
	arr := strings.Split(version, ".")
	assertions.Equal(3, len(arr))

	proto, err := strconv.ParseInt(arr[0], 10, 64)
	assertions.NoError(err)
	assertions.True(proto >= 0)

	major, err := strconv.ParseInt(arr[1], 10, 64)
	assertions.NoError(err)
	assertions.True(major >= 0)

	minor, err := strconv.ParseInt(arr[2], 10, 64)
	assertions.NoError(err)
	assertions.True(minor >= 0)
}

func TestVersion(t *testing.T) {
	assert := assert.New(t)
	proto := Proto(Full())
	major := Major(Full())
	minor := Minor(Full())
	parseVerion := fmt.Sprintf("%d.%d.%d", proto, major, minor)
	version := Full()
	assert.Equal(parseVerion, version)
}
