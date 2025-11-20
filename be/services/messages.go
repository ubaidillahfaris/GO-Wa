package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/models"
)

func (w *WhatsAppService) HandleIncomingMessage(sender, message string) {

	// Parsing / Mapping pesan ke struct Mongo
	qr := parseMessageToQuickResponse(message, sender)
	if qr.Petugas.Nama == "" && qr.Petugas.Jabatan == "" && qr.Petugas.DiPenugasan == "" {
		fmt.Println("⚠️ Pesan dilewati karena tidak mengandung Data Petugas")
		return
	}

	// Insert ke MongoDB
	res, err := db.Mongo.InsertQuickResponse(w.ctx, &qr)
	if err != nil {
		fmt.Println("❌ Gagal simpan ke Mongo:", err)
	} else {
		fmt.Println("✅ Pesan tersimpan dengan ID:", res.InsertedID)
	}
}
func parseMessageToQuickResponse(msg string, sender string) models.QuickResponse {
	lines := strings.Split(msg, "\n")
	qr := models.QuickResponse{
		Petugas:                models.PetugasInfo{},
		IdentifikasiKegiatanQR: models.KegiatanQRInfo{},
		OutputKegiatanQR:       models.OutputQRInfo{},
		CreatedAt:              time.Now().Unix(),
	}

	section := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
			// extractValue mengembalikan nilai dari suatu baris string yang dipisahkan oleh ":"
			// Contoh: "Nama: Budi Santoso" akan mengembalikan nilai "Budi Santoso"
			// Jika tidak ada delimiter ":" maka akan mengembalikan nilai kosong.
		}

		// Tentukan section
		switch line {
		case "Data Petugas":
			section = "Petugas"
			continue
		case "Identifikasi Kegiatan Q.R":
			section = "Identifikasi"
			continue
		case "Output Kegiatan QR":
			section = "Output"
			continue
		}

		// Parsing key:value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Mapping ke struct sesuai section
		switch section {
		case "Petugas":
			switch key {
			case "Nama":
				qr.Petugas.Nama = val
			case "Jabatan":
				qr.Petugas.Jabatan = val
			case "D.I Penugasan":
				qr.Petugas.DiPenugasan = val
			}
		case "Identifikasi":
			switch key {
			case "Metode Penugasan":
				qr.IdentifikasiKegiatanQR.MetodePenugasan = val
			case "Kegiatan Quick Respons":
				qr.IdentifikasiKegiatanQR.KegiatanQR = val
			case "D.I Quick Respons":
				qr.IdentifikasiKegiatanQR.DIQR = val
			case "Saluran Quick Respons":
				qr.IdentifikasiKegiatanQR.SaluranQR = val
			case "Ruas Bangunan Quick Respons":
				qr.IdentifikasiKegiatanQR.RuasBangunanQR = val
			case "Desa / Kecamatan / Kabupaten Quick Respons":
				qr.IdentifikasiKegiatanQR.DesaKecamatanKabQR = val
			case "UPT PSDA WS":
				qr.IdentifikasiKegiatanQR.UPTPSDAWS = val
			}
		case "Output":
			switch key {
			case "Luas Area Kegiatan":
				qr.OutputKegiatanQR.LuasAreaKegiatan = val
			case "Panjang Saluran":
				qr.OutputKegiatanQR.PanjangSaluran = val
			case "Menutup Bocoran":
				qr.OutputKegiatanQR.MenutupBocoran = val
			case "Angkat Sedimen":
				qr.OutputKegiatanQR.AngkatSedimen = val
			case "Pembersihan Sampah":
				qr.OutputKegiatanQR.PembersihanSampah = val
			case "Angkat / Potong Pohon":
				qr.OutputKegiatanQR.AngkatPotongPohon = val
			}
		}
	}

	return qr
}

/*************  ✨ Windsurf Command ⭐  *************/
/*******  c653c39f-5e64-4e37-ab05-7fb18afc12d8  *******/
func extractValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
