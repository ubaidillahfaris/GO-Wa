package services

import (
	"context"
	"fmt"
	"math/rand"
	"mime"
	"net/http"
	"path"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type MessageService struct {
	Parent    *WhatsAppService
	Signature string
}

func NewMessageService(parent *WhatsAppService, signature string) *MessageService {
	return &MessageService{
		Parent:    parent,
		Signature: signature,
	}
}

// SendMessageSafe handles sending messages to user or group
func (s *MessageService) SendMessage(ctx context.Context, jidStr string, text string, typing bool, receiverType string, messageType string, fileData []byte, filename, caption string) error {
	client := s.Parent.Client
	deviceName := s.Parent.DeviceName

	// ✅ Cek login
	if client.Store.ID == nil {
		return fmt.Errorf("[%s] belum login, scan QR dulu", deviceName)
	}

	// ✅ Auto reconnect
	if !client.IsConnected() {
		if err := client.Connect(); err != nil {
			return fmt.Errorf("[%s] gagal reconnect: %w", deviceName, err)
		}
		time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
	}

	var msg *waE2E.Message

	switch messageType {
	case "text":
		msg = &waE2E.Message{
			Conversation: &text,
		}

	case "file":
		if len(fileData) == 0 || filename == "" {
			return fmt.Errorf("[%s] fileData atau filename kosong", deviceName)
		}

		// Upload file ke WhatsApp
		resp, err := client.Upload(ctx, fileData, whatsmeow.MediaDocument)
		if err != nil {
			return fmt.Errorf("[%s] gagal upload file: %w", deviceName, err)
		}

		mimeType := http.DetectContentType(fileData)
		if mimeType == "application/octet-stream" {
			mimeType = mime.TypeByExtension(path.Ext(filename))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
		}

		docMsg := &waE2E.DocumentMessage{
			Caption:         &caption,
			FileName:        &filename,
			Mimetype:        &mimeType,
			URL:             &resp.URL,
			DirectPath:      &resp.DirectPath,
			MediaKey:        resp.MediaKey,
			FileEncSHA256:   resp.FileEncSHA256,
			FileSHA256:      resp.FileSHA256,
			FileLength:      &resp.FileLength,
			ThumbnailWidth:  proto.Uint32(50),
			ThumbnailHeight: proto.Uint32(50),
		}

		msg = &waE2E.Message{
			DocumentMessage: docMsg,
		}

	default:
		return fmt.Errorf("[%s] messageType tidak valid: %s", deviceName, messageType)
	}

	switch receiverType {
	case "user":
		return s.sendUserMessage(ctx, jidStr, msg, typing)
	case "group":
		return s.sendGroupMessage(ctx, jidStr, msg, typing)
	default:
		return fmt.Errorf("[%s] receiverType tidak valid: %s", deviceName, receiverType)
	}
}

func (s *MessageService) sendUserMessage(ctx context.Context, jidStr string, msg *waE2E.Message, typing bool) error {
	client := s.Parent.Client
	deviceName := s.Parent.DeviceName

	targetJID := types.NewJID(jidStr, types.DefaultUserServer)

	// Kirim presence & efek typing
	_ = client.SendPresence(types.PresenceAvailable)
	if typing {
		_ = client.SendChatPresence(targetJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
		time.Sleep(time.Duration(rand.Intn(1000)+700) * time.Millisecond)
	}

	// Kirim pesan
	_, err := client.SendMessage(ctx, targetJID, msg)
	if err != nil {
		return fmt.Errorf("[%s] gagal kirim pesan ke %s: %w", deviceName, jidStr, err)
	}

	fmt.Printf("✅ [%s] Message sent to user %s: %s\n", deviceName, jidStr, msg)
	return nil
}
func (s *MessageService) sendGroupMessage(ctx context.Context, jidStr string, msg *waE2E.Message, typing bool) error {
	client := s.Parent.Client
	deviceName := s.Parent.DeviceName
	var err error
	groupJID, err := types.ParseJID(fmt.Sprintf("%s@g.us", jidStr))

	// 🔁 Retry ambil info grup hingga 3 kali
	var groupInfo *types.GroupInfo

	for i := 1; i <= 3; i++ {
		_, cancel := context.WithTimeout(ctx, 10*time.Second)
		groupInfo, err = client.GetGroupInfo(groupJID)
		cancel()
		if err == nil {
			break
		}
		fmt.Printf("⚠️ [%s] Percobaan %d ambil info grup %s gagal: %v\n", deviceName, i, jidStr, err)
		time.Sleep(time.Duration(i*2) * time.Second) // delay bertahap
	}
	if err != nil {
		return fmt.Errorf("[%s] gagal ambil info grup %s setelah 3 percobaan: %w", deviceName, jidStr, err)
	}

	// Presence & efek typing
	_ = client.SendPresence(types.PresenceAvailable)
	if typing {
		_ = client.SendChatPresence(groupJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
		time.Sleep(time.Duration(rand.Intn(1000)+700) * time.Millisecond)
	}

	// Kirim pesan ke grup
	_, err = client.SendMessage(ctx, groupJID, msg)
	if err != nil {
		return fmt.Errorf("[%s] gagal kirim pesan ke grup %s: %w", deviceName, jidStr, err)
	}

	fmt.Printf("✅ [%s] Message sent to group %s (%d member): %s\n", deviceName, jidStr, len(groupInfo.Participants), msg)
	return nil
}
