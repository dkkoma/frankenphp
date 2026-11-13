//go:build linux || darwin || freebsd

package frankenphp_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runCrashingTestCLI(t *testing.T, flag string) (*exec.ExitError, string) {
	t.Helper()

	if _, err := os.Stat("internal/testcli/testcli"); err != nil {
		t.Skip("internal/testcli/testcli has not been compiled, run `cd internal/testcli/ && go build`")
	}

	var stderr bytes.Buffer
	cmd := exec.Command("internal/testcli/testcli", flag)
	cmd.Stderr = &stderr
	err := cmd.Run()

	exitError, ok := errors.AsType[*exec.ExitError](err)
	require.True(t, ok, "the process must not exit normally: %v", err)

	return exitError, stderr.String()
}

func TestCrashReportWhenCThreadSegfaults(t *testing.T) {
	exitError, stderr := runCrashingTestCLI(t, "--crash-in-c-thread")

	status, ok := exitError.Sys().(syscall.WaitStatus)
	require.True(t, ok)
	assert.True(t, status.Signaled(), "the process must still die from the signal")
	assert.Equal(t, syscall.SIGSEGV, status.Signal())

	assert.Contains(t, stderr, "frankenphp: fatal signal 11 (SIGSEGV)")
	assert.Contains(t, stderr, " addr=0x0 ")
}

func TestGoTracebackIsKeptWhenGoThreadSegfaultsInC(t *testing.T) {
	exitError, stderr := runCrashingTestCLI(t, "--crash-in-go-thread")

	assert.Equal(t, 2, exitError.ExitCode())
	assert.Contains(t, stderr, "frankenphp: fatal signal 11 (SIGSEGV)")
	assert.Contains(t, stderr, "signal arrived during cgo execution")
	assert.Contains(t, stderr, "main.crashInGoThread")
}
