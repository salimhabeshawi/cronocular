package background

import (
	"os"
	"os/exec"
)

func Start(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	childArgs := []string{"--headless"}
	for _, arg := range args {
		if arg == "--background" || arg == "-d" {
			continue
		}
		childArgs = append(childArgs, arg)
	}

	cmd := exec.Command(exe, childArgs...)
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer null.Close()

	cmd.Stdin = null
	cmd.Stdout = null
	cmd.Stderr = null
	prepareDetached(cmd)

	return cmd.Start()
}
