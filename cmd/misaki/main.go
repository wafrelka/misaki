package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"encoding/json"
	"github.com/burntsushi/toml"
	"misaki/pkg"
)

const (
	INITIAL_BACKOFF = 1
	MAXIMUM_BACKOFF = 300
	LISTENING_ADDRESS = "0.0.0.0:8080"
	CMD_SERVE = "serve"
	CMD_EXEC = "exec"
)

type MisakiMessage struct {
	CommandName string `json:"command_name"`
}

type AwsConfig struct {
	SecretAccessKey string `toml:"secret_access_key"`
	AccessKeyId string `toml:"access_key_id"`
	Region string `toml:"region"`
	QueueName string `toml:"queue"`
}

type Config struct {
	SlackWebhookUrl string `toml:"slack_webhook_url"`
	Aws AwsConfig `toml:"aws"`
	Commands []misaki.Command `toml:"commands"`
}

func print_error(text string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, text, values...)
}
func print_log(text string, values ...interface{}) {
	fmt.Fprintf(os.Stdout, text, values...)
}

func run_executor(q *misaki.SQSQueue, m *misaki.Mediator, s *misaki.Slack, commands []misaki.Command) int {

	m.Reset()

	for {

		sqs_msg, err := q.WaitMessage()
		if err != nil {
			print_error("cannot fetch SQS message: %v\n", err)
			print_error("wait for %d seconds\n", m.GetCurrent())
			m.Wait()
			m.Increment()
			continue
		}
		m.Reset()

		err = sqs_msg.Delete()
		if err != nil {
			print_error("failed to delete SQS message: %v\n", err)
			// continue to process the request
		}

		var m_msg MisakiMessage
		err = json.Unmarshal([]byte(sqs_msg.Content), &m_msg)
		if err != nil {
			print_error("invalid message: %v\n", err)
			continue
		}

		cmd_name := m_msg.CommandName
		print_log("command: %s\n", cmd_name)
		output := misaki.ProcessCommand(cmd_name, commands)
		err = s.Post(output)
		if err != nil {
			print_error("failed to post: %v\n", err)
		}
	}
}

func run_server(q *misaki.SQSQueue, m *misaki.Mediator, commands []misaki.Command) int {

	handler := misaki.NewMisakiHandler(func(cmd_name string) (string, int) {
		print_log("command: %s\n", cmd_name)
		m_msg := MisakiMessage {
			CommandName: cmd_name,
		}
		encoded, _ := json.Marshal(m_msg)
		err := q.PostMessage(string(encoded))
		if err != nil {
			return fmt.Sprintf("internal server error: %v", err), 500
		}
		return "OK", 200
	}, commands)

	server := &http.Server{
		Handler: handler,
	}

	listener, err := net.Listen("tcp", LISTENING_ADDRESS)
	if err != nil {
		print_error("cannot listen on %s: %v\n", LISTENING_ADDRESS, err)
		return 1
	}
	err = server.Serve(listener)
	if err != nil {
		print_error("cannot start server: %v\n", err)
		return 1
	}

	return 0
}

func run() int {

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s CMD CONFIG_FILE\n", os.Args[0])
		return 1
	}

	cmd := os.Args[1]
	config_file_path := os.Args[2]
	var config Config

	_, err := toml.DecodeFile(config_file_path, &config)
	if err != nil {
		print_error("cannot read config file: %v\n", err)
		return 1
	}

	if config.Aws.QueueName == "" {
		print_error("config error: empty SQS name\n")
		return 1
	}
	if cmd == CMD_EXEC && config.SlackWebhookUrl == "" {
		print_error("config error: empty Slack webhook url\n")
		return 1
	}

	q, err := misaki.NewSQSQueue(
		config.Aws.QueueName,
		config.Aws.Region,
		config.Aws.AccessKeyId,
		config.Aws.SecretAccessKey,
	)
	m := misaki.NewMediator(INITIAL_BACKOFF, MAXIMUM_BACKOFF)

	if err != nil {
		print_error("failed to initialize SQS queue: %v\n", err)
		return 1
	}

	if cmd == CMD_EXEC {
		s := misaki.NewSlack(config.SlackWebhookUrl)
		return run_executor(q, m, s, config.Commands)
	} else if cmd == CMD_SERVE {
		return run_server(q, m, config.Commands)
	} else {
		print_error("unknown command: %s\n", cmd)
		return 1
	}
}

func main() {
	os.Exit(run())
}
