package sub

import (
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	"log"
	"time"
)

func hotReload(svr *client.Service, cfgApi string, cfgApiSecret string) error {
	for {
		cfg, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgApi + "/" + cfgApiSecret)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if !svr.EqualProxyConf(pxyCfgs, visitorCfgs) {
			err := svr.ReloadConf(pxyCfgs, visitorCfgs)
			log.Printf("Reloaded proxy config from %s", cfgApi)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
		}
		if !svr.EqualCommonConf(cfg) {
			err := svr.ReloadCommonConf(cfg)
			log.Printf("Reloaded common config from %s", cfgApi)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
		}
		time.Sleep(3 * time.Second)
	}
}
