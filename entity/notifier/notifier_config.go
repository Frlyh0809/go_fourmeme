// entity/notifier/notifier_config.go
package notifier

type Notifier_config struct {
	Enabled          bool   `env:"NOTIFY_ENABLED"`
	TelegramEnabled  bool   `env:"TELEGRAM_ENABLED"`
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID   int64  `env:"TELEGRAM_CHAT_ID"`

	EmailEnabled bool   `env:"EMAIL_ENABLED"`
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT"`
	SMTPUser     string `env:"SMTP_USER"`
	SMTPPass     string `env:"SMTP_PASS"`
	EmailFrom    string `env:"EMAIL_FROM"`
	EmailTo      string `env:"EMAIL_TO"`
}
