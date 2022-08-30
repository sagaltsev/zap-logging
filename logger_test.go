package logging

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSet(t *testing.T) {
	Set("info")

	assert.IsType(t, &zap.SugaredLogger{}, Logger)
}

func TestLogAppStart(t *testing.T) {
	testArgs := struct {
		Arg1 string
		Arg2 string
	}{
		Arg1: "foo",
		Arg2: "bar",
	}

	output := captureOutput(func() {
		LogAppStart("starting test", testArgs)
	})

	assert.Contains(t, output, `starting test`)
	assert.Contains(t, output, `"type": "lifecycle"`)
	assert.Contains(t, output, `"event": "start"`)
	assert.Contains(t, output, `"Arg1": "foo"`)
	assert.Contains(t, output, `"Arg2": "bar"`)
}

func TestLogAppStop_NoError(t *testing.T) {
	output := captureOutput(func() {
		LogAppStop("stoping test", os.Interrupt, nil)
	})

	assert.Contains(t, output, `stoping test`)
	assert.Contains(t, output, `interrupt`)
	assert.Contains(t, output, `"type": "lifecycle"`)
	assert.Contains(t, output, `"event": "stop"`)
}

func TestLogAppStop_WithError(t *testing.T) {
	output := captureOutput(func() {
		LogAppStop("stoping test", os.Interrupt, errors.New("serverErr"))
	})

	assert.Contains(t, output, `stoping test`)
	assert.NotContains(t, output, `interrupt`)
	assert.Contains(t, output, `"type": "lifecycle"`)
	assert.Contains(t, output, `"event": "stop"`)
	assert.Contains(t, output, `serverErr`)
}

func TestLogRequest(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)
	r.RequestURI = "/test/path"
	r.RemoteAddr = "127.0.0.1"

	output := captureOutput(func() {
		LogRequest(r, http.StatusOK)
	})

	assert.Contains(t, output, `200 ->GET /test/path`)
	assert.Contains(t, output, `"type": "access"`)
	assert.Contains(t, output, `"event": "request"`)
	assert.Contains(t, output, `"remote_ip": "127.0.0.1"`)
	assert.Contains(t, output, `"host": "testurl.com"`)
	assert.Contains(t, output, `"url_path": "/test/path"`)
	assert.Contains(t, output, `"method": "GET"`)
	assert.Contains(t, output, `"status_code": 200`)
}

func TestLogWithCorrelationIDS(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		msg      string
		id       string
		userID   string
		lvl      string
	}{
		{
			"info level log test",
			Info,
			"test info message",
			"TEST-CORRELATION-ID",
			"TEST-USER-CORRELATION-ID",
			"INFO",
		},
		{
			"debug level log test",
			Debug,
			"test debug message",
			"TEST-CORRELATION-ID",
			"TEST-USER-CORRELATION-ID",
			"DEBUG",
		},
		{
			"error level log test",
			Error,
			"test error message",
			"TEST-CORRELATION-ID",
			"TEST-USER-CORRELATION-ID",
			"ERROR",
		},
		{
			"warn level log test",
			Warn,
			"test warn message",
			"TEST-CORRELATION-ID",
			"TEST-USER-CORRELATION-ID",
			"WARN",
		},
		{
			"panic level log test",
			Panic,
			"test panic message",
			"TEST-CORRELATION-ID",
			"TEST-USER-CORRELATION-ID",
			"PANIC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			out := captureOutput(func() {
				if tt.logLevel == Panic {
					defer func() {
						if r := recover(); r != nil {
							fmt.Println("Recovered in f", r)
						}
					}()
				}
				LogWithCorrelationIDS(tt.logLevel, tt.msg, tt.id, tt.userID)
			})

			assert.Contains(t, out, tt.lvl)
			assert.Contains(t, out, tt.msg)
			assert.Contains(t, out, fmt.Sprintf(`"correlation_id": "%s"`, tt.id))
			assert.Contains(t, out, fmt.Sprintf(`"user_correlation_id": "%s"`, tt.userID))

		})
	}
}

func TestLogWithCorrelationIDS_Fatal(t *testing.T) {
	tt := struct {
		logLevel string
		msg      string
		id       string
		userID   string
		lvl      string
	}{
		Fatal,
		"test fatal message",
		"TEST-CORRELATION-ID",
		"TEST-USER-CORRELATION-ID",
		"FATAL",
	}

	if os.Getenv("BE_CRASHER") == "1" {
		LogWithCorrelationIDS(tt.logLevel, tt.msg, tt.id, tt.userID)
		return
	}

	// Start the actual test in a different subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestLogWithCorrelationIDS_Fatal")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	stdout, errPipe := cmd.StdoutPipe()
	if errPipe != nil {
		t.Fatal(errPipe)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Check that the log fatal message is what we expected
	gotBytes, _ := ioutil.ReadAll(stdout)
	got := string(gotBytes)
	assert.Contains(t, got, tt.msg)

	// Check that the program exited
	err := cmd.Wait()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Process ran with err %v, want exit status 1", err)
	}
}

func TestGetRemoteIP_CaseOne(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)
	r.Header["X-Cluster-Client-Ip"] = []string{"1.2.3.4"}

	remoteIP := getRemoteIP(r)

	assert.Equal(t, "1.2.3.4", remoteIP)
}

func TestGetRemoteIP_CaseTwo(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)
	r.Header["X-Real-Ip"] = []string{"1.2.3.4"}

	remoteIP := getRemoteIP(r)

	assert.Equal(t, "1.2.3.4", remoteIP)
}

func TestGetRemoteIP_CaseThree(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://testurl.com", nil)
	r.RemoteAddr = "1.2.3.4:8080"

	remoteIP := getRemoteIP(r)

	assert.Equal(t, "1.2.3.4", remoteIP)
}

func captureOutput(funcToRun func()) string {
	var buffer bytes.Buffer

	oldLogger := logger

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buffer)

	logger = zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))

	funcToRun()
	writer.Flush()

	logger = oldLogger

	return buffer.String()
}
