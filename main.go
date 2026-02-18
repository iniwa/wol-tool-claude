package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// --- データ構造 ---

type Device struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	MAC     string `json:"mac"`
	IP      string `json:"ip"`    // ping監視用（空欄可）
	Online  bool   `json:"online"`
	LastSeen string `json:"last_seen"`
}

type Store struct {
	mu      sync.RWMutex
	Devices []*Device `json:"devices"`
	path    string
}

func NewStore(path string) *Store {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, s)
	}
	if s.Devices == nil {
		s.Devices = []*Device{}
	}
	return s
}

func (s *Store) Save() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(s.path, data, 0644)
}

func (s *Store) Add(d *Device) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	s.Devices = append(s.Devices, d)
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, d := range s.Devices {
		if d.ID == id {
			s.Devices = append(s.Devices[:i], s.Devices[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) Get(id string) *Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.Devices {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func (s *Store) All() []*Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Device, len(s.Devices))
	copy(out, s.Devices)
	return out
}

// --- WoL ---

func sendMagicPacket(mac string) error {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("invalid MAC: %w", err)
	}

	packet := make([]byte, 102)
	// 6バイトのFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	// MACアドレスを16回繰り返す
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:], hw)
	}

	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(packet)
	return err
}

// --- Ping ---

func pingHost(ip string) bool {
	if ip == "" {
		return false
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	}
	return cmd.Run() == nil
}

func (s *Store) startPingLoop() {
	go func() {
		for {
			for _, d := range s.All() {
				if d.IP == "" {
					continue
				}
				online := pingHost(d.IP)
				s.mu.Lock()
				d.Online = online
				if online {
					d.LastSeen = time.Now().Format("2006-01-02 15:04:05")
				}
				s.mu.Unlock()
			}
			s.Save()
			time.Sleep(30 * time.Second)
		}
	}()
}

// --- HTTP ハンドラ ---

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func jsonResp(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func setupRoutes(store *Store) *http.ServeMux {
	mux := http.NewServeMux()

	// WebUI (静的ファイル)
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// GET /api/devices — デバイス一覧
	mux.HandleFunc("/api/devices", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonResp(w, store.All(), 200)

		case http.MethodPost:
			var d Device
			if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
				jsonResp(w, map[string]string{"error": err.Error()}, 400)
				return
			}
			if d.Name == "" || d.MAC == "" {
				jsonResp(w, map[string]string{"error": "name and mac are required"}, 400)
				return
			}
			store.Add(&d)
			store.Save()
			jsonResp(w, d, 201)
		}
	}))

	// DELETE /api/devices/{id}
	mux.HandleFunc("/api/devices/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/devices/"):]
		if id == "" {
			http.NotFound(w, r)
			return
		}
		// POST /api/devices/{id}/wake
		if len(id) > 5 && id[len(id)-5:] == "/wake" {
			devID := id[:len(id)-5]
			d := store.Get(devID)
			if d == nil {
				jsonResp(w, map[string]string{"error": "not found"}, 404)
				return
			}
			if err := sendMagicPacket(d.MAC); err != nil {
				jsonResp(w, map[string]string{"error": err.Error()}, 500)
				return
			}
			log.Printf("WoL sent to %s (%s)", d.Name, d.MAC)
			jsonResp(w, map[string]string{"status": "sent"}, 200)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			if store.Delete(id) {
				store.Save()
				w.WriteHeader(204)
			} else {
				jsonResp(w, map[string]string{"error": "not found"}, 404)
			}
		}
	}))

	return mux
}

func main() {
	dataPath := "/data/devices.json"
	if p := os.Getenv("DATA_PATH"); p != "" {
		dataPath = p
	}
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	store := NewStore(dataPath)
	store.startPingLoop()

	mux := setupRoutes(store)
	addr := ":" + port
	log.Printf("WoL Tool starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
