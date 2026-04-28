package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- データ構造 ---

type Device struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Online       bool   `json:"online"`
	LastSeen     string `json:"last_seen"`
	ShutdownUser string `json:"shutdown_user"`
	ShutdownPass string `json:"shutdown_pass"`
}

// DeviceView は API レスポンス用の構造体。パスワードを含まず、
// パスワードが設定済みかどうかだけを has_shutdown_pass で示す。
type DeviceView struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	MAC              string `json:"mac"`
	IP               string `json:"ip"`
	Online           bool   `json:"online"`
	LastSeen         string `json:"last_seen"`
	ShutdownUser     string `json:"shutdown_user"`
	HasShutdownPass  bool   `json:"has_shutdown_pass"`
}

func toView(d *Device) DeviceView {
	return DeviceView{
		ID:              d.ID,
		Name:            d.Name,
		MAC:             d.MAC,
		IP:              d.IP,
		Online:          d.Online,
		LastSeen:        d.LastSeen,
		ShutdownUser:    d.ShutdownUser,
		HasShutdownPass: d.ShutdownPass != "",
	}
}

func toViews(ds []*Device) []DeviceView {
	out := make([]DeviceView, len(ds))
	for i, d := range ds {
		out[i] = toView(d)
	}
	return out
}

type Store struct {
	mu           sync.RWMutex
	Devices      []*Device `json:"devices"`
	PingInterval int       `json:"ping_interval"`
	path         string
}

func NewStore(path string, defaultInterval int) *Store {
	s := &Store{path: path, PingInterval: defaultInterval}
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
	os.WriteFile(s.path, data, 0600)
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

// Update は指定IDのデバイスを更新する。
// shutdownPass が空文字の場合は既存の値を保持する（UIに平文を露出させないため）。
// shutdownPass を明示的にクリアしたい場合は clearPass=true を指定する。
func (s *Store) Update(id, name, mac, ip, shutdownUser, shutdownPass string, clearPass bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.Devices {
		if d.ID == id {
			d.Name = name
			d.MAC = mac
			d.IP = ip
			d.ShutdownUser = shutdownUser
			if clearPass {
				d.ShutdownPass = ""
			} else if shutdownPass != "" {
				d.ShutdownPass = shutdownPass
			}
			return true
		}
	}
	return false
}

func (s *Store) All() []*Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Device, len(s.Devices))
	copy(out, s.Devices)
	return out
}

func (s *Store) GetPingInterval() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PingInterval
}

func (s *Store) SetPingInterval(sec int) {
	s.mu.Lock()
	s.PingInterval = sec
	s.mu.Unlock()
}

// --- バリデーション ---

func validIP(ip string) bool {
	if ip == "" {
		return true
	}
	return net.ParseIP(ip) != nil
}

func validMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}

// --- WoL ---

func sendMagicPacket(mac string) error {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("invalid MAC: %w", err)
	}

	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
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
	if !validIP(ip) || ip == "" {
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

func (s *Store) pingAllDevices() {
	devs := s.All()
	var wg sync.WaitGroup
	for _, d := range devs {
		if d.IP == "" {
			continue
		}
		wg.Add(1)
		d := d
		go func() {
			defer wg.Done()
			online := pingHost(d.IP)
			s.mu.Lock()
			d.Online = online
			if online {
				d.LastSeen = time.Now().Format("2006-01-02 15:04:05")
			}
			s.mu.Unlock()
		}()
	}
	wg.Wait()
	s.Save()
}

func (s *Store) pingOneDevice(id string) *Device {
	d := s.Get(id)
	if d == nil || d.IP == "" {
		return d
	}
	online := pingHost(d.IP)
	s.mu.Lock()
	d.Online = online
	if online {
		d.LastSeen = time.Now().Format("2006-01-02 15:04:05")
	}
	s.mu.Unlock()
	s.Save()
	return d
}

func (s *Store) startPingLoop() {
	go func() {
		for {
			interval := s.GetPingInterval()
			if interval > 0 {
				s.pingAllDevices()
				time.Sleep(time.Duration(interval) * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

// --- リモートシャットダウン ---

func shutdownWindows(ip, user, pass string) error {
	if !validIP(ip) || ip == "" {
		return fmt.Errorf("有効な IP アドレスが設定されていません")
	}
	if user == "" || pass == "" {
		return fmt.Errorf("シャットダウン用の認証情報が設定されていません")
	}
	// 引数注入対策: ユーザー名にハイフン始まりや空白を許可しない
	if strings.HasPrefix(user, "-") || strings.ContainsAny(user, " \t\n") {
		return fmt.Errorf("invalid username")
	}
	// パスワードは argv ではなく stdin で渡し、プロセスリストから秘匿する
	cmd := exec.Command("net", "rpc", "shutdown", "-I", ip, "-U", user, "-f")
	cmd.Stdin = strings.NewReader(pass + "\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shutdown failed: %s (%w)", string(output), err)
	}
	return nil
}

// --- HTTP ハンドラ ---

func jsonResp(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// basicAuth は AUTH_USER / AUTH_PASS が両方設定されている場合のみ Basic 認証を要求する。
// 未設定の場合は警告ログを出してスルーする（信頼された LAN 内専用想定）。
func basicAuth(user, pass string, h http.Handler) http.Handler {
	if user == "" || pass == "" {
		return h
	}
	expectedUser := []byte(user)
	expectedPass := []byte(pass)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), expectedUser) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), expectedPass) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="WoL Tool"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func setupRoutes(store *Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonResp(w, map[string]int{"ping_interval": store.GetPingInterval()}, 200)
		case http.MethodPut:
			var body struct {
				PingInterval int `json:"ping_interval"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				jsonResp(w, map[string]string{"error": err.Error()}, 400)
				return
			}
			if body.PingInterval < 0 {
				jsonResp(w, map[string]string{"error": "ping_interval must be >= 0"}, 400)
				return
			}
			store.SetPingInterval(body.PingInterval)
			store.Save()
			jsonResp(w, map[string]int{"ping_interval": body.PingInterval}, 200)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/ping/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		store.pingAllDevices()
		jsonResp(w, toViews(store.All()), 200)
	})

	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonResp(w, toViews(store.All()), 200)

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
			if !validMAC(d.MAC) {
				jsonResp(w, map[string]string{"error": "invalid MAC address"}, 400)
				return
			}
			if !validIP(d.IP) {
				jsonResp(w, map[string]string{"error": "invalid IP address"}, 400)
				return
			}
			store.Add(&d)
			store.Save()
			jsonResp(w, toView(&d), 201)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/devices/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/devices/"):]
		if id == "" {
			http.NotFound(w, r)
			return
		}

		// POST /api/devices/{id}/wake
		if strings.HasSuffix(id, "/wake") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			devID := strings.TrimSuffix(id, "/wake")
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

		// POST /api/devices/{id}/shutdown
		if strings.HasSuffix(id, "/shutdown") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			devID := strings.TrimSuffix(id, "/shutdown")
			d := store.Get(devID)
			if d == nil {
				jsonResp(w, map[string]string{"error": "not found"}, 404)
				return
			}
			if err := shutdownWindows(d.IP, d.ShutdownUser, d.ShutdownPass); err != nil {
				jsonResp(w, map[string]string{"error": err.Error()}, 500)
				return
			}
			log.Printf("Shutdown sent to %s (%s)", d.Name, d.IP)
			jsonResp(w, map[string]string{"status": "sent"}, 200)
			return
		}

		// POST /api/devices/{id}/ping
		if strings.HasSuffix(id, "/ping") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			devID := strings.TrimSuffix(id, "/ping")
			d := store.Get(devID)
			if d == nil {
				jsonResp(w, map[string]string{"error": "not found"}, 404)
				return
			}
			if d.IP == "" {
				jsonResp(w, map[string]string{"error": "IPアドレスが設定されていません"}, 400)
				return
			}
			updated := store.pingOneDevice(devID)
			jsonResp(w, toView(updated), 200)
			return
		}

		switch r.Method {
		case http.MethodPut:
			var body struct {
				Name           string `json:"name"`
				MAC            string `json:"mac"`
				IP             string `json:"ip"`
				ShutdownUser   string `json:"shutdown_user"`
				ShutdownPass   string `json:"shutdown_pass"`
				ClearShutdownPass bool `json:"clear_shutdown_pass"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				jsonResp(w, map[string]string{"error": err.Error()}, 400)
				return
			}
			if body.Name == "" || body.MAC == "" {
				jsonResp(w, map[string]string{"error": "name and mac are required"}, 400)
				return
			}
			if !validMAC(body.MAC) {
				jsonResp(w, map[string]string{"error": "invalid MAC address"}, 400)
				return
			}
			if !validIP(body.IP) {
				jsonResp(w, map[string]string{"error": "invalid IP address"}, 400)
				return
			}
			if store.Update(id, body.Name, body.MAC, body.IP, body.ShutdownUser, body.ShutdownPass, body.ClearShutdownPass) {
				store.Save()
				jsonResp(w, toView(store.Get(id)), 200)
			} else {
				jsonResp(w, map[string]string{"error": "not found"}, 404)
			}

		case http.MethodDelete:
			if store.Delete(id) {
				store.Save()
				w.WriteHeader(204)
			} else {
				jsonResp(w, map[string]string{"error": "not found"}, 404)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

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
	pingInterval := 30
	if v := os.Getenv("PING_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			pingInterval = n
		}
	}

	authUser := os.Getenv("AUTH_USER")
	authPass := os.Getenv("AUTH_PASS")
	if authUser == "" || authPass == "" {
		log.Println("WARNING: AUTH_USER/AUTH_PASS が未設定です。認証なしで起動します。LAN 内専用で運用してください。")
	}

	store := NewStore(dataPath, pingInterval)
	store.startPingLoop()

	mux := setupRoutes(store)
	handler := basicAuth(authUser, authPass, mux)

	addr := ":" + port
	log.Printf("WoL Tool starting on %s (ping interval: %ds)", addr, store.GetPingInterval())
	log.Fatal(http.ListenAndServe(addr, handler))
}
