//go:build (linux || darwin || freebsd) && !asan && !msan

package frankenphp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stackOverflowScriptEnv = "FRANKENPHP_TEST_STACK_OVERFLOW_SCRIPT"

func runUntilFatalSignal(t *testing.T, cmd *exec.Cmd) (syscall.Signal, string) {
	t.Helper()

	output, err := cmd.CombinedOutput()
	exitError, ok := errors.AsType[*exec.ExitError](err)
	require.True(t, ok, "the process must not exit normally: %s", output)

	status, ok := exitError.Sys().(syscall.WaitStatus)
	require.True(t, ok)
	require.True(t, status.Signaled(), "the process must still die from the signal: %s", output)

	return status.Signal(), string(output)
}

func runCrashingTestCLI(t *testing.T, flag string) (syscall.Signal, string) {
	t.Helper()

	if _, err := os.Stat("internal/testcli/testcli"); err != nil {
		t.Skip("internal/testcli/testcli has not been compiled, run `cd internal/testcli/ && go build`")
	}

	return runUntilFatalSignal(t, exec.Command("internal/testcli/testcli", flag))
}

func TestCrashReportWhenPHPThreadOverflowsItsStack(t *testing.T) {
	if script := os.Getenv(stackOverflowScriptEnv); script != "" {
		overflowPHPThreadStack(t, script)

		return
	}

	for _, script := range []string{"stack-overflow.php", "stack-overflow-worker.php"} {
		t.Run(script, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCrashReportWhenPHPThreadOverflowsItsStack$")
			cmd.Env = append(os.Environ(), stackOverflowScriptEnv+"="+script)

			sig, output := runUntilFatalSignal(t, cmd)
			assert.Contains(t, []syscall.Signal{syscall.SIGSEGV, syscall.SIGBUS}, sig)
			assert.Regexp(t, `frankenphp: fatal signal \d+ \(SIG(SEGV|BUS)\) code=\d+ addr=0x[0-9a-f]+ (pc=0x[0-9a-f]+ )?thread=php-`, output)
		})
	}
}

func overflowPHPThreadStack(t *testing.T, script string) {
	opts := &testOptions{
		nbParallelRequests: 1,
		phpIni:             map[string]string{"zend.max_allowed_stack_size": "-1", "memory_limit": "-1"},
	}
	if strings.Contains(script, "worker") {
		opts.workerScript = script
		opts.nbWorkers = 1
	}

	runTest(t, func(handler func(http.ResponseWriter, *http.Request), _ *httptest.Server, _ int) {
		testGet("http://example.com/"+script, handler, t)
	}, opts)

	t.Fatal("the PHP thread survived the stack overflow")
}

func TestCrashReportWhenCThreadSegfaults(t *testing.T) {
	sig, output := runCrashingTestCLI(t, "--crash-in-c-thread")

	assert.Equal(t, syscall.SIGSEGV, sig)
	assert.Contains(t, output, "frankenphp: fatal signal 11 (SIGSEGV)")
	assert.Contains(t, output, " addr=0x0 ")
}

func TestCrashReportWhenCThreadAborts(t *testing.T) {
	sig, output := runCrashingTestCLI(t, "--abort-in-c-thread")

	assert.Equal(t, syscall.SIGABRT, sig)
	assert.Contains(t, output, "frankenphp: fatal signal 6 (SIGABRT)")
}

func TestCrashReportWhenGoThreadSegfaultsInC(t *testing.T) {
	sig, output := runCrashingTestCLI(t, "--crash-in-go-thread")

	assert.Equal(t, syscall.SIGSEGV, sig)
	assert.Contains(t, output, "frankenphp: fatal signal 11 (SIGSEGV)")
	assert.Contains(t, output, " addr=0x0 ")
}
