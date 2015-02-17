package snappy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// snapConfig configures a installed snap in the given directory
//
// It takes a rawConfig string that is passed as the new configuration
// This string can be empty.
//
// It returns the newConfig from or an error
func snapConfig(snapDir, rawConfig string) (newConfig string, err error) {
	configScript := filepath.Join(snapDir, "meta", "hooks", "config")
	if _, err := os.Stat(configScript); err != nil {
		return "", fmt.Errorf("No config for '%s'", snapDir)
	}

	part := NewInstalledSnapPart(filepath.Join(snapDir, "meta", "package.yaml"))
	if part == nil {
		return "", fmt.Errorf("No snap found in '%s'", snapDir)
	}

	return runConfigScript(configScript, rawConfig, makeSnapHookEnv(part))
}

// runConfigScript is a helper that just runs the config script and passes
// the rawConfig via stdin and reads/returns the output
func runConfigScript(configScript, rawConfig string, env []string) (newConfig string, err error) {
	cmd := exec.Command(configScript)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	cmd.Env = env

	// meh, really golang?
	go func() {
		defer stdin.Close()
		io.Copy(stdin, strings.NewReader(rawConfig))
	}()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("config failed with: '%s'", output)
	}

	return string(output), nil
}
