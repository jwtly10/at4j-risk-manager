//go:build integration

package notifications

import (
	"fmt"
	"github.com/jwtly10/at4j-risk-manager/internal/config"
	"github.com/jwtly10/at4j-risk-manager/pkg/logger"
	"os"
	"strconv"
	"testing"
)

func TestTelegramNotifier_NotifyError(t *testing.T) {
	logger.InitLogger()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		t.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be set in env")
	}

	n := TelegramNotifier{
		cfg: &config.TelegramConfig{
			Token:  token,
			ChatId: chatID,
		},
	}

	// Create a fake error
	_, egErr := strconv.Atoi("string")
	n.NotifyError("Something failed with err", egErr)

	// Test nil error
	n.NotifyError("Something failed with no err", nil)
}

func TestTelegramNotifier_Notify(t *testing.T) {
	logger.InitLogger()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		t.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be set in env")
	}

	n := TelegramNotifier{
		cfg: &config.TelegramConfig{
			Token:  token,
			ChatId: chatID,
		},
	}

	n.Notify(fmt.Sprintf("Equity updated for broker %s: %.2f", "Some name", 21939.32))
}

func TestTelegramNotifier_BadAuth(t *testing.T) {
	logger.InitLogger()

	token := "invalid"
	chatID := "invalid"

	n := TelegramNotifier{
		cfg: &config.TelegramConfig{
			Token:  token,
			ChatId: chatID,
		},
	}

	// Create a fake error
	_, egErr := strconv.Atoi("string")

	n.NotifyError("Something failed", egErr)
}
