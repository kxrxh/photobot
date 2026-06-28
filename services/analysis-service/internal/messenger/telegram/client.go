package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"time"

	retryaddon "github.com/gofiber/fiber/v3/addon/retry"
	fiberclient "github.com/gofiber/fiber/v3/client"
	"github.com/valyala/fasthttp"
)

const apiBase = "https://api.telegram.org"

type Client struct {
	httpClient *fiberclient.Client
}

func NewClient(httpClient *fiberclient.Client) *Client {
	if httpClient == nil {
		httpClient = fiberclient.New()
		httpClient.SetTimeout(30 * time.Second)
	}
	return &Client{httpClient: httpClient}
}

func requestConfig(ctx context.Context) fiberclient.Config {
	timeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
		if timeout <= 0 {
			timeout = time.Second
		}
	}
	return fiberclient.Config{
		Ctx:     ctx,
		Timeout: timeout,
	}
}

func (c *Client) SendDocument(
	ctx context.Context,
	botToken string,
	chatID int64,
	doc []byte,
	filename string,
) error {
	if botToken == "" {
		return errors.New("bot token is required")
	}
	if len(doc) == 0 {
		return errors.New("document cannot be empty")
	}
	if filename == "" {
		filename = "report.pdf"
	}

	url := apiBase + "/bot" + botToken + "/sendDocument"

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	_ = w.WriteField("chat_id", strconv.FormatInt(chatID, 10))

	part, err := w.CreateFormFile("document", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(doc)); err != nil {
		return fmt.Errorf("write document: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	cfg := requestConfig(ctx)
	cfg.Header = map[string]string{
		"Content-Type": w.FormDataContentType(),
	}
	cfg.Body = body.Bytes()
	resp, err := c.postWithRetry(ctx, url, cfg)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Close()

	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("telegram API returned %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if !result.OK {
		return errors.New("telegram API returned ok=false")
	}
	return nil
}

func (c *Client) postWithRetry(
	ctx context.Context,
	url string,
	cfg fiberclient.Config,
) (*fiberclient.Response, error) {
	backoff := retryaddon.NewExponentialBackoff(retryaddon.Config{
		InitialInterval: 1 * time.Second,
		MaxBackoffTime:  4 * time.Second,
		Multiplier:      2,
		MaxRetryCount:   3,
	})

	var resp *fiberclient.Response
	var finalErr error
	err := backoff.Retry(func() error {
		if err := ctx.Err(); err != nil {
			finalErr = err
			return nil
		}

		var err error
		resp, err = c.httpClient.Post(url, cfg)
		if err != nil {
			return err
		}
		if resp.StatusCode() == fasthttp.StatusTooManyRequests ||
			resp.StatusCode() >= fasthttp.StatusInternalServerError {
			resp.Close()
			return fmt.Errorf("telegram retryable status: %d", resp.StatusCode())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if finalErr != nil {
		return nil, finalErr
	}
	return resp, nil
}
