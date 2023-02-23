package main

import (
	md52 "crypto/md5"
	"encoding/json"
	"fmt"
	net2 "github.com/fatedier/frp/pkg/util/net"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
	"time"
)

type regCmd struct {
	Cmd      string `json:"cmd"`
	HostName string `json:"hostname"`
}

func readFile(fileanme string) string {
	file, err := os.Open(fileanme)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf := make([]byte, 1024)
	n, _ := file.Read(buf)
	buf = buf[:n]
	return string(buf)
}

var filename = "/Users/yonezawayukari/GolandProjects/tfrp/conf/frpc.ini"
var cfgBody = readFile(filename)
var md5 = md52.Sum([]byte(cfgBody))

func main() {
	listenHost := ":8001"
	e := echo.New()
	go func() {
		for {
			body := readFile(filename)
			md5 := md52.Sum([]byte(body))
			if string(md5[:]) != string(cfgBody[:]) {
				cfgBody = body
				log.Printf("frpc.ini changed")
			}
			time.Sleep(5 * time.Second)
		}
	}()
	e.GET("/frp/:cfgApiSecret", func(c echo.Context) error {
		cfgApiSecret := c.Param("cfgApiSecret")
		log.Printf("cfgApiSecret:%s", cfgApiSecret)
		cfgBody, err := net2.DesECBEncrypt([]byte(cfgBody), net2.AesCipherKey)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, string(cfgBody))
	})
	e.GET("/create/:cfgApiSecret", func(c echo.Context) error {
		// build frpc
		//cfgApiSecret := c.Param("cfgApiSecret")
		return c.String(http.StatusOK, fmt.Sprintf(""))
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
