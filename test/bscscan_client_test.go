package test

import (
	"go_fourmeme/client"
	"go_fourmeme/log"
	"testing"
)

func TestGetTokenHolderList(t *testing.T) {
	Before()

	holders, total, err := client.GetTokenHolderList("0xb43676bf5dfac5014fb97eb1d87187839f894444", 1, 100)
	if err != nil {
		log.LogError("获取持有人失败: %v", err)
	} else {
		log.LogInfo("持有人数: %d, top100: %d", total, len(holders))
	}
}
