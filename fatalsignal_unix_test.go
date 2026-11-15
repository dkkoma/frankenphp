//go:build unix && !asan && !msan

package frankenphp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stackOverflowScriptEnv = "FRANKENPHP_TEST_STACK_OVERFLOW_SCRIPT"

func TestGoReportsFatalSignalOnPHPThread(t *testing.T) {
	if script := os.Getenv(stackOverflowScriptEnv); script != "" {
		overflowPHPThreadStack(t, script)

		return
	}

	for _, script := range []string{"stack-overflow.php", "stack-overflow-worker.php"} {
		t.Run(script, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestGoReportsFatalSignalOnPHPThread$")
			cmd.Env = append(os.Environ(), stackOverflowScriptEnv+"="+script, "GOTRACEBACK=single")
			output, err := cmd.CombinedOutput()

			exitErr, ok := errors.AsType[*exec.ExitError](err)
			require.True(t, ok, "the process survived: %s", output)
			assert.Equal(t, 2, exitErr.ExitCode(), "%s", output)
			assert.Contains(t, string(output), "signal arrived during cgo execution")
			assert.Contains(t, string(output), "go_frankenphp_run_php_thread")
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
