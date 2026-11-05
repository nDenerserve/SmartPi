package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// uploadRequest builds a multipart request like the web UI's Update tab
// sends it, with an unrelated field before the file to make sure the
// reader skips it.
func uploadRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("note", "ignored"); err != nil {
		t.Fatal(err)
	}
	if filename != "" {
		fw, err := mw.CreateFormFile("file", filename)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write(content)
	}
	mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/update/package", &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	return r
}

func TestReceiveUpload_StreamsToStagingWithoutTempFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp) // os.TempDir(), where ParseMultipartForm used to spill
	staging := filepath.Join(t.TempDir(), "uploads")

	// Bigger than ParseMultipartForm's old 32 MB memory limit would not be
	// practical in a unit test; what matters is that nothing goes through
	// os.TempDir() at all, whatever the size.
	content := bytes.Repeat([]byte("deb!"), 1<<18) // 1 MB
	path, status, msg := receiveUpload(httptest.NewRecorder(), uploadRequest(t, "smartpi_1.0_arm64.deb", content), staging)
	if status != 0 {
		t.Fatalf("status %d: %s", status, msg)
	}
	if !strings.HasPrefix(path, staging) {
		t.Fatalf("stored at %s, not in staging dir %s", path, staging)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("stored file differs (err %v, %d bytes)", err, len(got))
	}
	if left, _ := os.ReadDir(tmp); len(left) != 0 {
		t.Fatalf("temporary files left in TMPDIR: %v", left)
	}
}

func TestReceiveUpload_Rejects(t *testing.T) {
	staging := t.TempDir()
	cases := []struct {
		name string
		req  *http.Request
		want int
	}{
		{"not a .deb", uploadRequest(t, "evil.sh", []byte("#!/bin/sh")), http.StatusBadRequest},
		{"no file field", uploadRequest(t, "", nil), http.StatusBadRequest},
		{"not multipart", httptest.NewRequest(http.MethodPost, "/", strings.NewReader("x")), http.StatusBadRequest},
	}
	for _, c := range cases {
		_, status, msg := receiveUpload(httptest.NewRecorder(), c.req, staging)
		if status != c.want {
			t.Errorf("%s: status %d (%s), want %d", c.name, status, msg, c.want)
		}
	}
	if left, _ := os.ReadDir(staging); len(left) != 0 {
		t.Fatalf("rejected uploads left files behind: %v", left)
	}
}
