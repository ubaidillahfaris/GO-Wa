package quickresponse

import (
	"strings"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/internal/modules/quickresponse/domain"
)

// Parser parses WhatsApp messages into QuickResponse entities
type Parser struct{}

// NewParser creates a new QuickResponse parser
func NewParser() *Parser {
	return &Parser{}
}

// CanParse checks if a message contains Quick Response data
func (p *Parser) CanParse(message string) bool {
	// Check if message contains required sections
	return strings.Contains(message, "Data Petugas")
}

// Parse parses a WhatsApp message into a QuickResponse entity
func (p *Parser) Parse(message string) *domain.QuickResponse {
	lines := strings.Split(message, "\n")

	qr := &domain.QuickResponse{
		Officer:   domain.OfficerInfo{},
		Activity:  domain.ActivityInfo{},
		Output:    domain.OutputInfo{},
		CreatedAt: time.Now(),
	}

	section := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Determine section
		switch line {
		case "Data Petugas":
			section = "Officer"
			continue
		case "Identifikasi Kegiatan Q.R":
			section = "Activity"
			continue
		case "Output Kegiatan QR":
			section = "Output"
			continue
		}

		// Parse key:value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Map to struct based on section
		switch section {
		case "Officer":
			p.parseOfficer(key, val, &qr.Officer)
		case "Activity":
			p.parseActivity(key, val, &qr.Activity)
		case "Output":
			p.parseOutput(key, val, &qr.Output)
		}
	}

	return qr
}

// parseOfficer parses officer information
func (p *Parser) parseOfficer(key, value string, officer *domain.OfficerInfo) {
	switch key {
	case "Nama":
		officer.Name = value
	case "Jabatan":
		officer.Position = value
	case "D.I Penugasan":
		officer.Assignment = value
	}
}

// parseActivity parses activity information
func (p *Parser) parseActivity(key, value string, activity *domain.ActivityInfo) {
	switch key {
	case "Metode Penugasan":
		activity.Method = value
	case "Kegiatan Quick Respons":
		activity.ActivityType = value
	case "D.I Quick Respons":
		activity.IrrigationDI = value
	case "Saluran Quick Respons":
		activity.Channel = value
	case "Ruas Bangunan Quick Respons":
		activity.BuildingRoute = value
	case "Desa / Kecamatan / Kabupaten Quick Respons":
		activity.Location = value
	case "UPT PSDA WS":
		activity.WatershedUnit = value
	}
}

// parseOutput parses output information
func (p *Parser) parseOutput(key, value string, output *domain.OutputInfo) {
	switch key {
	case "Luas Area Kegiatan":
		output.AreaSize = value
	case "Panjang Saluran":
		output.ChannelLength = value
	case "Menutup Bocoran":
		output.LeaksClosed = value
	case "Angkat Sedimen":
		output.SedimentRemoved = value
	case "Pembersihan Sampah":
		output.TrashCleared = value
	case "Angkat / Potong Pohon":
		output.TreeCutRemoved = value
	}
}

// IsValid checks if a parsed QuickResponse is valid
func (p *Parser) IsValid(qr *domain.QuickResponse) bool {
	// At minimum, officer info must be present
	return qr.Officer.Name != "" ||
	       qr.Officer.Position != "" ||
	       qr.Officer.Assignment != ""
}
