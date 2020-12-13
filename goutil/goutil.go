package goutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunUnitTests function
func RunUnitTests(gopath, pkg string, stdout, stderr io.Writer) error {

	gocmd := filepath.Join(string(os.PathSeparator), "usr", "local", "go", "bin", "go")

	goTestCmd := exec.Command(gocmd, "test", "-v", "-short", pkg)

	goTestCmd.Env = addGoPathENV(gopath)

	goTestCmd.Stdout = stdout
	goTestCmd.Stderr = stderr

	return goTestCmd.Run()
}

func addGoPathENV(gopath string) []string {

	currEnv := os.Environ()

	newEnv := make([]string, 0, len(currEnv))
	for _, env := range currEnv {
		if !strings.HasPrefix(env, "GOPATH=") {
			newEnv = append(newEnv, env)
		}
	}

	return append(newEnv, fmt.Sprintf("GOPATH=%s", gopath))
}
