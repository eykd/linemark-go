package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var lmkBinary string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "lmk-acceptance-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	lmkBinary = filepath.Join(tmpDir, "lmk")
	build := exec.Command("go", "build", "-o", lmkBinary, "github.com/eykd/linemark-go")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("failed to build lmk binary: " + err.Error())
	}

	os.Exit(m.Run())
}
