package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type TelegramNotifier struct {
	botToken   string
	httpClient *http.Client
}

func NewTelegramNotifier(botToken string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TelegramNotifier) SendAlert(targetChatID string, message string) error {
	if t.botToken == "" {
		slog.Info("🔔 [TELEGRAM LOG (Token Belum Diset)]", "chat_id", targetChatID, "message", message)
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := map[string]string{
		"chat_id":    targetChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := t.httpClient.Post(apiURL, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("gagal mengirim pesan ke Telegram: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API mengembalikan status HTTP %d", resp.StatusCode)
	}

	slog.Info("Pesan Telegram berhasil dikirim", "chat_id", targetChatID)
	return nil
}
