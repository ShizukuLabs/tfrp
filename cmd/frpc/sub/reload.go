package sub

import (
	"crypto/md5"
	"fmt"
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	"log"
	"time"
)

func hotReload(svr *client.Service, cfgApi string, cfgApiSecret string) error {
	hash := ""
	for {
		cfg, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgApi + "/" + cfgApiSecret)
		if err != nil || fmt.Sprintf("%x", md5.Sum(cfg.CfgBody)) == hash {
			time.Sleep(5 * time.Second)
			continue
		} else {
			hash = fmt.Sprintf("%x", md5.Sum(cfg.CfgBody))
			err := svr.ReloadConf(pxyCfgs, visitorCfgs)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Printf("Reloaded config file: %s", cfgApi+"/"+cfgApiSecret)
		}
		time.Sleep(3 * time.Second)
	}
}
