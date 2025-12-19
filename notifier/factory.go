// notifier/factory.go
package notifier

import (
	"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/log"
)

// notifier/factory.go
func NewNotifier() Notifier {
	if !config.NotifierConfig.Enabled {
		log.LogInfo("消息推送总开关已关闭")
		return nil
	}

	var notifiers []Notifier

	if config.NotifierConfig.TelegramEnabled {
		tg := client.NewTelegramClient(config.NotifierConfig.TelegramBotToken, config.NotifierConfig.TelegramChatID)
		if tg != nil { // ← 关键：只加非 nil
			log.LogInfo("Telegram 客户端创建 成功")

			notifiers = append(notifiers, tg)
		} else {
			log.LogWarn("Telegram 客户端创建失败（配置错误），跳过")
		}
	}

	if config.NotifierConfig.EmailEnabled {
		email := client.NewMailClient(config.NotifierConfig)
		if email != nil { // ← 关键
			log.LogInfo("邮箱 客户端创建 成功")
			notifiers = append(notifiers, email)
		} else {
			log.LogWarn("邮箱客户端创建失败（配置错误），跳过")
		}
	}

	if len(notifiers) == 0 {
		log.LogInfo("无可用推送渠道")
		return nil
	}

	return &multiNotifier{notifiers: notifiers}
}
