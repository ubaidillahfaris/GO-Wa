package domain

import "time"

// QuickResponse represents a field work report from irrigation officers
type QuickResponse struct {
	ID        string
	Officer   OfficerInfo
	Activity  ActivityInfo
	Output    OutputInfo
	CreatedAt time.Time
}

// OfficerInfo contains information about the field officer
type OfficerInfo struct {
	Name       string // Nama petugas
	Position   string // Jabatan
	Assignment string // D.I Penugasan (irrigation area assignment)
}

// ActivityInfo contains information about the activity being reported
type ActivityInfo struct {
	Method        string // Metode Penugasan
	ActivityType  string // Kegiatan Quick Respons
	IrrigationDI  string // D.I Quick Respons
	Channel       string // Saluran Quick Respons
	BuildingRoute string // Ruas Bangunan Quick Respons
	Location      string // Desa / Kecamatan / Kabupaten Quick Respons
	WatershedUnit string // UPT PSDA WS
}

// OutputInfo contains the output/results of the activity
type OutputInfo struct {
	AreaSize        string // Luas Area Kegiatan
	ChannelLength   string // Panjang Saluran
	LeaksClosed     string // Menutup Bocoran
	SedimentRemoved string // Angkat Sedimen
	TrashCleared    string // Pembersihan Sampah
	TreeCutRemoved  string // Angkat / Potong Pohon
}

// QuickResponseRepository defines the contract for QuickResponse persistence
type QuickResponseRepository interface {
	// Save saves a quick response report
	Save(qr *QuickResponse) error

	// FindByID retrieves a quick response by ID
	FindByID(id string) (*QuickResponse, error)

	// FindAll retrieves all quick responses with pagination
	FindAll(skip, limit int) ([]*QuickResponse, error)

	// Delete removes a quick response
	Delete(id string) error

	// Count counts total quick responses
	Count() (int64, error)
}
