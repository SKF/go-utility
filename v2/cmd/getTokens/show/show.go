package show

import (
	"bytes"
	"os"
	"os/exec"
)

func Show(s string) {
	cmd := exec.Command("less")
	cmd.Stdin = bytes.NewReader([]byte(s))

	cmd.Stdout = os.Stdout

	cmd.Run()
}
