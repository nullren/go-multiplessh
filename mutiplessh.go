package multiplessh

import (
	"bufio"
	"os"
	"os/exec"
	"syscall"
)

func run(host string, command ...string) *exec.Cmd {
	cmd := exec.Command("ssh", append([]string{"-tt", host}, command...)...)
	// not totally certain about this one
	cmd.SysProcAtter = &syscall.SysProcAttr{Setpgid: true}
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
	if line, err := readline(r); err != nil {
		return err
	}
	// async send it
	go func() {
		c <- fmt.Sprintf("%s\t%s", host, line)
	}()
	return loopout(host, r, c)
}

func readline(r *bufio.Reader) (string, error) {
	bline, err := r.ReadBytes('\n')
	if err == nil {
		return string(bline), nil
	}
	return nil, err
}

// what to take a list of hosts, a command, and return a channel
func Run(hosts []string, command ...string) chan string {
	output := make(chan string)
	cmds := []*exec.Cmd{}

	for host := range hosts {
		cmd := run(host, command...)
		cmds = append(cmds, cmd)
		go gatherOutput(host, cmd, output)
	}

	return output
}