package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// sharedLine is a line as smartpireadout writes it: timestamp, 26 readings,
// trailing separator.
const sharedLine = "2024-01-01 12:00:00;1.5;2.5;3.5;0.1;230.1;230.2;230.3;345;575;805;0.95;0.96;0.97;" +
	"50.01;50.02;50.03;1.1;2.2;3.3;0.4;0.5;0.6;0.01;0.91;0.92;0.93;"

// writeShared rewrites path the way smartpiacFile.WriteSharedFile does:
// truncate in place, then write the line. It reports errors with t.Error so
// that it can run in other goroutines.
func writeShared(t testing.TB, path, line string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		t.Error(err)
	}
}

func TestReadRegisterValuesCompleteLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	writeShared(t, path, sharedLine)

	values, err := readRegisterValues(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 26 {
		t.Fatalf("got %d values, want 26", len(values))
	}
	if got := math.Float32frombits(values[4]); got != float32(230.1) {
		t.Errorf("V1 = %v, want 230.1", got)
	}
	if got := math.Float32frombits(values[25]); got != float32(0.93) {
		t.Errorf("PF3 = %v, want 0.93", got)
	}
}

func TestReadRegisterValuesIncompleteFile(t *testing.T) {
	fields := strings.Split(sharedLine, ";")
	cases := map[string]string{
		"empty":           "",
		"timestamp only":  "2026-11-07 15:21:18;",
		"ten fields":      strings.Join(fields[:10], ";"),
		"cut in a number": sharedLine[:len(sharedLine)-4],
		"cut after field": strings.Join(fields[:22], ";"),
		"not a number":    strings.Replace(sharedLine, ";230.1;", ";x;", 1),
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "smartpi_values")
			writeShared(t, path, content)
			if values, err := readRegisterValues(path); err == nil {
				t.Fatalf("got %d values and no error for %q", len(values), content)
			}
		})
	}
}

func TestReadRegisterValuesMissingFile(t *testing.T) {
	if _, err := readRegisterValues(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("no error for a missing file")
	}
}

// TestReadRegisterValuesWhileRewritten reads the file while it is rewritten
// continuously. Every read must return either an error or the complete line.
func TestReadRegisterValuesWhileRewritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	writeShared(t, path, sharedLine)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				writeShared(t, path, sharedLine)
			}
		}
	}()

	complete, failed := 0, 0
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		values, err := readRegisterValues(path)
		if err != nil {
			failed++
			continue
		}
		if len(values) != 26 {
			t.Fatalf("got %d values without an error", len(values))
		}
		complete++
	}
	close(stop)
	wg.Wait()

	if complete == 0 {
		t.Fatal("no complete read at all")
	}
	t.Logf("%d complete reads, %d incomplete ones skipped", complete, failed)
}

// TestReadRegisterValuesRetryRecovers starts with an empty file, as a read
// right after the truncate finds it, and writes the line while the reader is
// still retrying.
func TestReadRegisterValuesRetryRecovers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	writeShared(t, path, "")

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(20 * time.Millisecond)
		writeShared(t, path, sharedLine)
	}()

	values, err := readRegisterValuesRetry(path, 50, 10*time.Millisecond)
	<-done
	if err != nil {
		t.Fatalf("line was written during the retries, got %v", err)
	}
	if len(values) != 26 {
		t.Fatalf("got %d values, want 26", len(values))
	}
}

func TestReadRegisterValuesRetryGivesUp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	writeShared(t, path, "")

	start := time.Now()
	if _, err := readRegisterValuesRetry(path, 3, 10*time.Millisecond); err == nil {
		t.Fatal("no error for a file that stays empty")
	}
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Fatalf("gave up after %v, before the third attempt", elapsed)
	}
}

func TestLoadRegisters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	writeShared(t, path, sharedLine)
	values, err := readRegisterValues(path)
	if err != nil {
		t.Fatal(err)
	}

	registers := make([]uint16, 2*len(values)+2)
	if err := loadRegisters(registers, path); err != nil {
		t.Fatal(err)
	}
	for i, want := range values {
		if got := uint32(registers[2*i])<<16 | uint32(registers[2*i+1]); got != want {
			t.Errorf("value %d in registers %d and %d = %v, want %v", i, 2*i, 2*i+1,
				math.Float32frombits(got), math.Float32frombits(want))
		}
	}
	if tail := registers[2*len(values):]; tail[0] != 0 || tail[1] != 0 {
		t.Errorf("registers after the last value were written: %v", tail)
	}
}

func TestLoadRegistersKeepsRegistersOnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	writeShared(t, path, "")

	registers := []uint16{1, 2, 3, 4}
	if err := loadRegisters(registers, path); err == nil {
		t.Fatal("no error for an empty file")
	}
	for i, want := range []uint16{1, 2, 3, 4} {
		if registers[i] != want {
			t.Fatalf("registers = %v after a failed read, want them unchanged", registers)
		}
	}
}

func TestWaitForFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi_values")
	go func() {
		time.Sleep(700 * time.Millisecond)
		writeShared(t, path, sharedLine)
	}()
	if err := waitForFile(path, 5*time.Second); err != nil {
		t.Fatalf("file appeared but waitForFile returned %v", err)
	}
}

func TestWaitForFileTimeout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "never")
	start := time.Now()
	if err := waitForFile(path, time.Second); !os.IsNotExist(err) {
		t.Fatalf("got %v, want a not-exist error", err)
	}
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Fatalf("gave up after %v, before the timeout", elapsed)
	}
}
