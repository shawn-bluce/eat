package cmd

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestLockFailureWarnsOnce(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	lockWarned.Store(false)

	for range 300 {
		warnLockFailure(errors.New("cannot allocate memory"))
	}

	if n := strings.Count(buf.String(), "\n"); n != 1 {
		t.Fatalf("got %d warning lines, want 1:\n%s", n, buf.String())
	}
	if !strings.Contains(buf.String(), "cannot allocate memory") {
		t.Errorf("warning should include the cause, got %q", buf.String())
	}
}
