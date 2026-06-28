package messenger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	fiberclient "github.com/gofiber/fiber/v3/client"
	"github.com/valyala/fasthttp"
)

const (
	telegramAPIBase = "https://api.telegram.org"
	maxAPIBase      = "https://platform-api.max.ru"
)

type Notifier struct {
	httpClient *fiberclient.Client
}

func NewNotifier() *Notifier {
	c := fiberclient.New()
	c.SetTimeout(15 * time.Second)
	return &Notifier{httpClient: c}
}

func (n *Notifier) SendText(
	ctx context.Context,
	platform, botToken string,
	userID int64,
	text string,
) error {
	if botToken == "" {
		return errors.New("bot token is required")
	}
	switch platform {
	case "telegram":
		return n.sendTelegramText(ctx, botToken, userID, text)
	case "max":
		return n.sendMaxText(ctx, botToken, userID, text)
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
}

func (n *Notifier) sendTelegramText(
	ctx context.Context,
	botToken string,
	chatID int64,
	text string,
) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, botToken)
	body, err := json.Marshal(map[string]string{
		"chat_id": strconv.FormatInt(chatID, 10),
		"text":    text,
	})
	if err != nil {
		return err
	}
	cfg := fiberclient.Config{Ctx: ctx, Timeout: 15 * time.Second, Body: body}
	cfg.Header = map[string]string{"Content-Type": "application/json"}
	resp, err := n.httpClient.Post(url, cfg)
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("telegram sendMessage: %d %s", resp.StatusCode(), string(resp.Body()))
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return err
	}
	if !result.OK {
		return errors.New("telegram API returned ok=false")
	}
	return nil
}

func (n *Notifier) sendMaxText(
	ctx context.Context,
	botToken string,
	userID int64,
	text string,
) error {
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/messages?user_id=%s", maxAPIBase, strconv.FormatInt(userID, 10))
	cfg := fiberclient.Config{Ctx: ctx, Timeout: 15 * time.Second, Body: body}
	cfg.Header = map[string]string{
		"Authorization": botToken,
		"Content-Type":  "application/json",
	}
	resp, err := n.httpClient.Post(url, cfg)
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("max sendMessage: %d %s", resp.StatusCode(), string(resp.Body()))
	}
	return nil
}
