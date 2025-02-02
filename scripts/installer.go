package scripts

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// InstallScript() downloads a .lua file from the given URL and saves it to scriptsDir.
func InstallScript(scriptUrl, scriptsDir string) error {
	urlParsed, err := url.Parse(scriptUrl)
	if err != nil {
		return fmt.Errorf("could not parse url: %s", err)
	}
	urlPath, _ := url.QueryUnescape(urlParsed.EscapedPath())

	fileName := filepath.Base(urlPath)
	if fileName == "" || fileName == "." {
		return fmt.Errorf("could not derive file name from url: %s", scriptUrl)
	}

	if !strings.HasSuffix(fileName, ".lua") {
		return fmt.Errorf(
			"refusing to install non-.lua file: %s",
			scriptUrl,
		)
	}

	resp, err := http.Get(scriptUrl)
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

	fmt.Printf("Installed script: %s -> %s\n", scriptUrl, localPath)
	return nil
}
