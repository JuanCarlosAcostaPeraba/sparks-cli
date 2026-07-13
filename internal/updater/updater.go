package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	selfupdate "github.com/minio/selfupdate"
)

const (
	defaultAPIURL    = "https://api.github.com/repos/JuanCarlosAcostaPeraba/sparks-cli/releases/latest"
	maxMetadataBytes = 5 << 20
	maxArchiveBytes  = 200 << 20
	maxBinaryBytes   = 100 << 20
)

type Result struct {
	CurrentVersion string
	LatestVersion  string
	Updated        bool
}

type applyFunc func(io.Reader, string, os.FileMode) error

type Updater struct {
	currentVersion string
	apiURL         string
	goos           string
	goarch         string
	executablePath string
	client         *http.Client
	apply          applyFunc
}

type release struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func New(currentVersion string) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		apiURL:         defaultAPIURL,
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
		client:         &http.Client{Timeout: 30 * time.Second},
		apply: func(binary io.Reader, target string, mode os.FileMode) error {
			return selfupdate.Apply(binary, selfupdate.Options{TargetPath: target, TargetMode: mode})
		},
	}
}

func (u *Updater) Update(ctx context.Context) (Result, error) {
	current, err := parseVersion(u.currentVersion)
	if err != nil {
		return Result{}, fmt.Errorf("read current version: %w", err)
	}

	release, err := u.latestRelease(ctx)
	if err != nil {
		return Result{}, err
	}
	latest, err := parseVersion(release.TagName)
	if err != nil {
		return Result{}, fmt.Errorf("read latest release version: %w", err)
	}

	result := Result{CurrentVersion: current.String(), LatestVersion: latest.String()}
	if compareVersions(current, latest) >= 0 {
		return result, nil
	}

	archiveName, binaryName, err := assetNames(latest.String(), u.goos, u.goarch)
	if err != nil {
		return Result{}, err
	}
	archiveURL, ok := findAsset(release.Assets, archiveName)
	if !ok {
		return Result{}, fmt.Errorf("release asset %q not found", archiveName)
	}
	checksumsURL, ok := findAsset(release.Assets, "checksums.txt")
	if !ok {
		return Result{}, errors.New("release asset \"checksums.txt\" not found")
	}

	archive, err := u.download(ctx, archiveURL, maxArchiveBytes)
	if err != nil {
		return Result{}, fmt.Errorf("download %s: %w", archiveName, err)
	}
	checksums, err := u.download(ctx, checksumsURL, maxMetadataBytes)
	if err != nil {
		return Result{}, fmt.Errorf("download checksums: %w", err)
	}
	if err := verifyChecksum(archiveName, archive, checksums); err != nil {
		return Result{}, err
	}

	binary, err := extractBinary(archiveName, binaryName, archive)
	if err != nil {
		return Result{}, err
	}
	target, mode, err := u.target()
	if err != nil {
		return Result{}, err
	}
	if err := u.apply(bytes.NewReader(binary), target, mode); err != nil {
		return Result{}, fmt.Errorf("replace executable: %w", err)
	}

	result.Updated = true
	return result, nil
}

func (u *Updater) latestRelease(ctx context.Context) (release, error) {
	data, err := u.download(ctx, u.apiURL, maxMetadataBytes)
	if err != nil {
		return release{}, fmt.Errorf("check latest release: %w", err)
	}
	var latest release
	if err := json.Unmarshal(data, &latest); err != nil {
		return release{}, fmt.Errorf("decode latest release: %w", err)
	}
	if strings.TrimSpace(latest.TagName) == "" {
		return release{}, errors.New("latest release has no tag")
	}
	return latest, nil
}

func (u *Updater) download(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "sparks-updater/"+u.currentVersion)

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds %d bytes", limit)
	}
	return data, nil
}

func (u *Updater) target() (string, os.FileMode, error) {
	target := u.executablePath
	if target == "" {
		var err error
		target, err = os.Executable()
		if err != nil {
			return "", 0, fmt.Errorf("locate executable: %w", err)
		}
		if resolved, resolveErr := filepath.EvalSymlinks(target); resolveErr == nil {
			target = resolved
		}
	}
	info, err := os.Stat(target)
	if err != nil {
		return "", 0, fmt.Errorf("inspect executable: %w", err)
	}
	return target, info.Mode(), nil
}

func findAsset(assets []releaseAsset, name string) (string, bool) {
	for _, asset := range assets {
		if asset.Name == name && asset.URL != "" {
			return asset.URL, true
		}
	}
	return "", false
}

func assetNames(version, goos, goarch string) (string, string, error) {
	if goarch != "amd64" && goarch != "arm64" {
		return "", "", fmt.Errorf("updates are not available for architecture %s", goarch)
	}
	if goos == "windows" {
		if goarch == "arm64" {
			return "", "", errors.New("updates are not available for windows/arm64")
		}
		return fmt.Sprintf("sparks_%s_windows_%s.zip", version, goarch), "sparks.exe", nil
	}
	if goos != "darwin" && goos != "linux" {
		return "", "", fmt.Errorf("updates are not available for %s/%s", goos, goarch)
	}
	return fmt.Sprintf("sparks_%s_%s_%s.tar.gz", version, goos, goarch), "sparks", nil
}

func verifyChecksum(name string, archive, checksums []byte) error {
	want := ""
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		filename := strings.TrimPrefix(fields[len(fields)-1], "*")
		if filename == name {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("checksum for %q not found", name)
	}
	expected, err := hex.DecodeString(want)
	if err != nil {
		return fmt.Errorf("decode checksum for %q: %w", name, err)
	}
	actual := sha256.Sum256(archive)
	if !bytes.Equal(expected, actual[:]) {
		return fmt.Errorf("checksum mismatch for %q", name)
	}
	return nil
}

func extractBinary(archiveName, binaryName string, archive []byte) ([]byte, error) {
	if strings.HasSuffix(archiveName, ".zip") {
		reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, fmt.Errorf("open update archive: %w", err)
		}
		for _, file := range reader.File {
			if !file.FileInfo().IsDir() && filepath.Base(file.Name) == binaryName {
				opened, err := file.Open()
				if err != nil {
					return nil, fmt.Errorf("open updated binary: %w", err)
				}
				defer opened.Close()
				return readBinary(opened)
			}
		}
		return nil, fmt.Errorf("binary %q not found in update archive", binaryName)
	}

	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open update archive: %w", err)
	}
	defer gz.Close()
	tarReader := tar.NewReader(gz)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read update archive: %w", err)
		}
		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == binaryName {
			return readBinary(tarReader)
		}
	}
	return nil, fmt.Errorf("binary %q not found in update archive", binaryName)
}

func readBinary(reader io.Reader) ([]byte, error) {
	binary, err := io.ReadAll(io.LimitReader(reader, maxBinaryBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read updated binary: %w", err)
	}
	if len(binary) > maxBinaryBytes {
		return nil, errors.New("updated binary is too large")
	}
	return binary, nil
}

type semVersion struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

func parseVersion(raw string) (semVersion, error) {
	normalized := strings.TrimPrefix(strings.TrimSpace(raw), "v")
	normalized = strings.SplitN(normalized, "+", 2)[0]
	parts := strings.SplitN(normalized, "-", 2)
	numbers := strings.Split(parts[0], ".")
	if len(numbers) != 3 {
		return semVersion{}, fmt.Errorf("invalid semantic version %q", raw)
	}
	values := make([]int, 3)
	for i, number := range numbers {
		value, err := strconv.Atoi(number)
		if err != nil || value < 0 {
			return semVersion{}, fmt.Errorf("invalid semantic version %q", raw)
		}
		values[i] = value
	}
	parsed := semVersion{major: values[0], minor: values[1], patch: values[2]}
	if len(parts) == 2 {
		if parts[1] == "" {
			return semVersion{}, fmt.Errorf("invalid semantic version %q", raw)
		}
		parsed.prerelease = parts[1]
	}
	return parsed, nil
}

func (v semVersion) String() string {
	value := fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	if v.prerelease != "" {
		value += "-" + v.prerelease
	}
	return value
}

func compareVersions(left, right semVersion) int {
	for _, pair := range [][2]int{{left.major, right.major}, {left.minor, right.minor}, {left.patch, right.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if left.prerelease == right.prerelease {
		return 0
	}
	if left.prerelease == "" {
		return 1
	}
	if right.prerelease == "" {
		return -1
	}
	return strings.Compare(left.prerelease, right.prerelease)
}
