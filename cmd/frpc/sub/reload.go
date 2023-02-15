package sub

import (
	"crypto/md5"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
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
		url := cfgApi + "/" + cfgApiSecret
		req, err := http.Get(url)
		if err != nil {
			time.Sleep(60 * time.Second)
			continue
		}
		defer req.Body.Close()
		cfgBody, _ := ioutil.ReadAll(req.Body)
		cfgBody_hash := fmt.Sprintf("%x", md5.Sum(cfgBody))
		if cfgBody_hash != hash {
			hash = cfgBody_hash
		}
		time.Sleep(3 * time.Second)
	}
}
