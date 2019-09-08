package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func execShell(s string) (res string, errMsg string) {
	defer func() {
		if p := recover(); p != nil {
			errMsg = fmt.Sprint(p)
		}
	}()

	cmd := exec.Command("/bin/bash", "-c", s)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		panic(fmt.Sprintf(
			"%s: %s", err, stderr.String()))
	}

	return stdout.String(), ""
}
