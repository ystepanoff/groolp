package scripts_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ystepanoff/groolp/scripts"
)

func TestInstallScript_RefusesNonLuaFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := scripts.LuaInstaller.InstallScript(
		"https://example.com/script.txt",
		tmpDir,
	)
	require.Error(t, err, "expected error for non-.lua file")
	require.Contains(t, err.Error(), "refusing to install non-.lua file")
}

func TestInstallScript_EmptyFileName(t *testing.T) {
	tmpDir := t.TempDir()

	err := scripts.LuaInstaller.InstallScript("https://example.com", tmpDir)
	require.Error(t, err, "expected error when filename cannot be derived")
	require.Contains(t, err.Error(), "could not derive file name")
}

func TestInstallScript_HttpError(t *testing.T) {
	tmpDir := t.TempDir()

	err := scripts.LuaInstaller.InstallScript(
		"http://127.0.0.1:9999/failing.lua",
		tmpDir,
	)
	require.Error(
		t,
		err,
		"expected error on connection refused or invalid endpoint",
	)
	require.Contains(t, err.Error(), "failed to download script")
}

func TestInstallScript_NonOkResponse(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test server returning 404
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Not found", http.StatusNotFound)
		}),
	)
	defer ts.Close()

	testURL := ts.URL + "/test.lua"
	err := scripts.LuaInstaller.InstallScript(testURL, tmpDir)
	require.Error(t, err, "expected error due to non-OK response")
	require.Contains(t, err.Error(), "failed to fetch script. status: 404")
}

func TestInstallScript_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test server returning some Lua code
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(
				w,
				"-- sample lua script\nprint('Hello from test!')\n",
			)
		}),
	)
	defer ts.Close()

	testURL := ts.URL + "/hello.lua"
	err := scripts.LuaInstaller.InstallScript(testURL, tmpDir)
	require.NoError(t, err, "expected successful download")

	filePath := filepath.Join(tmpDir, "hello.lua")
	info, err := os.Stat(filePath)
	require.NoError(t, err, "the downloaded .lua file should exist")
	require.False(
		t,
		info.IsDir(),
		"the downloaded file should not be a directory",
	)

	contents, err := os.ReadFile(filePath)
	require.NoError(t, err, "should be able to read downloaded file contents")
	require.Contains(
		t,
		string(contents),
		"sample lua script",
		"downloaded file content mismatch",
	)
}

func TestInstallScript_FileWriteError(t *testing.T) {
	tmpDir := t.TempDir()

	// A test server returning 200 with some content
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "some content")
		}),
	)
	defer ts.Close()

	testURL := ts.URL + "/test.lua"

	err := os.Chmod(tmpDir, 0500)
	require.NoError(t, err, "failed to chmod to read-only")

	err = scripts.LuaInstaller.InstallScript(testURL, tmpDir)
	require.Error(
		t,
		err,
		"expected error due to inability to create file in read-only directory",
	)
	require.Contains(t, err.Error(), "failed to create local file")
}
