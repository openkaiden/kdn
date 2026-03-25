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

package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProjectConfigLoader(t *testing.T) {
	t.Parallel()

	t.Run("creates loader successfully", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if loader == nil {
			t.Error("Expected non-nil loader")
		}
	})

	t.Run("returns error for empty storage dir", func(t *testing.T) {
		t.Parallel()

		_, err := NewProjectConfigLoader("")
		if !errors.Is(err, ErrInvalidPath) {
			t.Errorf("Expected ErrInvalidPath, got %v", err)
		}
	})

	t.Run("converts to absolute path", func(t *testing.T) {
		t.Parallel()

		loader, err := NewProjectConfigLoader(".")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Access the internal field to verify it's absolute
		impl := loader.(*projectConfigLoader)
		if !filepath.IsAbs(impl.storageDir) {
			t.Errorf("Expected absolute path, got %s", impl.storageDir)
		}
	})
}

func TestProjectConfigLoader_Load(t *testing.T) {
	t.Parallel()

	t.Run("returns empty config when file doesn't exist", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		cfg, err := loader.Load("github.com/user/repo")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if cfg == nil {
			t.Error("Expected non-nil config")
		}

		// Should be empty config
		if cfg.Environment != nil || cfg.Mounts != nil {
			t.Error("Expected empty config")
		}
	})

	t.Run("returns empty config when project not found and no global config", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create projects.json with only specific project
		projectsJSON := `{
  "github.com/kortex-hub/kortex-cli": {
    "environment": [
      {
        "name": "PROJECT_VAR",
        "value": "project-value"
      }
    ]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		// Try to load different project
		cfg, err := loader.Load("github.com/other/repo")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if cfg == nil {
			t.Error("Expected non-nil config")
		}

		// Should be empty config
		if cfg.Environment != nil || cfg.Mounts != nil {
			t.Error("Expected empty config")
		}
	})

	t.Run("loads project config successfully", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		projectsJSON := `{
  "github.com/kortex-hub/kortex-cli": {
    "environment": [
      {
        "name": "PROJECT_VAR",
        "value": "project-value"
      },
      {
        "name": "API_KEY",
        "secret": "project-secret"
      }
    ],
    "mounts": {
      "dependencies": ["../project-dep"],
      "configs": [".gitconfig"]
    }
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		cfg, err := loader.Load("github.com/kortex-hub/kortex-cli")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if cfg == nil {
			t.Fatal("Expected non-nil config")
		}

		// Check environment variables
		if cfg.Environment == nil || len(*cfg.Environment) != 2 {
			t.Fatalf("Expected 2 environment variables, got %v", cfg.Environment)
		}

		env := *cfg.Environment
		if env[0].Name != "PROJECT_VAR" || *env[0].Value != "project-value" {
			t.Error("Expected PROJECT_VAR=project-value")
		}
		if env[1].Name != "API_KEY" || *env[1].Secret != "project-secret" {
			t.Error("Expected API_KEY with secret")
		}

		// Check mounts
		if cfg.Mounts == nil {
			t.Fatal("Expected mounts to be set")
		}

		if cfg.Mounts.Dependencies == nil || len(*cfg.Mounts.Dependencies) != 1 {
			t.Error("Expected 1 dependency")
		}

		if cfg.Mounts.Configs == nil || len(*cfg.Mounts.Configs) != 1 {
			t.Error("Expected 1 config")
		}
	})

	t.Run("loads global config when project not found", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create projects.json with global config only
		projectsJSON := `{
  "": {
    "mounts": {
      "configs": [".gitconfig", ".ssh"]
    }
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		// Load any project - should get global config
		cfg, err := loader.Load("github.com/any/repo")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if cfg == nil {
			t.Fatal("Expected non-nil config")
		}

		// Should have global config
		if cfg.Mounts == nil || cfg.Mounts.Configs == nil {
			t.Fatal("Expected global config to be loaded")
		}

		configs := *cfg.Mounts.Configs
		if len(configs) != 2 {
			t.Errorf("Expected 2 configs, got %d", len(configs))
		}
	})

	t.Run("merges global and project-specific config", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		projectsJSON := `{
  "": {
    "environment": [
      {
        "name": "GLOBAL_VAR",
        "value": "global-value"
      },
      {
        "name": "OVERRIDE_ME",
        "value": "global-override"
      }
    ],
    "mounts": {
      "configs": [".gitconfig", ".ssh"]
    }
  },
  "github.com/kortex-hub/kortex-cli": {
    "environment": [
      {
        "name": "PROJECT_VAR",
        "value": "project-value"
      },
      {
        "name": "OVERRIDE_ME",
        "value": "project-override"
      }
    ],
    "mounts": {
      "dependencies": ["../project-dep"]
    }
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		cfg, err := loader.Load("github.com/kortex-hub/kortex-cli")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if cfg == nil {
			t.Fatal("Expected non-nil config")
		}

		// Check environment variables - should have both global and project-specific
		if cfg.Environment == nil {
			t.Fatal("Expected environment to be set")
		}

		env := *cfg.Environment
		envMap := make(map[string]string)
		for _, e := range env {
			if e.Value != nil {
				envMap[e.Name] = *e.Value
			}
		}

		// Should have GLOBAL_VAR from global config
		if envMap["GLOBAL_VAR"] != "global-value" {
			t.Errorf("Expected GLOBAL_VAR from global config, got %v", envMap["GLOBAL_VAR"])
		}

		// Should have PROJECT_VAR from project config
		if envMap["PROJECT_VAR"] != "project-value" {
			t.Errorf("Expected PROJECT_VAR from project config, got %v", envMap["PROJECT_VAR"])
		}

		// OVERRIDE_ME should be from project config (not global)
		if envMap["OVERRIDE_ME"] != "project-override" {
			t.Errorf("Expected OVERRIDE_ME to be overridden by project config, got %v", envMap["OVERRIDE_ME"])
		}

		// Check mounts - should have both global and project-specific
		if cfg.Mounts == nil {
			t.Fatal("Expected mounts to be set")
		}

		// Should have configs from global
		if cfg.Mounts.Configs == nil || len(*cfg.Mounts.Configs) != 2 {
			t.Errorf("Expected 2 configs from global, got %v", cfg.Mounts.Configs)
		}

		// Should have dependencies from project
		if cfg.Mounts.Dependencies == nil || len(*cfg.Mounts.Dependencies) != 1 {
			t.Errorf("Expected 1 dependency from project, got %v", cfg.Mounts.Dependencies)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Write invalid JSON
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte("not valid json"), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		_, err = loader.Load("github.com/user/repo")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}

		if !errors.Is(err, ErrInvalidProjectConfig) {
			t.Errorf("Expected ErrInvalidProjectConfig, got %v", err)
		}
	})

	t.Run("returns error for invalid global configuration", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create config with both value and secret (invalid)
		projectsJSON := `{
  "": {
    "environment": [
      {
        "name": "BAD_VAR",
        "value": "value",
        "secret": "secret"
      }
    ]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		_, err = loader.Load("github.com/user/repo")
		if err == nil {
			t.Error("Expected error for invalid configuration")
		}

		if !errors.Is(err, ErrInvalidProjectConfig) {
			t.Errorf("Expected ErrInvalidProjectConfig, got %v", err)
		}
	})

	t.Run("returns error for invalid project-specific configuration", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		projectsJSON := `{
  "github.com/kortex-hub/kortex-cli": {
    "environment": [
      {
        "name": "BAD_VAR",
        "value": "value",
        "secret": "secret"
      }
    ]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		_, err = loader.Load("github.com/kortex-hub/kortex-cli")
		if err == nil {
			t.Error("Expected error for invalid configuration")
		}

		if !errors.Is(err, ErrInvalidProjectConfig) {
			t.Errorf("Expected ErrInvalidProjectConfig, got %v", err)
		}
	})

	t.Run("loads multiple projects from same file", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		projectsJSON := `{
  "github.com/user/repo1": {
    "environment": [{"name": "REPO1_VAR", "value": "repo1-value"}]
  },
  "github.com/user/repo2": {
    "environment": [{"name": "REPO2_VAR", "value": "repo2-value"}]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, ProjectsConfigFile), []byte(projectsJSON), 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}

		loader, err := NewProjectConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		// Load repo1
		cfg1, err := loader.Load("github.com/user/repo1")
		if err != nil {
			t.Errorf("Failed to load repo1 config: %v", err)
		}
		if cfg1.Environment == nil || len(*cfg1.Environment) != 1 {
			t.Error("Expected repo1 environment")
		}

		// Load repo2
		cfg2, err := loader.Load("github.com/user/repo2")
		if err != nil {
			t.Errorf("Failed to load repo2 config: %v", err)
		}
		if cfg2.Environment == nil || len(*cfg2.Environment) != 1 {
			t.Error("Expected repo2 environment")
		}
	})
}

func TestProjectConfigLoader_ModuleDesignPattern(t *testing.T) {
	t.Parallel()

	t.Run("interface can be implemented", func(t *testing.T) {
		t.Parallel()

		var _ ProjectConfigLoader = (*projectConfigLoader)(nil)
	})
}
