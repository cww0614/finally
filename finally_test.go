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

func expectTriggered(t *testing.T, output []byte, n int) {
	re := regexp.MustCompile("(?m)Triggered")
	triggered := len(re.FindAll(output, -1))
	if triggered != n {
		t.Fatal("Expect \"Triggered\" to be printed", n, "times", "got", triggered)
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

	expectTriggered(t, output.Bytes(), 1)
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

	expectTriggered(t, output.Bytes(), 1)
}

func TestNormalExit(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/simple.go")
	if err != nil {
		t.Fatal("Failed to launch test progam", err)
	}

	defer cleanup()

	cmd.Wait()

	expectTriggered(t, output.Bytes(), 1)
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

	expectTriggered(t, output.Bytes(), 1)
}

func TestMultiple(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/multiple.go")
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

	expectTriggered(t, output.Bytes(), 3)
}

func TestMultipleNormal(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/multiple.go")
	if err != nil {
		t.Fatal("Failed to launch test progam", err)
	}

	defer cleanup()

	cmd.Wait()

	expectTriggered(t, output.Bytes(), 3)
}

func TestOrder(t *testing.T) {
	cmd, output, cleanup, err := runTestProgram("tests/order.go")
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

	if output.String() != "Triggered3\nTriggered2\nTriggered1\n" {
		fmt.Printf("DEBUG: output = %+v\n", output)
		t.Fatal("The order is different from defer!")
	}
}
