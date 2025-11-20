package services

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

type QrCode struct {
	Code string `json:"code"`
}

func (q *QrCode) EncodeCode() (string, error) {
	// generate PNG sebagai byte slice
	png, err := qrcode.Encode(q.Code, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	// convert byte slice ke base64
	b64 := base64.StdEncoding.EncodeToString(png)
	return b64, nil
}

func (q *QrCode) ToPNG() ([]byte, error) {
	return qrcode.Encode(q.Code, qrcode.Medium, 256)
}
