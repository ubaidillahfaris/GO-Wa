package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsAppService struct {
	Mongo      *db.MongoService
	Client     *whatsmeow.Client
	MessageSvc *MessageService
	ctx        context.Context
	cancel     context.CancelFunc
	DeviceName string

	qrMu     sync.Mutex
	latestQR string

	IsConnected bool
	ConnectedMu sync.Mutex
	sem         chan struct{}
}

type ContactInfo struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

type GroupSummary struct {
	JID               string   `json:"jid"`
	Name              string   `json:"name"`
	Topic             string   `json:"topic"`
	Participants      []string `json:"participants"`
	AdminJIDs         []string `json:"admins"`
	MemberCount       int      `json:"member_count"`
	IsLocked          bool     `json:"is_locked"`
	IsAnnounce        bool     `json:"is_announce"`
	IsEphemeral       bool     `json:"is_ephemeral"`
	DisappearingTimer uint32   `json:"disappearing_timer"`
}

func (m *WhatsAppManager) NewWhatsAppService(parent context.Context, deviceName string) (*WhatsAppService, error) {
	ctx, cancel := context.WithCancel(parent)

	dbPath := fmt.Sprintf("file:./stores/%s_store.db?_foreign_keys=on", deviceName)
	container, err := sqlstore.New(ctx, "sqlite3", dbPath, waLog.Stdout("DB-"+deviceName, "ERROR", true))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("gagal buat sqlstore: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("gagal ambil device: %w", err)
	}
	if deviceStore == nil {
		deviceStore = container.NewDevice()
	}

	clientLog := waLog.Stdout("Client-"+deviceName, "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	svc := &WhatsAppService{
		Client:     client,
		ctx:        ctx,
		cancel:     cancel,
		DeviceName: deviceName,
		sem:        make(chan struct{}, 10),
	}

	svc.MessageSvc = NewMessageService(svc, "")

	svc.registerEventHandlers()
	return svc, nil
}

func (w *WhatsAppService) registerEventHandlers() {
	w.Client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			w.ConnectedMu.Lock()
			w.IsConnected = true
			w.ConnectedMu.Unlock()
			fmt.Printf("🟢 [%s] Connected\n", w.DeviceName)

		case *events.Disconnected:
			w.ConnectedMu.Lock()
			w.IsConnected = false
			w.ConnectedMu.Unlock()
			fmt.Printf("🔴 [%s] Disconnected\n", w.DeviceName)

		case *events.Message:
			if !v.Info.IsFromMe && v.Message.GetConversation() != "" {
				sender := v.Info.Sender.User
				msg := v.Message.GetConversation()
				fmt.Printf("📩 [%s] Pesan dari %s: %s\n", w.DeviceName, sender, msg)

				go func() {
					w.sem <- struct{}{}
					defer func() { <-w.sem }()
					w.HandleIncomingMessage(sender, msg)
				}()
			}
		}
	})
}

func (w *WhatsAppService) GenerateQR() (string, error) {
	w.qrMu.Lock()
	defer w.qrMu.Unlock()

	if w.Client.Store.ID != nil && w.Client.IsConnected() {
		return "", nil
	}

	if w.latestQR != "" {
		return w.latestQR, nil
	}

	qrChan, _ := w.Client.GetQRChannel(w.ctx)

	if err := w.Client.Connect(); err != nil {
		return "", fmt.Errorf("[%s] gagal connect: %w", w.DeviceName, err)
	}

	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			w.latestQR = evt.Code
			return evt.Code, nil
		}
		return "", fmt.Errorf("[%s] event tak dikenal: %s", w.DeviceName, evt.Event)
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("[%s] timeout menunggu QR", w.DeviceName)
	}
}

func (w *WhatsAppService) LatestQR() string {
	return w.latestQR
}

func (w *WhatsAppService) Status() string {
	if w.Client == nil {
		return "uninitialized"
	}
	if w.Client.Store.ID == nil {
		return "not_logged_in"
	}
	if w.Client.IsConnected() {
		return "connected"
	}
	return "disconnected"
}

func (w *WhatsAppService) Disconnect() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("⚠️ [%s] Panic saat disconnect: %v\n", w.DeviceName, r)
		}
	}()
	if w.Client != nil {
		w.Client.Disconnect()
	}
	w.cancel()
	fmt.Printf("🔌 [%s] Disconnected dan context dibatalkan\n", w.DeviceName)
}

func (w *WhatsAppService) ListContacts() ([]ContactInfo, error) {
	if w.Client == nil {
		return nil, fmt.Errorf("[%s] client belum diinisialisasi", w.DeviceName)
	}

	// Asumsi: ada w.Client.Store.Contacts dan method GetAllContacts
	contactsMap, err := w.Client.Store.Contacts.GetAllContacts(w.ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s] gagal ambil kontak: %w", w.DeviceName, err)
	}

	result := make([]ContactInfo, 0, len(contactsMap))
	for jid, info := range contactsMap {
		name := info.PushName // atau info.Name sesuai struct library
		if name == "" {
			name = jid.User
		}
		result = append(result, ContactInfo{
			JID:  jid.String(),
			Name: name,
		})
	}

	return result, nil
}

func (w *WhatsAppService) ListGroups() ([]GroupSummary, error) {
	if w.Client == nil {
		return nil, fmt.Errorf("[%s] client belum diinisialisasi", w.DeviceName)
	}

	groupMap, err := w.Client.GetJoinedGroups(w.ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s] gagal ambil grup: %w", w.DeviceName, err)
	}

	var groups []GroupSummary
	for _, g := range groupMap {
		var participants []string
		var admins []string
		for _, p := range g.Participants {
			participants = append(participants, p.JID.String())
			if p.IsAdmin || p.IsSuperAdmin {
				admins = append(admins, p.JID.String())
			}
		}

		groups = append(groups, GroupSummary{
			JID:               g.JID.String(),
			Name:              g.GroupName.Name,
			Topic:             g.GroupTopic.Topic,
			Participants:      participants,
			AdminJIDs:         admins,
			MemberCount:       len(g.Participants),
			IsLocked:          g.GroupLocked.IsLocked,
			IsAnnounce:        g.GroupAnnounce.IsAnnounce,
			IsEphemeral:       g.GroupEphemeral.IsEphemeral,
			DisappearingTimer: g.GroupEphemeral.DisappearingTimer,
		})
	}

	return groups, nil
}
