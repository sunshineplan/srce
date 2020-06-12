package misc

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"
)

type result struct {
	err    error
	stdout []byte
	stderr []byte
}

// Run command and return result
func Run(cmd *exec.Cmd) (string, error) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	done := make(chan result)
	go func() {
		res := new(result)
		res.stdout, _ = ioutil.ReadAll(stdout)
		res.stderr, _ = ioutil.ReadAll(stderr)
		res.err = cmd.Wait()
		done <- *res
	}()
	select {
	case <-time.After(30 * time.Second):
		return "Process still running.", nil
	case r := <-done:
		return fmt.Sprintf("Output:\n%s\n\nError:\n%s", r.stdout, r.stderr), r.err
	}
}
