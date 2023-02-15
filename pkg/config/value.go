// Copyright 2020 The frp Authors
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

package config

import (
	"bytes"
	net2 "github.com/fatedier/frp/pkg/util/net"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
)

var glbEnvs map[string]string

func init() {
	glbEnvs = make(map[string]string)
	envs := os.Environ()
	for _, env := range envs {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		glbEnvs[pair[0]] = pair[1]
	}
}

type Values struct {
	Envs map[string]string // environment vars
}

func GetValues() *Values {
	return &Values{
		Envs: glbEnvs,
	}
}

func RenderContent(in []byte) (out []byte, err error) {
	tmpl, errRet := template.New("frp").Parse(string(in))
	if errRet != nil {
		err = errRet
		return
	}

	buffer := bytes.NewBufferString("")
	v := GetValues()
	err = tmpl.Execute(buffer, v)
	if err != nil {
		return
	}
	out = buffer.Bytes()
	return
}

func GetRenderedConfFromFile(path string) (out []byte, err error) {
	var b []byte
	if !strings.Contains(path, ":") {
		b, err = os.ReadFile(path)
		if err != nil {
			return
		}
	} else {
		// get
		res, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		b, err = io.ReadAll(res.Body)
		b, _ = net2.DesECBDecrypt(b, net2.AesCipherKey)
	}

	out, err = RenderContent(b)
	return
}
