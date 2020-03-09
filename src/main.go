package main

import (
	"fmt"
	"os"
	"net/url"
	"os/exec"
	"strings"
	"net/http"
	"encoding/json"
	"time"
	"github.com/burntsushi/toml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type DirectiveConfig struct {
	Name string `toml:"name"`
	Command [][]string `toml:"cmd"`
	Output bool `toml:"output"`
}

type AwsConfig struct {
	SecretAccessKey string `toml:"secret_access_key"`
	AccessKeyId string `toml:"access_key_id"`
	Region string `toml:"region"`
	QueueUrl string `toml:"queue_url"`
}

type SlackConfig struct {
	WebhookUrl string `toml:"webhook_url"`
	Channel string `toml:"channel"`
}

type Config struct {
	Slack SlackConfig `toml: "slack"`
	Aws AwsConfig `toml:"aws"`
	Directives []DirectiveConfig `toml:"directives"`
}

type SlackMessage struct {
	Text string `json:"text"`
	ChannelId string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Timestamp string `json:"timestamp"`
}

type SlackPost struct {
	Text string `json:"text"`
	Channel string `json:"channel,omitempty"`
	ThreadTs string `json:"thread_ts,omitempty"`
}

const (
	INITIAL_BACKOFF = 1
	MAX_BACKOFF = 3600
)

func post_to_slack(webhook_url, text, channel, thread_ts string) {

	data := SlackPost {
		Text: text,
		Channel: channel,
		ThreadTs: thread_ts,
	}
	payload, _ := json.Marshal(&data)
	_, err := http.PostForm(webhook_url, url.Values{"payload": {string(payload)}})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to post to Slack: %v\n", err)
	}

}

func process_slack_message(msg_text, webhook_url, channel string, directives []DirectiveConfig) {

	var msg SlackMessage

	err := json.Unmarshal([]byte(msg_text), &msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse Slack message: %v\n", err)
		return
	}

	if channel != "" && (channel != msg.ChannelName) {
		return
	}

	fmt.Printf("message: \"%s\" (%s)\n", msg.Text, msg.ChannelName)

	if !strings.HasPrefix(msg.Text, "misaki ") {
		return
	}

	name := strings.TrimPrefix(msg.Text, "misaki ")
	var dir *DirectiveConfig = nil

	for _, d := range directives {
		if d.Name == name {
			dir = &d
			break
		}
	}

	if dir == nil {
		post_to_slack(
			webhook_url,
			fmt.Sprintf("unknown command name: %s", name),
			msg.ChannelId,
			msg.Timestamp,
		)
		return
	}

	all_out := []string{}
	reply := ""

	for _, c := range dir.Command {

		out, err := exec.Command(c[0], c[1:]...).Output()

		if err != nil {
			cj := strings.Join(c, " ")
			reply = fmt.Sprintf("[%s] error: %v\n", cj, err)
			break
		}
		all_out = append(all_out, string(out))
	}

	if reply == "" {
		if dir.Output {
			o := strings.Join(all_out, "\n")
			reply = fmt.Sprintf("OK: %s\n", o)
		} else {
			reply = "OK"
		}
	}

	post_to_slack(
		webhook_url,
		reply,
		msg.ChannelId,
		msg.Timestamp,
	)
}

func run() int {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s CONFIG_FILE\n", os.Args[0])
		return 1
	}

	config_file_path := os.Args[1]
	var config Config

	_, err := toml.DecodeFile(config_file_path, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read config file: %v\n", err)
		return 1
	}

	if config.Aws.QueueUrl == "" {
		fmt.Fprintf(os.Stderr, "error: empty SQS url\n")
		return 1
	}
	if config.Slack.WebhookUrl == "" {
		fmt.Fprintf(os.Stderr, "error: empty Slack webhook url\n")
		return 1
	}

	aws_config := aws.NewConfig()
	if config.Aws.Region != "" {
		aws_config.Region = aws.String(config.Aws.Region)
	}
	if config.Aws.SecretAccessKey != "" {
		creds := credentials.NewStaticCredentials(
			config.Aws.AccessKeyId, config.Aws.SecretAccessKey, "")
		aws_config.Credentials = creds
	}

	session, err := session.NewSession(aws_config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot start AWS session: %v\n", err)
		return 1
	}

	client := sqs.New(session)

	sqs_params := sqs.ReceiveMessageInput {
		QueueUrl: aws.String(config.Aws.QueueUrl),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds: aws.Int64(20),
	}

	backoff := INITIAL_BACKOFF

	for {

		resp, err := client.ReceiveMessage(&sqs_params)

		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to fetch SQS message: %v\n", err)
			backoff = backoff * 2
			if backoff > MAX_BACKOFF {
				backoff = MAX_BACKOFF
			}
			time.Sleep(time.Duration(backoff) * time.Second)
			continue
		}
		backoff = INITIAL_BACKOFF

		for _, msg := range resp.Messages {

			msg_body := msg.Body
			if msg_body == nil {
				fmt.Fprintf(os.Stderr, "broken SQS message?\n")
				continue
			}

			process_slack_message(
				*msg_body,
				config.Slack.WebhookUrl,
				config.Slack.Channel,
				config.Directives,
			)

			del_params := sqs.DeleteMessageInput {
				QueueUrl: aws.String(config.Aws.QueueUrl),
				ReceiptHandle: msg.ReceiptHandle,
			}

			_, err := client.DeleteMessage(&del_params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to delete SQS message: %v\n", err)
			}
		}
	}
}

func main() {
	os.Exit(run())
}
