package max

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

const apiBase = "https://platform-api.max.ru"

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
	userID int64,
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

	uploadURL, err := c.getUploadURL(ctx, botToken)
	if err != nil {
		return err
	}

	attachToken, err := c.uploadFile(ctx, botToken, uploadURL, doc, filename)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	return c.sendMessage(ctx, botToken, userID, attachToken)
}

func (c *Client) getUploadURL(ctx context.Context, token string) (string, error) {
	cfg := requestConfig(ctx)
	cfg.Header = map[string]string{
		"Authorization": token,
	}
	resp, err := c.httpClient.Post(apiBase+"/uploads?type=file", cfg)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Close()

	bodyBytes := resp.Body()
	if resp.StatusCode() != fasthttp.StatusOK {
		return "", fmt.Errorf(
			"get upload URL: MAX API returned %d: %s",
			resp.StatusCode(),
			string(bodyBytes),
		)
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if result.URL == "" {
		return "", errors.New("MAX API returned empty upload URL")
	}
	return result.URL, nil
}

func (c *Client) uploadFile(
	ctx context.Context,
	token, uploadURL string,
	doc []byte,
	filename string,
) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	part, err := w.CreateFormFile("data", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(doc)); err != nil {
		return "", fmt.Errorf("write document: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("close multipart: %w", err)
	}

	cfg := requestConfig(ctx)
	cfg.Header = map[string]string{
		"Authorization": token,
		"Content-Type":  w.FormDataContentType(),
	}
	cfg.Body = body.Bytes()
	resp, err := c.httpClient.Post(uploadURL, cfg)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Close()

	bodyBytes := resp.Body()
	if resp.StatusCode() != fasthttp.StatusOK {
		return "", fmt.Errorf(
			"upload file: MAX API returned %d: %s",
			resp.StatusCode(),
			string(bodyBytes),
		)
	}

	var withToken struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(bodyBytes, &withToken); err == nil && withToken.Token != "" {
		return withToken.Token, nil
	}
	return string(bodyBytes), nil
}

func (c *Client) sendMessage(
	ctx context.Context,
	token string,
	userID int64,
	attachToken string,
) error {
	backoff := retryaddon.NewExponentialBackoff(retryaddon.Config{
		InitialInterval: 2 * time.Second,
		MaxBackoffTime:  8 * time.Second,
		Multiplier:      2,
		MaxRetryCount:   5,
	})

	var finalErr error
	err := backoff.Retry(func() error {
		if err := ctx.Err(); err != nil {
			finalErr = err
			return nil
		}

		retryable, err := c.sendMessageOnce(ctx, token, userID, attachToken)
		if err == nil {
			finalErr = nil
			return nil
		}
		if !retryable {
			finalErr = err
			return nil
		}
		return err
	})
	if err != nil {
		return err
	}
	return finalErr
}

func (c *Client) sendMessageOnce(
	ctx context.Context,
	token string,
	userID int64,
	attachToken string,
) (bool, error) {
	payload := struct {
		Text        string        `json:"text"`
		Attachments []interface{} `json:"attachments"`
	}{
		Text: "Отчёт по анализу",
		Attachments: []interface{}{
			map[string]interface{}{
				"type": "file",
				"payload": map[string]interface{}{
					"token": attachToken,
				},
			},
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/messages?user_id=%s", apiBase, strconv.FormatInt(userID, 10))
	cfg := requestConfig(ctx)
	cfg.Header = map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}
	cfg.Body = bodyBytes
	resp, err := c.httpClient.Post(url, cfg)
	if err != nil {
		return true, fmt.Errorf("send message failed: %w", err)
	}
	defer resp.Close()

	respBody := resp.Body()

	var result struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(respBody, &result)
	if result.Error != nil && result.Error.Code == "attachment.not.ready" {
		return true, fmt.Errorf("MAX API error: %s - %s", result.Error.Code, result.Error.Message)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return false, fmt.Errorf(
			"send message: MAX API returned %d: %s",
			resp.StatusCode(),
			string(respBody),
		)
	}

	if result.Error != nil {
		return false, fmt.Errorf("MAX API error: %s - %s", result.Error.Code, result.Error.Message)
	}
	return false, nil
}
