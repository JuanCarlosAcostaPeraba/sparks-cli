package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateDownloadsVerifiesAndAppliesRelease(t *testing.T) {
	binary := []byte("new sparks binary")
	archive := zipArchive(t, "sparks.exe", binary)
	archiveName := "sparks_0.2.0_windows_amd64.zip"
	digest := sha256.Sum256(archive)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_ = json.NewEncoder(w).Encode(release{
				TagName: "v0.2.0",
				Assets: []releaseAsset{
					{Name: archiveName, URL: serverURL(r) + "/archive"},
					{Name: "checksums.txt", URL: serverURL(r) + "/checksums"},
				},
			})
		case "/archive":
			_, _ = w.Write(archive)
		case "/checksums":
			_, _ = fmt.Fprintf(w, "%x  %s\n", digest, archiveName)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "sparks.exe")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	var applied []byte
	updater := New("0.1.0")
	updater.apiURL = server.URL + "/latest"
	updater.client = server.Client()
	updater.goos = "windows"
	updater.goarch = "amd64"
	updater.executablePath = target
	var stages []ProgressStage
	updater.OnProgress(func(stage ProgressStage) {
		stages = append(stages, stage)
	})
	updater.apply = func(reader io.Reader, gotTarget string, _ os.FileMode) error {
		if gotTarget != target {
			t.Fatalf("target = %q, want %q", gotTarget, target)
		}
		var err error
		applied, err = io.ReadAll(reader)
		return err
	}

	result, err := updater.Update(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Updated || result.CurrentVersion != "0.1.0" || result.LatestVersion != "0.2.0" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !bytes.Equal(applied, binary) {
		t.Fatalf("applied binary = %q, want %q", applied, binary)
	}
	wantStages := []ProgressStage{ProgressChecking, ProgressDownloading, ProgressVerifying, ProgressInstalling}
	if fmt.Sprint(stages) != fmt.Sprint(wantStages) {
		t.Fatalf("progress stages = %v, want %v", stages, wantStages)
	}
}

func TestDownloadRetriesBeforeUsingFallback(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		http.Error(w, "try again", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	updater := New("1.0.0")
	updater.client = server.Client()
	fallbackCalls := 0
	updater.fallback = func(_ context.Context, url string, limit int64) ([]byte, error) {
		fallbackCalls++
		if url != server.URL || limit != 16 {
			t.Fatalf("fallback called with url=%q limit=%d", url, limit)
		}
		return []byte("fallback"), nil
	}

	data, err := updater.download(context.Background(), server.URL, 16)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "fallback" || attempts != maxDownloadTries || fallbackCalls != 1 {
		t.Fatalf("data=%q attempts=%d fallbackCalls=%d", data, attempts, fallbackCalls)
	}
}

func TestDownloadDoesNotFallbackForPermanentHTTPError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	updater := New("1.0.0")
	updater.client = server.Client()
	updater.fallback = func(context.Context, string, int64) ([]byte, error) {
		t.Fatal("fallback should not be used for a permanent HTTP error")
		return nil, nil
	}

	_, err := updater.download(context.Background(), server.URL, 16)
	if err == nil || !strings.Contains(err.Error(), "404 Not Found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateSkipsCurrentOrNewerVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(release{TagName: "v0.2.0"})
	}))
	defer server.Close()

	updater := New("0.3.0")
	updater.apiURL = server.URL
	updater.client = server.Client()
	updater.apply = func(io.Reader, string, os.FileMode) error {
		t.Fatal("update should not be applied")
		return nil
	}

	result, err := updater.Update(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated || result.LatestVersion != "0.2.0" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestUpdateRejectsChecksumMismatch(t *testing.T) {
	archiveName := "sparks_0.2.0_linux_amd64.tar.gz"
	archive := tarGzipArchive(t, "sparks", []byte("binary"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_ = json.NewEncoder(w).Encode(release{
				TagName: "v0.2.0",
				Assets: []releaseAsset{
					{Name: archiveName, URL: serverURL(r) + "/archive"},
					{Name: "checksums.txt", URL: serverURL(r) + "/checksums"},
				},
			})
		case "/archive":
			_, _ = w.Write(archive)
		case "/checksums":
			_, _ = fmt.Fprintf(w, "%064d  %s\n", 0, archiveName)
		}
	}))
	defer server.Close()

	updater := New("0.1.0")
	updater.apiURL = server.URL + "/latest"
	updater.client = server.Client()
	updater.goos = "linux"
	updater.goarch = "amd64"
	updater.apply = func(io.Reader, string, os.FileMode) error {
		t.Fatal("update should not be applied")
		return nil
	}

	_, err := updater.Update(context.Background())
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}

func TestExtractBinaryFromTarGzip(t *testing.T) {
	want := []byte("unix binary")
	archive := tarGzipArchive(t, "folder/sparks", want)
	got, err := extractBinary("sparks_0.2.0_linux_amd64.tar.gz", "sparks", archive)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("binary = %q, want %q", got, want)
	}
}

func TestAssetNamesRejectUnsupportedPlatform(t *testing.T) {
	if _, _, err := assetNames("0.2.0", "windows", "arm64"); err == nil {
		t.Fatal("expected windows/arm64 error")
	}
	if _, _, err := assetNames("0.2.0", "plan9", "amd64"); err == nil {
		t.Fatal("expected unsupported OS error")
	}
}

func TestVersionComparison(t *testing.T) {
	stable, err := parseVersion("v1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	prerelease, err := parseVersion("1.2.3-next")
	if err != nil {
		t.Fatal(err)
	}
	older, err := parseVersion("1.2.2")
	if err != nil {
		t.Fatal(err)
	}
	if compareVersions(prerelease, stable) >= 0 || compareVersions(stable, older) <= 0 {
		t.Fatal("unexpected semantic version ordering")
	}
	if _, err := parseVersion("development"); err == nil {
		t.Fatal("expected invalid version error")
	}
}

func TestDefaultApplyReplacesTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "sparks-test")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	updater := New("0.1.0")
	if err := updater.apply(strings.NewReader("new"), target, 0o755); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new" {
		t.Fatalf("target contains %q", contents)
	}
}

func zipArchive(t *testing.T, name string, contents []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func tarGzipArchive(t *testing.T, name string, contents []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func serverURL(r *http.Request) string {
	return "http://" + r.Host
}
