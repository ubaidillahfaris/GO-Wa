package services

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/util/random"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

const WebMessageIDPrefix = "3EB0"

var pbSerializer = store.SignalProtobufSerializer

type SendService struct {
	WhatsappSvc *WhatsAppService
}

type MessageDebugTimings struct {
	LIDFetch time.Duration
	Queue    time.Duration

	Marshal         time.Duration
	GetParticipants time.Duration
	GetDevices      time.Duration
	GroupEncrypt    time.Duration
	PeerEncrypt     time.Duration

	Send  time.Duration
	Resp  time.Duration
	Retry time.Duration
}

type nodeExtraParams struct {
	botNode         *waBinary.Node
	metaNode        *waBinary.Node
	additionalNodes *[]waBinary.Node
	addressingMode  types.AddressingMode
}

func (s *SendService) NewSendService(ctx context.Context, w *WhatsAppService) *SendService {
	return &SendService{}
}

func (s *SendService) GenerateMessageID() types.MessageID {

	cli := s.WhatsappSvc.Client

	if cli != nil && cli.MessengerConfig != nil {
		return types.MessageID(strconv.FormatInt(GenerateFacebookMessageID(), 10))
	}

	data := make([]byte, 8, 8+20+16)
	binary.BigEndian.PutUint64(data, uint64(time.Now().Unix()))
	ownID := s.getOwnID()
	if !ownID.IsEmpty() {
		data = append(data, []byte(ownID.User)...)
		data = append(data, []byte("@c.us")...)
	}
	data = append(data, random.Bytes(16)...)
	hash := sha256.Sum256(data)
	return WebMessageIDPrefix + strings.ToUpper(hex.EncodeToString(hash[:9]))
}

func GenerateFacebookMessageID() int64 {
	const randomMask = (1 << 22) - 1
	return (time.Now().UnixMilli() << 22) | (int64(binary.BigEndian.Uint32(random.Bytes(4))) & randomMask)
}

func (s *SendService) getOwnID() types.JID {
	cli := s.WhatsappSvc.Client
	if cli == nil {
		return types.EmptyJID
	}
	return cli.Store.GetJID()
}

func (s *SendService) sendGroupMessage(ctx context.Context, groupJID string, message string) (*types.MessageID, error) {
	to := types.NewJID(groupJID, types.GroupServer)
	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	_, err := s.WhatsappSvc.Client.SendMessage(context.Background(), to, msg)
	if err != nil {
		return nil, fmt.Errorf("send group failed: %w", err)
	}

	return nil, nil
}
