package multiplessh

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func run(host string, command ...string) *exec.Cmd {
	cmd := exec.Command("ssh", append([]string{"-tt", host}, command...)...)
	// not totally certain about this one
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdin, _ = os.Open("/dev/null")
	return cmd
}

func killpg(cmd *exec.Cmd) {
	if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
		syscall.Kill(-pgid, 1)
	}
}

func gatherOutput(host string, cmd *exec.Cmd, c chan string) error {
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	return loopout(host, bufio.NewReader(out), c)
}

func loopout(host string, r *bufio.Reader, c chan string) error {
	line, err := readline(r)
	if err != nil {
		return err
	}
	// async send it
	go func(host, line string) {
		c <- fmt.Sprintf("%s\t%s", host, line)
	}(host, line)
	return loopout(host, r, c)
}

func readline(r *bufio.Reader) (string, error) {
	bline, err := r.ReadBytes('\n')
	if err == nil {
		return string(bline), nil
	}
	return "", err
}

// what to take a list of hosts, a command, and return a channel
func Run(hosts []string, command ...string) chan string {
	output := make(chan string)
	cmds := []*exec.Cmd{}

	for _, host := range hosts {
		cmd := run(host, command...)
		cmds = append(cmds, cmd)
		go gatherOutput(host, cmd, output)
	}

	return output
}
