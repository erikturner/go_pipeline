package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CommitInfo - struct reflecting commit info
type CommitInfo struct {
	Description string `json:"description,omitempty" example:"Fixed defect"`
	Commit      string `json:"commit,omitempty" example:"0f69a1de4e9f2af1d004bfbb756455a7f89d03b0"`
	Author      string `json:"author,omitempty" example:"eschw"`
	Date        string `json:"date,omitempty" example:"July 5, 2016 at 10:07:07 AM EDT"`
}

// IsRepo function
func IsRepo(dir string) bool {

	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return false
	}

	return true
}

// Clone function
func Clone(url, dir string) error {

	cloneCmd := exec.Command("git", "clone", url, dir)
	return cloneCmd.Run()
}

// Fetch function
func Fetch(dir string) error {

	fetchCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", dir), fmt.Sprintf("--git-dir=%s", filepath.Join(dir, ".git")), "fetch")
	return fetchCmd.Run()
}

// Pull function
func Pull(dir string) error {

	pullCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", dir), fmt.Sprintf("--git-dir=%s", filepath.Join(dir, ".git")), "pull")
	return pullCmd.Run()
}

// Reset function
func Reset(dir string) error {

	resetCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", dir), fmt.Sprintf("--git-dir=%s", filepath.Join(dir, ".git")), "reset", "--hard")
	return resetCmd.Run()
}

// Clean function
func Clean(dir string) error {

	cleanCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", dir), fmt.Sprintf("--git-dir=%s", filepath.Join(dir, ".git")), "clean", "-f", "-d")
	return cleanCmd.Run()
}

// Checkout function
func Checkout(dir, branch string) error {

	checkoutCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", dir), fmt.Sprintf("--git-dir=%s", filepath.Join(dir, ".git")), "checkout", branch)
	return checkoutCmd.Run()
}
