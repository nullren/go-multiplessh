package multiplessh

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func readline(r *bufio.Reader) (string, error) {
	bline, err := r.ReadBytes('\n')
	if err == nil {
		return string(bline), nil
	}
	return "", err
}

func loopout(host string, oc chan string, r *bufio.Reader) error {
	line, err := readline(r)
	if err != nil {
		return err
	}
	oc <- fmt.Sprintf("%s\t%s", host, line)
	return loopout(host, oc, r)
}

func gatheroutput(host string, oc chan string, cmd *exec.Cmd) error {
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	return loopout(host, oc, bufio.NewReader(out))
}

func run(host string, oc chan string, command ...string) *exec.Cmd {
	cmd := exec.Command("ssh", append([]string{"-tt", host}, command...)...)
	cmd.Stdin, _ = os.Open("/dev/null")
	go gatheroutput(host, oc, cmd)
	return cmd
}

func Run(hosts []string, command ...string) chan string {
	oc := make(chan string)

	for _, host := range hosts {
		run(host, oc, command...)
	}

	return oc
}
