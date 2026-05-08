package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SlackNotifier struct {
	WebhookURL string
	client     *http.Client
}

func NewSlackNotifier(url string) *SlackNotifier {
	return &SlackNotifier{
		WebhookURL: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SlackNotifier) Notify(message string) error {
	if strings.TrimSpace(s.WebhookURL) == "" {
		return errors.New("slack webhook is not configured")
	}

	payload := map[string]string{"text": message}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack notification failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}
