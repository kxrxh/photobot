package testutil

import (
	"net/url"
	"strconv"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func BuildValidInitData(token, userJSON string) string {
	now := time.Now()
	hash := initdata.Sign(map[string]string{"user": userJSON}, token, now)
	ts := strconv.FormatInt(now.Unix(), 10)
	return "auth_date=" + ts + "&user=" + url.QueryEscape(userJSON) + "&hash=" + hash
}
