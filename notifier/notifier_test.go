// notifier/notifier_integration_test.go
package notifier_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/log"
	"go_fourmeme/notifier"

	"github.com/joho/godotenv"
)

func TestNotifier_RealSend_WithConfig(t *testing.T) {
	// 只在本地运行（避免 CI 发送垃圾消息）
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("跳过集成测试（CI 环境）")
	}
	log.InitLogger()

	// 1. 加载项目根目录的 .env 文件
	// 注意路径：测试文件在 notifier/ 目录，.env 在项目根，所以用 "../.env"
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatalf("加载 .env 失败: %v\n请确保项目根目录存在 .env 文件", err)
	}

	// 2. 加载推送配置（调用你的 loader）
	config.LoadNotifierConfig()

	// 4. 创建真实 Notifier
	n := notifier.NewNotifier()
	if n == nil {
		t.Fatalf("Notifier 创建失败，请检查 .env 中的推送配置是否正确")
	}

	// 5. 构造测试消息
	title := "【Fourmeme 推送集成测试】"
	message := fmt.Sprintf(
		"测试成功！\n时间: %s\n配置加载正常\n渠道: Telegram=%v, Email=%v",
		time.Now().Format("2006-01-02 15:04:05"),
		config.NotifierConfig.TelegramEnabled,
		config.NotifierConfig.EmailEnabled,
	)

	// 6. 异步发送（不阻塞测试）
	done := make(chan error, 1)
	go func() {
		done <- n.Send(title, message)
	}()

	// 7. 等待结果（超时 15 秒）
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("真实推送失败: %v", err)
		} else {
			t.Log("真实推送成功！请检查你的 Telegram 和邮箱")
		}
	case <-time.After(60 * time.Second):
		t.Error("推送超时（可能网络问题或配置错误）")
	}
}
