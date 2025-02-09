package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ystepanoff/groolp/cli"
)

func TestInitGroolpDirectory_AlreadyExistsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	groolpDir := filepath.Join(tmpDir, ".groolp")

	err := os.Mkdir(groolpDir, 0755)
	require.NoError(t, err, "failed to create .groolp directory")

	err = cli.InitGroolpDirectory(groolpDir)
	require.NoError(
		t,
		err,
		"InitGroolpDirectory should not fail if directory already exists",
	)
}

func TestInitGroolpDirectory_AlreadyExistsFile(t *testing.T) {
	tmpDir := t.TempDir()
	groolpFile := filepath.Join(tmpDir, ".groolp")

	f, err := os.Create(groolpFile)
	require.NoError(t, err, "failed to create .groolp file")
	f.Close()

	err = cli.InitGroolpDirectory(groolpFile)
	require.Error(
		t,
		err,
		"expected an error if .groolp is a file instead of a directory",
	)
}

func TestInitGroolpDirectory_Success(t *testing.T) {
	tmpDir := t.TempDir()
	groolpDir := filepath.Join(tmpDir, ".groolp")

	err := cli.InitGroolpDirectory(groolpDir)
	require.NoError(
		t,
		err,
		"InitGroolpDirectory should succeed on fresh directory",
	)

	tasksPath := filepath.Join(groolpDir, "tasks.yaml")
	info, err := os.Stat(tasksPath)
	require.NoError(t, err, "tasks.yaml should be created")
	require.False(t, info.IsDir(), "tasks.yaml should be a file")

	scriptsPath := filepath.Join(groolpDir, "scripts")
	info, err = os.Stat(scriptsPath)
	require.NoError(t, err, "scripts directory should be created")
	require.True(t, info.IsDir(), "scripts should be a directory")

	sampleScript := filepath.Join(scriptsPath, "sample.lua")
	info, err = os.Stat(sampleScript)
	require.NoError(t, err, "sample script should exist in scripts directory")
	require.False(t, info.IsDir(), "sample.lua should be a file")
}

func TestInitTasksConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	groolpDir := filepath.Join(tmpDir, ".groolp")

	err := cli.InitGroolpDirectory(groolpDir)
	require.NoError(t, err, "should successfully initialize .groolp dir")

	config, err := cli.InitTasksConfig(groolpDir)
	require.NoError(
		t,
		err,
		"InitTasksConfig should succeed after initialization",
	)
	require.NotNil(t, config, "config should not be nil")

	require.Contains(
		t,
		config.Tasks,
		"sample-yaml-task",
		"Expected sample-yaml-task in config",
	)
}

func TestInitTasksConfig_NoSuchDir(t *testing.T) {
	tmpDir := t.TempDir()
	groolpDir := filepath.Join(tmpDir, ".groolp")

	config, err := cli.InitTasksConfig(groolpDir)
	require.Error(
		t,
		err,
		"InitTasksConfig should fail if tasks.yaml doesn't exist",
	)
	require.Nil(t, config, "config should be nil on error")
}

func TestInitTasksConfig_VerifySampleTaskDefinition(t *testing.T) {
	tmpDir := t.TempDir()
	groolpDir := filepath.Join(tmpDir, ".groolp")

	err := cli.InitGroolpDirectory(groolpDir)
	require.NoError(t, err)

	config, err := cli.InitTasksConfig(groolpDir)
	require.NoError(t, err)
	require.NotNil(t, config)

	sampleTask, ok := config.Tasks["sample-yaml-task"]
	require.True(t, ok, "sample-yaml-task should be present")
	require.Equal(
		t,
		"echo Hello from tasks.yaml!",
		sampleTask.Action,
		"mismatch in sample task action",
	)
	require.Equal(
		t,
		"A sample task from tasks.yaml",
		sampleTask.Description,
		"mismatch in sample task description",
	)
}
