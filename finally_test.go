package finally_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/chrisww/finally"
)

func Example() {
	finally.RegisterShutdownHook()

	cleanup := finally.Wrap(func() {
		fmt.Println("Cleaning up")
	})

	defer cleanup()

	fmt.Println("Doing something")
}

func runTestProgram(name string) (*exec.Cmd, *bytes.Buffer, func(), error) {
	buildOutput, err := exec.Command("go", "build", name).CombinedOutput()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Build failed, output: %s", string(buildOutput))
	}

	lastDot := strings.LastIndex(name, ".")
	executableName := "./" + filepath.Base(name[:lastDot])

	cleanup := func() {
		os.Remove(executableName)
	}

	outputBuffer := new(bytes.Buffer)
	cmd := exec.Command(executableName)
	cmd.Stdout = outputBuffer

	err = cmd.Start()
	return cmd, outputBuffer, cleanup, err
}

func expectTriggeredOnce(t *testing.T, output []byte) {
	re := regexp.MustCompile("(?m)Triggered")
	triggered := len(re.FindAll(output, -1))
	if triggered != 1 {
		t.Fatal("Expect \"Triggered\" to be printed exactly once, got", triggered)
	}
}

func TestInterrupt(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/simple.go")
	if err != nil {
		t.Fatal("Failed to launch test progam", err)
	}

	defer cleanup()

	time.Sleep(1 * time.Second)

	err = cmd.Process.Signal(os.Interrupt)
	if err != nil {
		t.Fatal("Failed to send signal")
	}

	cmd.Wait()

	expectTriggeredOnce(t, output.Bytes())
}

func TestSIGTERM(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/simple.go")
	if err != nil {
		t.Fatal("Failed to launch test progam", err)
	}

	defer cleanup()

	time.Sleep(1 * time.Second)

	err = cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		t.Fatal("Failed to send signal")
	}

	cmd.Wait()

	expectTriggeredOnce(t, output.Bytes())
}

func TestNormalExit(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/simple.go")
	if err != nil {
		t.Fatal("Failed to launch test progam", err)
	}

	defer cleanup()

	cmd.Wait()

	expectTriggeredOnce(t, output.Bytes())
}

func TestPanic(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/panic.go")
	if err != nil {
		t.Fatal("Failed to launch test progam", err)
	}

	defer cleanup()

	time.Sleep(1 * time.Second)

	err = cmd.Process.Signal(os.Interrupt)
	if err != nil {
		t.Fatal("Failed to send signal")
	}

	cmd.Wait()

	expectTriggeredOnce(t, output.Bytes())
}
