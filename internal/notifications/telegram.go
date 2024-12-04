package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jwtly10/at4j-risk-manager/internal/config"
	"github.com/jwtly10/at4j-risk-manager/pkg/logger"
	"io"
	"net/http"
)

const TELEGRAM_URL = "https://api.telegram.org/bot"

type TelegramNotifier struct {
	cfg *config.TelegramConfig
}

func NewTelegramNotifier(cfg *config.TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{cfg: cfg}
}

type TelegramBody struct {
	ChatId    string `json:"chat_id"`
	ParseMode string `json:"parse_mode"`
	Text      string `json:"text"`
}

// NotifyError sends an error message to a telegram chat, formatted in HTML.
func (t *TelegramNotifier) NotifyError(message string, err error) {

	htmlMessage := ""
	if err == nil {
		htmlMessage = fmt.Sprintf(
			"[GO-RMS] ERROR ⚠️\n"+
				"<b>Error:</b> %s\n",
			message,
		)
	} else {
		htmlMessage = fmt.Sprintf(
			"[GO-RMS] ERROR ⚠️\n"+
				"<b>Error:</b> %s\n"+
				"<pre>%v</pre>\n",
			message,
			err,
		)
	}

	err = notifyHtml(t.cfg.Token, t.cfg.ChatId, htmlMessage)
	// If we fail, there's nothing to handle really so just log and continue
	if err != nil {
		logger.Errorf("Error sending telegram message: %v", err)
	}
}

// Notify sends a generic message to a telegram chat, formatted in HTML.
func (t *TelegramNotifier) Notify(message string) {
	htmlMessage := fmt.Sprintf(
		"[GO-RMS] 🚨\n"+
			"%s\n",
		message,
	)
	err := notifyHtml(t.cfg.Token, t.cfg.ChatId, htmlMessage)
	if err != nil {
		logger.Errorf("Error sending telegram message: %v", err)
	}
}

// notifyHTML sends a message to a telegram chat using the HTML parse mode.
func notifyHtml(token, chatId, message string) error {
	logger.Debugf("Sending telegram message: %s", message)

	url := TELEGRAM_URL + token + "/sendMessage"

	client := http.DefaultClient

	body := TelegramBody{
		ChatId:    chatId,
		ParseMode: "HTML",
		Text:      message,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling telegram body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("error creating http request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending telegram message: %v", err)
	}

	if resp.StatusCode != http.StatusOK {

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}

		return fmt.Errorf("unexpected status code from telegram: %d: res: %s", resp.StatusCode, string(b))
	}

	logger.Debugf("Telegram message sent successfully")

	return nil
}
