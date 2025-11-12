// Package models
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type QuickResponse struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	Petugas                PetugasInfo        `bson:"petugas"`
	IdentifikasiKegiatanQR KegiatanQRInfo     `bson:"identifikasi_kegiatan_qr"`
	OutputKegiatanQR       OutputQRInfo       `bson:"output_kegiatan_qr"`
	CreatedAt              int64              `bson:"created_at"`
}

type PetugasInfo struct {
	Nama        string `bson:"nama"`
	Jabatan     string `bson:"jabatan"`
	DiPenugasan string `bson:"di_penugasan"`
}

type KegiatanQRInfo struct {
	MetodePenugasan    string `bson:"metode_penugasan"`
	KegiatanQR         string `bson:"kegiatan_qr"`
	DIQR               string `bson:"di_qr"`
	SaluranQR          string `bson:"saluran_qr"`
	RuasBangunanQR     string `bson:"ruas_bangunan_qr"`
	DesaKecamatanKabQR string `bson:"desa_kecamatan_kab_qr"`
	UPTPSDAWS          string `bson:"upt_psda_ws"`
}

type OutputQRInfo struct {
	LuasAreaKegiatan  string `bson:"luas_area_kegiatan"`
	PanjangSaluran    string `bson:"panjang_saluran"`
	MenutupBocoran    string `bson:"menutup_bocoran"`
	AngkatSedimen     string `bson:"angkat_sedimen"`
	PembersihanSampah string `bson:"pembersihan_sampah"`
	AngkatPotongPohon string `bson:"angkat_potong_pohon"`
}
