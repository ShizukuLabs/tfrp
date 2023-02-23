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

package main

import (
	_ "github.com/fatedier/frp/assets/frpc"
	"github.com/fatedier/frp/cmd/frpc/sub"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
)

var cfgApi string
var cfgApiSecret string
var debug string

func main() {
	if debug == "" {
		log.Info("Starting frpc %s", version.Full())
		log.Info("Starting frpc %s", version.Full())
		log.Info("cfgApi: %s", cfgApi)
		log.Info("cfgApiSecret: %s", cfgApiSecret)
	}
	sub.Execute(cfgApi, cfgApiSecret, debug)
}
