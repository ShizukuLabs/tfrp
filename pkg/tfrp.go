package pkg

import (
	"crypto/md5"
	"fmt"
	"time"
)

type TfrpService struct {
	id              string
	localIp         string
	localPort       string
	remotePort      string
	Type            string
	plugin          string
	pluginLocalPath string
}

type TfrpClient struct {
	ip      string
	name    string
	online  bool
	service []*TfrpService
}

type Tfrp struct {
	addr        string
	port        string
	token       string
	adminUser   string
	adminPasswd string
	clients     []*TfrpClient
}

// 生成frp配置文件的common部分
func (t *Tfrp) GenFrpCommonCfg() (config string) {
	config = fmt.Sprintf(`[common]
server_addr = %s
server_port = %s
token = %s
`)
	return
}

// 生成frp配置文件的client部分
func (t *Tfrp) GenFrpClientCfg(c *TfrpClient) (config string) {
	for _, service := range c.service {
		config += fmt.Sprintf(`[%s]
type = %s
local_ip = %s
local_port = %s
remote_port = %s
`, service.id, service.Type, service.localIp, service.localPort, service.remotePort)
		if service.plugin != "" {
			config += fmt.Sprintf(`plugin = %s
plugin_local_path = %s
`, service.plugin, service.pluginLocalPath)
		}
	}
	return
}

// 生成frp配置文件
func (t *Tfrp) GenFrpCfg(c *TfrpClient) string {
	return t.GenFrpCommonCfg() + t.GenFrpClientCfg(c)
}

// 查找TfrpClient
func (t *Tfrp) FindClient(name string) *TfrpClient {
	for _, c := range t.clients {
		if c.name == name {
			return c
		}
	}
	return nil
}

// 查找TfrpService
func (t *Tfrp) FindService(name string) *TfrpService {
	for _, c := range t.clients {
		for _, s := range c.service {
			if s.id == name {
				return s
			}
		}
	}
	return nil
}

// 添加TfrpClient
func (t *Tfrp) AddClient(c *TfrpClient) {
	t.clients = append(t.clients, c)
}

// 添加TfrpService
func (t *Tfrp) AddService(c *TfrpClient, s *TfrpService) {
	c.service = append(c.service, s)
}

// 删除TfrpClient
func (t *Tfrp) DelClient(name string) {
	for i, c := range t.clients {
		if c.name == name {
			t.clients = append(t.clients[:i], t.clients[i+1:]...)
			return
		}
	}
}

// 删除TfrpService
func (t *Tfrp) DelService(name string) {
	for _, c := range t.clients {
		for i, s := range c.service {
			if s.id == name {
				c.service = append(c.service[:i], c.service[i+1:]...)
				return
			}
		}
	}
}

// 更新TfrpClient
func (t *Tfrp) UpdateClient(c *TfrpClient) {
	for i, client := range t.clients {
		if client.name == c.name {
			t.clients[i] = c
			return
		}
	}
}

// 更新TfrpService
func (t *Tfrp) UpdateService(s *TfrpService) {
	for _, c := range t.clients {
		for i, service := range c.service {
			if service.id == s.id {
				c.service[i] = s
				return
			}
		}
	}
}

// 新建
func NewTfrp(addr, port, token, adminUser, adminPasswd string) *Tfrp {
	return &Tfrp{
		addr:        addr,
		port:        port,
		token:       token,
		adminUser:   adminUser,
		adminPasswd: adminPasswd,
	}
}

// 新建TfrpClient
func NewTfrpClient(name string) *TfrpClient {
	return &TfrpClient{
		name: name,
	}
}

// 新建TfrpService
func NewTfrpService(id, local_ip, local_port, remote_port, Type, plugin, plugin_local_path string) *TfrpService {
	return &TfrpService{
		id:              id,
		localIp:         local_ip,
		localPort:       local_port,
		remotePort:      remote_port,
		Type:            Type,
		plugin:          plugin,
		pluginLocalPath: plugin_local_path,
	}
}

// 生成id
func GenID() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().GoString())))
}
