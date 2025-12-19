// notifier/notifier.go
package notifier

import "go_fourmeme/log"

type Notifier interface {
	Send(title, message string) error
}

// multiNotifier 多渠道组合（任一成功即算成功）
type multiNotifier struct {
	notifiers []Notifier
}

func (m *multiNotifier) Send(title, message string) error {
	var lastErr error
	for _, n := range m.notifiers {
		if err := n.Send(title, message); err != nil {
			lastErr = err
			log.LogWarn("推送渠道失败: %v", err)
			continue
		}
		log.LogInfo("推送成功: %s", title)
		return nil // 任一成功即返回
	}
	return lastErr
}
