package sub

import (
	"crypto/md5"
	"fmt"
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Hot reload",
	RunE: func(cmd *cobra.Command, args []string) error {
		go hotReload(cfgApi, cfgApiSecret)
		return nil
	},
}

func hotReload(cfgApi string, cfgApiSecret string) error {
	hash := ""
	for {
		cfg, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgApi + "/" + cfgApiSecret)
		if err != nil || fmt.Sprintf("%x", md5.Sum(cfg.CfgBody)) == hash {
			time.Sleep(60 * time.Second)
			continue
		} else {
			hash = fmt.Sprintf("%x", md5.Sum(cfg.CfgBody))
			svr.Close()
			svr, err = client.NewService(cfg, pxyCfgs, visitorCfgs, cfgApi+"/"+cfgApiSecret)
			if err != nil {
				time.Sleep(60 * time.Second)
				continue
			}
			go svr.Run()
			log.Printf("Reloaded config file: %s", cfgApi+"/"+cfgApiSecret)
		}
		time.Sleep(3 * time.Second)
	}
}
