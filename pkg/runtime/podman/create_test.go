// Copyright 2026 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package podman

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workspace "github.com/kortex-hub/kortex-cli-api/workspace-configuration/go"
	"github.com/kortex-hub/kortex-cli/pkg/runtime"
)

func TestValidateDependencyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "one level up - OK",
			path:        "../main",
			expectError: false,
		},
		{
			name:        "one level up, down, up again - OK",
			path:        "../foo/../bar",
			expectError: false,
		},
		{
			name:        "two levels up - ERROR",
			path:        "../../main",
			expectError: true,
		},
		{
			name:        "one level up, down, then two levels up - ERROR",
			path:        "../foo/../../bar",
			expectError: true,
		},
		{
			name:        "current directory with subdirs - OK",
			path:        "./subdir/file",
			expectError: false,
		},
		{
			name:        "complex valid path - OK",
			path:        "../main/sub1/../sub2",
			expectError: false,
		},
		{
			name:        "complex invalid path - ERROR",
			path:        "../main/../../../invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateDependencyPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path %q, got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for path %q, got %v", tt.path, err)
				}
			}
		})
	}
}

func TestValidateCreateParams(t *testing.T) {
	t.Parallel()

	// Use a real temp directory for cross-platform testing
	tempSourcePath := t.TempDir()

	tests := []struct {
		name        string
		params      runtime.CreateParams
		expectError bool
		errorType   error
	}{
		{
			name: "valid parameters",
			params: runtime.CreateParams{
				Name:       "test-workspace",
				SourcePath: tempSourcePath,
			},
			expectError: false,
		},
		{
			name: "missing name",
			params: runtime.CreateParams{
				Name:       "",
				SourcePath: tempSourcePath,
			},
			expectError: true,
			errorType:   runtime.ErrInvalidParams,
		},
		{
			name: "missing source path",
			params: runtime.CreateParams{
				Name:       "test-workspace",
				SourcePath: "",
			},
			expectError: true,
			errorType:   runtime.ErrInvalidParams,
		},
		{
			name:        "missing both",
			params:      runtime.CreateParams{},
			expectError: true,
			errorType:   runtime.ErrInvalidParams,
		},
		{
			name: "valid dependency path",
			params: runtime.CreateParams{
				Name:       "test-workspace",
				SourcePath: tempSourcePath,
				WorkspaceConfig: &workspace.WorkspaceConfiguration{
					Mounts: &workspace.Mounts{
						Dependencies: &[]string{"../main"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid dependency path - too many levels up",
			params: runtime.CreateParams{
				Name:       "test-workspace",
				SourcePath: tempSourcePath,
				WorkspaceConfig: &workspace.WorkspaceConfiguration{
					Mounts: &workspace.Mounts{
						Dependencies: &[]string{"../../main"},
					},
				},
			},
			expectError: true,
			errorType:   runtime.ErrInvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &podmanRuntime{}
			err := p.validateCreateParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestCreateInstanceDirectory(t *testing.T) {
	t.Parallel()

	t.Run("creates instance directory", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		p := &podmanRuntime{storageDir: storageDir}

		instanceDir, err := p.createInstanceDirectory("test-workspace")
		if err != nil {
			t.Fatalf("createInstanceDirectory() failed: %v", err)
		}

		expectedDir := filepath.Join(storageDir, "instances", "test-workspace")
		if instanceDir != expectedDir {
			t.Errorf("Expected instance directory %s, got %s", expectedDir, instanceDir)
		}

		// Verify directory exists
		info, err := os.Stat(instanceDir)
		if err != nil {
			t.Errorf("Instance directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Instance path is not a directory")
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		p := &podmanRuntime{storageDir: storageDir}

		instanceDir, err := p.createInstanceDirectory("test-workspace")
		if err != nil {
			t.Fatalf("createInstanceDirectory() failed: %v", err)
		}

		// Verify both "instances" and "test-workspace" directories exist
		instancesDir := filepath.Join(storageDir, "instances")
		if _, err := os.Stat(instancesDir); err != nil {
			t.Errorf("Instances directory was not created: %v", err)
		}
		if _, err := os.Stat(instanceDir); err != nil {
			t.Errorf("Instance directory was not created: %v", err)
		}
	})
}

func TestCreateContainerfile(t *testing.T) {
	t.Parallel()

	t.Run("creates Containerfile with correct content", func(t *testing.T) {
		t.Parallel()

		instanceDir := t.TempDir()
		p := &podmanRuntime{}

		err := p.createContainerfile(instanceDir)
		if err != nil {
			t.Fatalf("createContainerfile() failed: %v", err)
		}

		// Verify Containerfile exists and starts with expected FROM line
		containerfilePath := filepath.Join(instanceDir, "Containerfile")
		content, err := os.ReadFile(containerfilePath)
		if err != nil {
			t.Fatalf("Failed to read Containerfile: %v", err)
		}

		expectedFirstLine := "FROM registry.fedoraproject.org/fedora:latest\n"
		lines := strings.Split(string(content), "\n")
		if len(lines) == 0 || lines[0]+"\n" != expectedFirstLine {
			t.Errorf("Expected Containerfile to start with:\n%s\nGot:\n%s", expectedFirstLine, lines[0])
		}
	})
}

func TestBuildContainerArgs(t *testing.T) {
	t.Parallel()

	t.Run("basic args without config", func(t *testing.T) {
		t.Parallel()

		p := &podmanRuntime{}
		// Use t.TempDir() for cross-platform path handling
		sourcePath := t.TempDir()
		params := runtime.CreateParams{
			Name:       "test-workspace",
			SourcePath: sourcePath,
		}
		imageName := "kortex-cli-test-workspace"

		args, err := p.buildContainerArgs(params, imageName)
		if err != nil {
			t.Fatalf("buildContainerArgs() failed: %v", err)
		}

		// Verify basic structure
		expectedArgs := []string{
			"create",
			"--name", "test-workspace",
			"-v", fmt.Sprintf("%s:/workspace/sources:Z", sourcePath),
			"-w", "/workspace/sources",
			"kortex-cli-test-workspace",
			"sleep", "infinity",
		}

		if len(args) != len(expectedArgs) {
			t.Fatalf("Expected %d args, got %d\nExpected: %v\nGot: %v", len(expectedArgs), len(args), expectedArgs, args)
		}

		for i, expected := range expectedArgs {
			if args[i] != expected {
				t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
			}
		}
	})

	t.Run("with environment variables", func(t *testing.T) {
		t.Parallel()

		p := &podmanRuntime{}

		debugValue := "true"
		apiKeySecret := "github-token"
		emptyValue := ""

		envVars := []workspace.EnvironmentVariable{
			{Name: "DEBUG", Value: &debugValue},
			{Name: "API_KEY", Secret: &apiKeySecret},
			{Name: "EMPTY", Value: &emptyValue},
		}

		// Use t.TempDir() for cross-platform path handling
		sourcePath := t.TempDir()
		params := runtime.CreateParams{
			Name:       "test-workspace",
			SourcePath: sourcePath,
			WorkspaceConfig: &workspace.WorkspaceConfiguration{
				Environment: &envVars,
			},
		}
		imageName := "kortex-cli-test-workspace"

		args, err := p.buildContainerArgs(params, imageName)
		if err != nil {
			t.Fatalf("buildContainerArgs() failed: %v", err)
		}

		// Check that environment variables are included
		argsStr := strings.Join(args, " ")

		if !strings.Contains(argsStr, "-e DEBUG=true") {
			t.Error("Expected DEBUG=true environment variable")
		}
		// Secrets should use --secret flag with type=env,target=ENV_VAR format
		if !strings.Contains(argsStr, "--secret github-token,type=env,target=API_KEY") {
			t.Error("Expected --secret github-token,type=env,target=API_KEY")
		}
		if !strings.Contains(argsStr, "-e EMPTY=") {
			t.Error("Expected EMPTY= environment variable")
		}
	})

	t.Run("with dependency mounts", func(t *testing.T) {
		t.Parallel()

		p := &podmanRuntime{}

		deps := []string{"../main", "../shared"}
		mounts := workspace.Mounts{
			Dependencies: &deps,
		}

		// Create a real temp directory structure for cross-platform testing
		tempDir := t.TempDir()
		projectsDir := filepath.Join(tempDir, "projects")
		currentDir := filepath.Join(projectsDir, "current")
		mainDir := filepath.Join(projectsDir, "main")
		sharedDir := filepath.Join(projectsDir, "shared")

		// Create the directories
		os.MkdirAll(currentDir, 0755)
		os.MkdirAll(mainDir, 0755)
		os.MkdirAll(sharedDir, 0755)

		params := runtime.CreateParams{
			Name:       "test-workspace",
			SourcePath: currentDir,
			WorkspaceConfig: &workspace.WorkspaceConfiguration{
				Mounts: &mounts,
			},
		}
		imageName := "kortex-cli-test-workspace"

		args, err := p.buildContainerArgs(params, imageName)
		if err != nil {
			t.Fatalf("buildContainerArgs() failed: %v", err)
		}

		// Check that dependencies are mounted
		argsStr := strings.Join(args, " ")

		// Source is mounted at /workspace/sources
		// Dependencies preserve relative paths from source:
		// ../main is mounted at /workspace/main
		// From /workspace/sources, "../main" resolves to /workspace/main
		if !strings.Contains(argsStr, "-v") {
			t.Error("Expected volume mounts")
		}

		// Build expected mount strings with cross-platform paths
		expectedMainMount := fmt.Sprintf("%s:/workspace/main:Z", mainDir)
		expectedSharedMount := fmt.Sprintf("%s:/workspace/shared:Z", sharedDir)

		if !strings.Contains(argsStr, expectedMainMount) {
			t.Errorf("Expected main dependency mount %q, got: %s", expectedMainMount, argsStr)
		}
		if !strings.Contains(argsStr, expectedSharedMount) {
			t.Errorf("Expected shared dependency mount %q, got: %s", expectedSharedMount, argsStr)
		}
	})

	t.Run("with config mounts", func(t *testing.T) {
		t.Parallel()

		p := &podmanRuntime{}

		configs := []string{".claude", ".gitconfig"}
		mounts := workspace.Mounts{
			Configs: &configs,
		}

		params := runtime.CreateParams{
			Name:       "test-workspace",
			SourcePath: "/path/to/source",
			WorkspaceConfig: &workspace.WorkspaceConfiguration{
				Mounts: &mounts,
			},
		}
		imageName := "kortex-cli-test-workspace"

		args, err := p.buildContainerArgs(params, imageName)
		if err != nil {
			t.Fatalf("buildContainerArgs() failed: %v", err)
		}

		// Get user home directory for verification
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		// Check that configs are mounted
		argsStr := strings.Join(args, " ")

		expectedClaude := filepath.Join(homeDir, ".claude") + ":/home/claude/.claude:Z"
		expectedGitconfig := filepath.Join(homeDir, ".gitconfig") + ":/home/claude/.gitconfig:Z"

		if !strings.Contains(argsStr, expectedClaude) {
			t.Errorf("Expected .claude config mount: %s", expectedClaude)
		}
		if !strings.Contains(argsStr, expectedGitconfig) {
			t.Errorf("Expected .gitconfig config mount: %s", expectedGitconfig)
		}
	})

	t.Run("with all options combined", func(t *testing.T) {
		t.Parallel()

		p := &podmanRuntime{}

		debugValue := "true"
		envVars := []workspace.EnvironmentVariable{
			{Name: "DEBUG", Value: &debugValue},
		}
		deps := []string{"../main"}
		configs := []string{".claude"}
		mounts := workspace.Mounts{
			Dependencies: &deps,
			Configs:      &configs,
		}

		// Create a real temp directory structure for cross-platform testing
		tempDir := t.TempDir()
		projectsDir := filepath.Join(tempDir, "projects")
		currentDir := filepath.Join(projectsDir, "current")
		mainDir := filepath.Join(projectsDir, "main")

		// Create the directories
		os.MkdirAll(currentDir, 0755)
		os.MkdirAll(mainDir, 0755)

		params := runtime.CreateParams{
			Name:       "test-workspace",
			SourcePath: currentDir,
			WorkspaceConfig: &workspace.WorkspaceConfiguration{
				Environment: &envVars,
				Mounts:      &mounts,
			},
		}
		imageName := "kortex-cli-test-workspace"

		args, err := p.buildContainerArgs(params, imageName)
		if err != nil {
			t.Fatalf("buildContainerArgs() failed: %v", err)
		}

		// Verify all components are present
		argsStr := strings.Join(args, " ")

		// Check structure
		if !strings.Contains(argsStr, "create") {
			t.Error("Expected 'create' command")
		}
		if !strings.Contains(argsStr, "--name test-workspace") {
			t.Error("Expected container name")
		}
		if !strings.Contains(argsStr, "-e DEBUG=true") {
			t.Error("Expected environment variable")
		}

		// Build expected mount strings with cross-platform paths
		expectedSourceMount := fmt.Sprintf("%s:/workspace/sources:Z", currentDir)
		expectedMainMount := fmt.Sprintf("%s:/workspace/main:Z", mainDir)

		if !strings.Contains(argsStr, expectedSourceMount) {
			t.Errorf("Expected source mount %q", expectedSourceMount)
		}
		if !strings.Contains(argsStr, expectedMainMount) {
			t.Errorf("Expected dependency mount %q", expectedMainMount)
		}
		if !strings.Contains(argsStr, ":/home/claude/.claude:Z") {
			t.Error("Expected config mount")
		}
		if !strings.Contains(argsStr, "-w /workspace/sources") {
			t.Error("Expected working directory")
		}
		if !strings.Contains(argsStr, imageName) {
			t.Error("Expected image name")
		}
		if !strings.Contains(argsStr, "sleep infinity") {
			t.Error("Expected sleep infinity command")
		}
	})
}
