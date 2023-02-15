package main

import (
	"encoding/json"
	"fmt"
	net2 "github.com/fatedier/frp/pkg/util/net"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type regCmd struct {
	Cmd      string `json:"cmd"`
	HostName string `json:"hostname"`
}

func main() {
	listenHost := ":8001"
	e := echo.New()
	e.GET("/frp/:cfgApiSecret", func(c echo.Context) error {
		cfgApiSecret := c.Param("cfgApiSecret")
		log.Printf("cfgApiSecret:%s", cfgApiSecret)
		cfgBody := []byte(`
[common]
server_addr = 127.0.0.1
server_port = 10000

[download2]
type = tcp
local_ip = 127.0.0.1
local_port = 10001
remote_port = 10002

`)
		cfgBody, err := net2.DesECBEncrypt(cfgBody, net2.AesCipherKey)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, string(cfgBody))
	})
	e.POST("/frp/:cfgApiSecret", func(c echo.Context) error {
		cfgApiSecret := c.Param("cfgApiSecret")
		cmdReq := new(regCmd)
		body := c.Request().Body
		defer body.Close()
		reqBody := make([]byte, 1024)
		n, _ := body.Read(reqBody)
		reqBody = reqBody[:n]
		reqBody, _ = net2.DesECBDecrypt(reqBody, net2.AesCipherKey)
		_ = json.Unmarshal(reqBody, cmdReq)
		log.Printf("cfgApiSecret:%s, cmd:%s, hostname:%s", cfgApiSecret, cmdReq.Cmd, cmdReq.HostName)
		return c.String(http.StatusOK, fmt.Sprintf(""))
	})
	e.Logger.Fatal(e.Start(listenHost))
}
