// client/telegram_client.go
package client

import (
	"fmt"

	"go_fourmeme/log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramClient struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramClient(token string, chatID int64) *telegramClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.LogError("Telegram Bot 初始化失败（推送将失效）: %v", err)
		return &telegramClient{} // 返回空对象，Send 会直接返回错误
	}
	log.LogInfo("Telegram 推送客户端初始化成功")
	return &telegramClient{bot: bot, chatID: chatID}
}

func (t *telegramClient) Send(title, message string) error {
	if t.bot == nil {
		return fmt.Errorf("telegram bot 未初始化")
	}

	msg := tgbotapi.NewMessage(t.chatID, fmt.Sprintf("<b>%s</b>\n\n%s", title, message))
	msg.ParseMode = "HTML"

	_, err := t.bot.Send(msg)
	if err != nil {
		log.LogError("Telegram 推送失败: %v", err)
		return err
	}
	return nil
}
