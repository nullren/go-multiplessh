package main

import (
	"github.com/nullren/go-multiplessh"
)

func main() {
	hosts := []string{"earth", "gemini"}
	oc := multiplessh.Run(hosts, "ps", "-ef")

	for line := range oc {
		print(line)
	}
}
