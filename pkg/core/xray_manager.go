package core

import (
	"CryMeVPN/pkg/storage"
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var GitHubUrl = "https://api.github.com/repos/XTLS/Xray-core/releases/latest"

var fileGitXray = map[string]map[string]string{
	"windows":{
		"amd64": "windows-64.zip",
		"arm64": "windows-arm64-v8a.zip",
	},
	"linux": {
		"amd64": "linux-64.zip",
		"arm64": "linux-arm64-v8a.zip",
	},
	"darwin": {
		"amd64": "macos-64.zip",
		"arm64": "macos-arm64-v8a.zip",
	},
}

var XrayCMD *exec.Cmd

type GitHubRelease struct {
	Assets []GitHubAssets
}

type GitHubAssets struct {
	Name string `json:"name"`
	URL string `json:"browser_download_url"`
	Digest string `json:"digest"`

}

func Sha256Validation(path, digest string) error {
	want := strings.TrimPrefix(digest, "sha256:")
	if want == "" || want == digest {
		return fmt.Errorf("release has no sha256")
	}

	f,err := os.Open(path)

	if err != nil {
		return err
	}

	defer f.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash,f);err != nil {
		return err
	}

	got := hex.EncodeToString(hash.Sum(nil))

	if !strings.EqualFold(got, want) {
		return fmt.Errorf("fiele break sha256: got: %s - want: %s",got,want)
	}

	return nil
} 

func GetXrayPath() (string, error) {
	dir, err := storage.GetClearPath()

	if err != nil {
		return "",err
	}

	binDir := filepath.Join(dir,"bin")

	if err := os.MkdirAll(binDir,0755); err != nil {
		return "",err
	}

	nameBinFile := "xray"

	if runtime.GOOS == "windows" {nameBinFile = "xray.exe"}

	return filepath.Join(binDir,nameBinFile),nil
}

func GetLastXrayAsset() (GitHubAssets, error) {
	var httpClient = &http.Client{Timeout: 3 * time.Minute}

	respon,err := httpClient.Get(GitHubUrl)

	if err != nil {
		return GitHubAssets{},err
	}


	if respon.StatusCode != http.StatusOK {
		return GitHubAssets{},fmt.Errorf("github api died: %s",respon.Status)
	}

	defer respon.Body.Close()

	var releases GitHubRelease

	if err := json.NewDecoder(respon.Body).Decode(&releases); err != nil {
		return GitHubAssets{},err
	}

	suffixOs := fileGitXray[runtime.GOOS][runtime.GOARCH]

	for _,assets := range releases.Assets {
		if suffixOs != "" && strings.Contains(assets.Name, suffixOs) {
			return assets,nil
		}
	}


	return GitHubAssets{}, fmt.Errorf("suitable asset not found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func DownloadXray(url, path string) error {
	var httpClient = &http.Client{Timeout: 3 * time.Minute}

	respon,err := httpClient.Get(url)

	if err != nil {
		return err
	}

	defer respon.Body.Close()

	if respon.StatusCode != http.StatusOK {
		return errors.New("bad status respon")
	}

	out,err := os.Create(path)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, respon.Body)
	return err
}

func ExtractXray(zipPath,path string) error {
	reader, err := zip.OpenReader(zipPath)

	if err != nil {
		return err
	}

	defer reader.Close()

	for _, file := range reader.File {
		if filepath.Base(file.Name) == "xray" || filepath.Base(file.Name) == "xray.exe" {
			rc, err := file.Open()
			if err != nil {
				return err
			}

			defer rc.Close()

			destPath := filepath.Join(path, filepath.Base(file.Name))
			tempPath := destPath + ".tmp"

			out, err := os.Create(tempPath)

			if err != nil {
				return err
			}


			if _, err = io.Copy(out, rc); err != nil {
				out.Close()
				os.Remove(tempPath)
				return err
			}

			if err = out.Close();err != nil {
				os.Remove(tempPath)
				return err
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				_ = os.Chmod(tempPath,0755)
			}

			return os.Rename(tempPath,destPath)
		}
	}

	return errors.New("archive haven't xray binary.")
}