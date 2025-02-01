package scripts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// InstallScript() downloads a .lua file from the given URL and saves it to scriptsDir.
func InstallScript(url, scriptsDir string) error {
	if !strings.HasSuffix(url, ".lua") {
		// require .lua extension
		return fmt.Errorf("refusing to install non-.lua file: %s", url)
	}

	fileName := filepath.Base(url)
	if fileName == "" {
		return fmt.Errorf("could not derive filename from url: %s", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch script. status: %d", resp.StatusCode)
	}

	localPath := filepath.Join(scriptsDir, fileName)
	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write script data: %w", err)
	}

	fmt.Printf("Installed script: %s -> %s\n", url, localPath)
	return nil
}
