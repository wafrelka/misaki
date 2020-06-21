package misaki

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type SlackMessage struct {
	Text string `json:"text"`
}

type Slack struct {
	webhook_url string
}

func NewSlack(webhook_url string) *Slack {
	return &Slack{
		webhook_url: webhook_url,
	}
}

func (s *Slack) Post(text string) error {

	data := SlackMessage {
		Text: text,
	}
	payload, _ := json.Marshal(&data)
	_, err := http.PostForm(s.webhook_url, url.Values{"payload": {string(payload)}})

	return err
}
