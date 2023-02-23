package pkg

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type TfrpTG struct {
	tfrp    *Tfrp
	wg      *sync.WaitGroup
	ctx     context.Context
	tgToken string
	tg      *tgbotapi.BotAPI
}

func NewTfrpTG(tgToken string, tfrp *Tfrp) *TfrpTG {
	return &TfrpTG{tfrp: tfrp, tgToken: tgToken}
}

// 開始監聽frps的api
func (t *TfrpTG) Start() (err error) {
	t.wg = &sync.WaitGroup{}
	t.wg.Add(1)
	go t.UpdateClientStatus()
	//創建telegram機器人
	if t.tg, err = tgbotapi.NewBotAPI(t.tgToken); err != nil {
		return err
	}
	return
}

// 等待
func (t *TfrpTG) Wait() {
	t.wg.Wait()
}

// 查看所有在線服務器
func (t *TfrpTG) GetOnlineClients() (clients []*TfrpClient) {
	for _, client := range t.tfrp.clients {
		if client.online {
			clients = append(clients, client)
		}
	}
	return
}

// 從某個服務器查看所有在線服務
func (t *TfrpTG) GetOnlineServices(client *TfrpClient) (services []*TfrpService) {
	for _, service := range client.service {
		if service.online {
			services = append(services, service)
		}
	}
	return
}

// 添加服務器
func (t *TfrpTG) AddClient(client *TfrpClient) {
	t.tfrp.AddClient(client)
}

// 添加服務
func (t *TfrpTG) AddService(name string, service *TfrpService) {
	t.tfrp.AddService(t.tfrp.FindClient(name), service)
}

// 刪除服務器
func (t *TfrpTG) DelClient(name string) {
	t.tfrp.DelClient(name)
}

// 刪除服務
func (t *TfrpTG) DelService(id string) {
	t.tfrp.DelService(id)
}

// 更新服務器
func (t *TfrpTG) UpdateClient(client *TfrpClient) {
	t.tfrp.UpdateClient(client)
}

// 更新服務
func (t *TfrpTG) UpdateService(service *TfrpService) {
	t.tfrp.UpdateService(service)
}

// 從frps的api更新服務器狀態
func (t *TfrpTG) UpdateClientStatus() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}
	}
}
