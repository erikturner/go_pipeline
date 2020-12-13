package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"awesomeProject/goutil"
)

func runUnitTests(wo *workOrder) error {

	/*-----------------------------------------------------------------
	|
	| Loop over the all directory in the src dir
	|
	*-----------------------------------------------------------------*/
	return unitTestWalk(wo, wo.srcDir)
}

func unitTestWalk(wo *workOrder, startingDir string) error {

	return filepath.Walk(startingDir, func(path string, info os.FileInfo, err error) error {

		// Only run tests in directories
		if !info.IsDir() {
			return nil
		}

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		hasTestFiles := false
		for _, f := range files {
			if strings.HasSuffix(f.Name(), "_test.go") {
				hasTestFiles = true
				break
			}
		}

		// Use buffer so all the test output can be sent at once. This prevents it from getting
		// intermix with other goroutine's output
		if hasTestFiles {
			buff := &bytes.Buffer{}
			pkgToTest := strings.TrimPrefix(path, wo.srcDir)
			pkgToTest = strings.TrimPrefix(pkgToTest, string(os.PathSeparator))
			err := goutil.RunUnitTests(wo.workspaceDir, pkgToTest, buff, buff)
			respond(wo.w, fmt.Sprintf("\n%s", buff.String()))
			if err != nil {
				return err
			}
		}

		return nil
	})
}
