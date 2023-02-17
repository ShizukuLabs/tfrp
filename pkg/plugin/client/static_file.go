// Copyright 2018 fatedier, fatedier@gmail.com
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

package plugin

import (
	"encoding/base64"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"os"

	frpNet "github.com/fatedier/frp/pkg/util/net"
)

const PluginStaticFile = "static_file"

func init() {
	Register(PluginStaticFile, NewStaticFilePlugin)
}

type StaticFilePlugin struct {
	localPath   string
	stripPrefix string
	httpUser    string
	httpPasswd  string

	l *Listener
	s *http.Server
}

func NewStaticFilePlugin(params map[string]string) (Plugin, error) {
	localPath := params["plugin_local_path"]
	stripPrefix := params["plugin_strip_prefix"]
	httpUser := params["plugin_http_user"]
	httpPasswd := params["plugin_http_passwd"]

	listener := NewProxyListener()

	sp := &StaticFilePlugin{
		localPath:   localPath,
		stripPrefix: stripPrefix,
		httpUser:    httpUser,
		httpPasswd:  httpPasswd,

		l: listener,
	}
	var prefix string
	if stripPrefix != "" {
		prefix = "/" + stripPrefix + "/"
	} else {
		prefix = "/"
	}

	router := mux.NewRouter()
	router.Use(frpNet.NewHTTPAuthMiddleware(httpUser, httpPasswd).Middleware)
	router.PathPrefix(prefix).Handler(frpNet.MakeHTTPGzipHandler(http.StripPrefix(prefix, http.FileServer(http.Dir(localPath))))).Methods("GET")
	router.PathPrefix(prefix).HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		path := localPath + req.URL.Path
		bodyB64Buffer, _ := io.ReadAll(req.Body)
		defer req.Body.Close()
		body, _ := base64.StdEncoding.DecodeString(string(bodyB64Buffer))
		// write file
		writer, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			_, _ = res.Write([]byte(err.Error()))
			return
		}
		_, _ = writer.Write(body)
		_ = writer.Close()
		res.Write([]byte("ok"))

	}).Methods("POST")
	sp.s = &http.Server{
		Handler: router,
	}
	go func() {
		_ = sp.s.Serve(listener)
	}()
	return sp, nil
}

func (sp *StaticFilePlugin) Handle(conn io.ReadWriteCloser, realConn net.Conn, extraBufToLocal []byte) {
	wrapConn := frpNet.WrapReadWriteCloserToConn(conn, realConn)
	_ = sp.l.PutConn(wrapConn)
}

func (sp *StaticFilePlugin) Name() string {
	return PluginStaticFile
}

func (sp *StaticFilePlugin) Close() error {
	sp.s.Close()
	sp.l.Close()
	return nil
}
