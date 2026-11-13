package cmd

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

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

func TestKeepChunksAliveSkipsLockedChunks(t *testing.T) {
	locked := memoryChunk{buf: make([]byte, 2*memoryPageStride), locked: true}
	unlocked := memoryChunk{buf: make([]byte, 2*memoryPageStride)}

	keepChunksAlive([]memoryChunk{locked, unlocked})

	if locked.buf[0] != 0 || locked.buf[memoryPageStride] != 0 {
		t.Error("locked chunk was touched; it is already pinned in RAM")
	}
	if unlocked.buf[0] != 1 || unlocked.buf[memoryPageStride] != 1 {
		t.Error("every page of an unlocked chunk should be touched")
	}
}

func TestMemoryKeepAliveIsCheap(t *testing.T) {
	if memoryKeepAliveInterval < time.Second {
		t.Fatalf("keep-alive interval %v rewrites every page too often", memoryKeepAliveInterval)
	}
}
