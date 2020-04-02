package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func process_command(cmd_name string, commands []Command) string {

	var cmd *Command = nil

	for _, c := range commands {
		if c.Name == cmd_name {
			cmd = &c
			break
		}
	}

	if cmd == nil {
		return fmt.Sprintf("[%s] unknown command", cmd_name)
	}

	resp_list := []string{}
	start := time.Now()

	for _, p := range cmd.Programs {

		var err error

		if cmd.Output {
			resp, err2 := exec.Command(p[0], p[1:]...).Output()
			err = err2
			r := strings.TrimSuffix(string(resp), "\n")
			resp_list = append(resp_list, r)
		} else {
			err = exec.Command(p[0], p[1:]...).Run()
		}

		if err != nil {
			joined := strings.Join(p, " ")
			return fmt.Sprintf("[%s] error: `%s` %v", cmd_name, joined, err)
		}
	}

	elapsed_sec := time.Since(start).Seconds()

	if len(resp_list) > 0 {
		joined := strings.Join(resp_list, "\n")
		return fmt.Sprintf("[%s] OK / %.1fs\n```\n%s\n```", cmd_name, elapsed_sec, joined)
	} else {
		return fmt.Sprintf("[%s] OK / %.1fs", cmd_name, elapsed_sec)
	}
}
