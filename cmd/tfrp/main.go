package main

import (
	"encoding/json"
	"flag"
	"fmt"
	net2 "github.com/fatedier/frp/pkg/util/net"
	"github.com/labstack/echo/v4"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"os"
)

type regCmd struct {
	Cmd      string `json:"cmd"`
	HostName string `json:"hostname"`
}

func readFile(fileanme string) string {
	file, err := os.Open(fileanme)
	if err != nil {
		return ""
	}
	defer file.Close()
	buf := make([]byte, 4096)
	n, _ := file.Read(buf)
	buf = buf[:n]
	return string(buf)
}

var tfrpIniPath string
var tfrpPort int
var frpcPath string

func init() {
	flag.StringVar(&tfrpIniPath, "c", "tfrp.ini", "tfrp ini path")
	flag.IntVar(&tfrpPort, "p", 8001, "tfrp port")
}

func readFromSecret(cfgApiSecret string) []byte {
	return []byte(readFile(frpcPath + "/" + cfgApiSecret + ".ini"))
}

func main() {
	flag.Parse()
	log.Printf("tfrpIniPath:%s", tfrpIniPath)
	// load ini
	cfg, err := ini.Load(tfrpIniPath)
	if err != nil {
		log.Panic(err)
	}
	frpcPath = cfg.Section("common").Key("path").String()
	log.Printf("frpcPath:%s", frpcPath)
	e := echo.New()
	e.GET("/frp/:cfgApiSecret", func(c echo.Context) error {
		cfgApiSecret := c.Param("cfgApiSecret")
		log.Printf("cfgApiSecret:%s", cfgApiSecret)
		cfgBody, err := net2.DesECBEncrypt(readFromSecret(cfgApiSecret), net2.AesCipherKey)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, string(cfgBody))
	})
	e.GET("/create/:cfgApiSecret", func(c echo.Context) error {
		// build frpc
		cfgApiSecret := c.Param("cfgApiSecret")
		//make file
		create, err := os.Create(frpcPath + "/" + cfgApiSecret + ".ini")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer create.Close()
		return c.String(http.StatusOK, fmt.Sprintf(""))
	})
	e.POST("/frp/:cfgApiSecret", func(c echo.Context) error {
		cfgApiSecret := c.Param("cfgApiSecret")
		cmdReq := new(regCmd)
		body := c.Request().Body
		defer body.Close()
		reqBody := make([]byte, 4096)
		n, _ := body.Read(reqBody)
		reqBody = reqBody[:n]
		reqBody, _ = net2.DesECBDecrypt(reqBody, net2.AesCipherKey)
		_ = json.Unmarshal(reqBody, cmdReq)
		log.Printf("cfgApiSecret:%s, cmd:%s, hostname:%s", cfgApiSecret, cmdReq.Cmd, cmdReq.HostName)
		return c.String(http.StatusOK, fmt.Sprintf(""))
	})
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", tfrpPort)))
}
