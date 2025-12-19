// client/mail_client.go
package client

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"go_fourmeme/entity/notifier"
	"go_fourmeme/log"
)

type mailClient struct {
	cfg notifier.Notifier_config
}

func NewMailClient(cfg notifier.Notifier_config) *mailClient {
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPass == "" || cfg.EmailFrom == "" || cfg.EmailTo == "" {
		log.LogWarn("邮箱配置不完整，邮箱推送禁用")
		return nil
	}
	log.LogInfo("邮箱推送客户端初始化成功 (Host: %s:%d)", cfg.SMTPHost, cfg.SMTPPort)
	return &mailClient{cfg: cfg}
}

func (m *mailClient) Send(title, message string) error {
	// 地址
	addr := fmt.Sprintf("%s:%d", m.cfg.SMTPHost, m.cfg.SMTPPort)

	// 普通 TCP 连接
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP 连接失败: %w", err)
	}
	defer c.Close()

	// 启动 TLS (STARTTLS)
	if err = c.StartTLS(&tls.Config{ServerName: m.cfg.SMTPHost}); err != nil {
		return fmt.Errorf("STARTTLS 失败: %w", err)
	}

	// 认证
	auth := smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}

	// 发件人
	if err = c.Mail(m.cfg.EmailFrom); err != nil {
		return err
	}

	// 收件人
	if err = c.Rcpt(m.cfg.EmailTo); err != nil {
		return err
	}

	// 数据
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	body := fmt.Sprintf("Subject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", title, message)
	_, err = wc.Write([]byte(body))
	if err != nil {
		return err
	}

	return c.Quit()
}
