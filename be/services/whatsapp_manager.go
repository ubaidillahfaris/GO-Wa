package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

type WhatsAppManager struct {
	instances map[string]*WhatsAppService
	mu        sync.RWMutex
}

var (
	manager     *WhatsAppManager
	managerOnce sync.Once
)

func NewWhatsAppManager() *WhatsAppManager {
	m := &WhatsAppManager{
		instances: make(map[string]*WhatsAppService),
	}

	storesDir := "./stores"

	if err := os.MkdirAll(storesDir, 0755); err != nil {
		panic("❌ gagal buat folder stores: " + err.Error())
	}

	files, err := os.ReadDir(storesDir)
	if err != nil {
		fmt.Println("⚠️ gagal membaca folder stores:", err)
		return m
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "_store.db") {
			continue
		}

		deviceName := strings.TrimSuffix(f.Name(), "_store.db")
		ctx := context.Background()

		svc, err := m.NewWhatsAppService(ctx, deviceName)
		if err != nil {
			fmt.Printf("❌ gagal load device %s: %v\n", deviceName, err)
			continue
		}

		m.instances[deviceName] = svc
		fmt.Printf("✅ berhasil load device: %s\n", deviceName)
	}

	return m
}

func GetWhatsAppManager() *WhatsAppManager {
	managerOnce.Do(func() {
		manager = NewWhatsAppManager()
	})
	return manager
}

func (m *WhatsAppManager) GetOrCreateDevice(ctx context.Context, deviceName string) (*WhatsAppService, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if svc, ok := m.instances[deviceName]; ok {
		return svc, nil
	}

	svc, err := m.NewWhatsAppService(ctx, deviceName)
	if err != nil {
		return nil, err
	}

	m.instances[deviceName] = svc
	return svc, nil
}

func (m *WhatsAppManager) RemoveDevice(deviceName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if svc, ok := m.instances[deviceName]; ok {
		svc.Disconnect()
		delete(m.instances, deviceName)
		fmt.Printf("🧹 device %s dihapus dari manager\n", deviceName)
	}
}

func (m *WhatsAppManager) ListDevices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]string, 0, len(m.instances))
	for name := range m.instances {
		devices = append(devices, name)
	}
	return devices
}
