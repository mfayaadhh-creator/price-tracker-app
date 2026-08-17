package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LoginSession merepresentasikan sesi autentikasi Telegram sementara
type LoginSession struct {
	Code      string    `json:"code"`
	Verified  bool      `json:"verified"`
	UserPhone string    `json:"user_phone"` // Telegram Chat ID
	FirstName string    `json:"first_name"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type TelegramAuthManager struct {
	botToken string
	botUser  string
	mu       sync.RWMutex
	sessions map[string]*LoginSession
	client   *http.Client
}

func NewTelegramAuthManager(botToken string, botUser string, webhookURL string) *TelegramAuthManager {
	if botUser == "" {
		botUser = "mf_pricetracker_bot"
	}

	mgr := &TelegramAuthManager{
		botToken: botToken,
		botUser:  botUser,
		sessions: make(map[string]*LoginSession),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Dual-Mode: Jika webhookURL disediakan, daftarkan webhook ke Telegram.
	// Jika tidak, jalankan background long-polling listener (cocok untuk local dev).
	if botToken != "" {
		if webhookURL != "" {
			fullWebhook := strings.TrimRight(webhookURL, "/") + "/api/v1/auth/telegram/webhook"
			go func() {
				if err := mgr.SetWebhook(fullWebhook); err != nil {
					slog.Error("Gagal mendaftarkan Telegram Webhook", "url", fullWebhook, "error", err)
				} else {
					slog.Info("🚀 Telegram Webhook berhasil didaftarkan", "url", fullWebhook)
				}
			}()
		} else {
			go mgr.startTelegramUpdateListener()
		}
	}

	return mgr
}

// SetWebhook mendaftarkan endpoint webhook publik ke Telegram API
func (m *TelegramAuthManager) SetWebhook(webhookURL string) error {
	if m.botToken == "" {
		return fmt.Errorf("bot token kosong")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook?url=%s", m.botToken, webhookURL)
	resp, err := m.client.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram setWebhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// ProcessTelegramUpdate memproses payload JSON update yang dikirim oleh Telegram via Webhook
func (m *TelegramAuthManager) ProcessTelegramUpdate(body []byte) (bool, error) {
	var u struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			Text string `json:"text"`
			From struct {
				ID        int64  `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	}

	if err := json.Unmarshal(body, &u); err != nil {
		return false, err
	}

	text := strings.TrimSpace(u.Message.Text)
	if strings.Contains(text, "AUTH_") {
		parts := strings.Fields(text)
		for _, part := range parts {
			if strings.HasPrefix(part, "AUTH_") {
				authCode := strings.TrimSpace(part)
				chatID := fmt.Sprintf("%d", u.Message.Chat.ID)
				firstName := u.Message.From.FirstName
				username := u.Message.From.Username

				if m.VerifySession(authCode, chatID, firstName, username) {
					slog.Info("🎉 User berhasil login via Telegram Webhook!", "chat_id", chatID, "name", firstName, "code", authCode)

					replyMsg := fmt.Sprintf(
						"🎉 <b>Halo %s!</b>\n\n"+
							"✅ Login ke <b>Price Tracker Web</b> berhasil terverifikasi!\n"+
							"Silakan kembali ke browser Anda untuk melihat dashboard privat.",
						firstName,
					)
					m.sendReply(chatID, replyMsg)
					return true, nil
				}
				break
			}
		}
	}
	return false, nil
}

// CreateLoginSession membuat kode acak baru dan mengembalikan deep link bot Telegram
func (m *TelegramAuthManager) CreateLoginSession() (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Bersihkan sesi lama yang kadaluarsa (> 10 menit)
	now := time.Now()
	for code, s := range m.sessions {
		if now.Sub(s.CreatedAt) > 10*time.Minute {
			delete(m.sessions, code)
		}
	}

	// Buat kode unik acak (misal: AUTH_3a8f9c)
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	code := "AUTH_" + hex.EncodeToString(b)

	m.sessions[code] = &LoginSession{
		Code:      code,
		Verified:  false,
		CreatedAt: now,
	}

	deepLink := fmt.Sprintf("https://t.me/%s?start=%s", m.botUser, code)
	return code, deepLink
}

// GetSession memeriksa status sesi login
func (m *TelegramAuthManager) GetSession(code string) (*LoginSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[code]
	return s, ok
}

// VerifySession memvalidasi sesi ketika user menekan /start AUTH_xxx di bot
func (m *TelegramAuthManager) VerifySession(code, chatID, firstName, username string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[code]
	if !ok {
		return false
	}

	s.Verified = true
	s.UserPhone = chatID
	s.FirstName = firstName
	s.Username = username
	return true
}

// startTelegramUpdateListener melakukan polling berkala ke Telegram API untuk menangkap pesan /start AUTH_xxx
func (m *TelegramAuthManager) startTelegramUpdateListener() {
	lastUpdateID := 0

	for {
		time.Sleep(2 * time.Second)

		if m.botToken == "" {
			continue
		}

		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=5", m.botToken, lastUpdateID+1)
		resp, err := m.client.Get(apiURL)
		if err != nil {
			continue
		}

		var result struct {
			Ok     bool `json:"ok"`
			Result []struct {
				UpdateID int `json:"update_id"`
				Message  struct {
					Text string `json:"text"`
					From struct {
						ID        int64  `json:"id"`
						FirstName string `json:"first_name"`
						Username  string `json:"username"`
					} `json:"from"`
					Chat struct {
						ID int64 `json:"id"`
					} `json:"chat"`
				} `json:"message"`
			} `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Ok {
			for _, u := range result.Result {
				if u.UpdateID > lastUpdateID {
					lastUpdateID = u.UpdateID
				}

				text := strings.TrimSpace(u.Message.Text)
				if strings.Contains(text, "AUTH_") {
					parts := strings.Fields(text)
					for _, part := range parts {
						if strings.HasPrefix(part, "AUTH_") {
							authCode := strings.TrimSpace(part)
							chatID := fmt.Sprintf("%d", u.Message.Chat.ID)
							firstName := u.Message.From.FirstName
							username := u.Message.From.Username

							if m.VerifySession(authCode, chatID, firstName, username) {
								slog.Info("🎉 User berhasil login via Telegram Bot!", "chat_id", chatID, "name", firstName, "code", authCode)

								// Kirim pesan balasan konfirmasi ke Telegram user
								replyMsg := fmt.Sprintf(
									"🎉 <b>Halo %s!</b>\n\n"+
										"✅ Login ke <b>Price Tracker Web</b> berhasil terverifikasi!\n"+
										"Silakan kembali ke browser Anda untuk melihat dashboard privat.",
									firstName,
								)
								m.sendReply(chatID, replyMsg)
							}
							break
						}
					}
				}
			}
		}
		resp.Body.Close()
	}
}

func (m *TelegramAuthManager) sendReply(chatID, message string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", m.botToken)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	}
	b, _ := json.Marshal(payload)
	_, _ = m.client.Post(apiURL, "application/json", strings.NewReader(string(b)))
}
