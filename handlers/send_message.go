package handlers

import (
	"github.com/ubaidillahfaris/whatsapp.git/services"
)

type SendMessageHandler struct {
	ManagerSvc *services.WhatsAppManager
}

func NewSendMessageHandler(manager *services.WhatsAppManager) *SendMessageHandler {
	return &SendMessageHandler{
		ManagerSvc: services.NewWhatsAppManager(),
	}
}
