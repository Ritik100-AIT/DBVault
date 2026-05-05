package notify

import "fmt"

type SlackNotifier struct {
	WebhookURL string
}

func NewSlackNotifier(url string) *SlackNotifier {
	return &SlackNotifier{WebhookURL: url}
}

func (s *SlackNotifier) Notify(message string) error {
	fmt.Println("slack notify stub:", message)
	return nil
}
