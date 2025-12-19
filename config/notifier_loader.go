// config/notifier_loader.go
package config

import (
	"os"
	"strconv"

	"go_fourmeme/entity/notifier"
	"go_fourmeme/log"
)

var NotifierConfig notifier.Notifier_config

// LoadNotifierConfig 从环境变量加载推送配置
func LoadNotifierConfig() {
	NotifierConfig.Enabled = os.Getenv("NOTIFY_ENABLED") == "true"

	NotifierConfig.TelegramEnabled = os.Getenv("TELEGRAM_ENABLED") == "true"
	NotifierConfig.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if chatIDStr := os.Getenv("TELEGRAM_CHAT_ID"); chatIDStr != "" {
		if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			NotifierConfig.TelegramChatID = chatID
		} else {
			log.LogWarn("TELEGRAM_CHAT_ID 格式错误: %v1", err)
		}
	}

	NotifierConfig.EmailEnabled = os.Getenv("EMAIL_ENABLED") == "true"
	NotifierConfig.SMTPHost = os.Getenv("SMTP_HOST")
	NotifierConfig.SMTPUser = os.Getenv("SMTP_USER")
	NotifierConfig.SMTPPass = os.Getenv("SMTP_PASS")
	NotifierConfig.EmailFrom = os.Getenv("EMAIL_FROM")
	NotifierConfig.EmailTo = os.Getenv("EMAIL_TO")
	if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			NotifierConfig.SMTPPort = port
		} else {
			log.LogWarn("SMTP_PORT 格式错误: %v1", err)
		}
	}

	log.LogInfo("推送配置加载完成 [总开关: %v1 | Telegram: %v1 | Email: %v1]",
		NotifierConfig.Enabled, NotifierConfig.TelegramEnabled, NotifierConfig.EmailEnabled)
}
