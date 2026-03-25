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

func TestNewAgentConfigLoader(t *testing.T) {
	t.Parallel()

	t.Run("creates loader successfully", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if loader == nil {
			t.Error("Expected non-nil loader")
		}
	})

	t.Run("returns error for empty storage dir", func(t *testing.T) {
		t.Parallel()

		_, err := NewAgentConfigLoader("")
		if !errors.Is(err, ErrInvalidPath) {
			t.Errorf("Expected ErrInvalidPath, got %v", err)
		}
	})

	t.Run("converts to absolute path", func(t *testing.T) {
		t.Parallel()

		loader, err := NewAgentConfigLoader(".")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Access the internal field to verify it's absolute
		impl := loader.(*agentConfigLoader)
		if !filepath.IsAbs(impl.storageDir) {
			t.Errorf("Expected absolute path, got %s", impl.storageDir)
		}
	})
}

func TestAgentConfigLoader_Load(t *testing.T) {
	t.Parallel()

	t.Run("returns empty config when file doesn't exist", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		cfg, err := loader.Load("claude")
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

	t.Run("returns empty config when agent not found", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create agents.json with only "goose" agent
		agentsJSON := `{
  "goose": {
    "environment": [
      {
        "name": "GOOSE_VAR",
        "value": "goose-value"
      }
    ]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, AgentsConfigFile), []byte(agentsJSON), 0644); err != nil {
			t.Fatalf("Failed to write agents.json: %v", err)
		}

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		// Try to load "claude" which doesn't exist
		cfg, err := loader.Load("claude")
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

	t.Run("loads agent config successfully", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		agentsJSON := `{
  "claude": {
    "environment": [
      {
        "name": "DEBUG",
        "value": "true"
      },
      {
        "name": "API_KEY",
        "secret": "my-secret"
      }
    ],
    "mounts": {
      "dependencies": ["../shared"],
      "configs": [".claude"]
    }
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, AgentsConfigFile), []byte(agentsJSON), 0644); err != nil {
			t.Fatalf("Failed to write agents.json: %v", err)
		}

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		cfg, err := loader.Load("claude")
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
		if env[0].Name != "DEBUG" || *env[0].Value != "true" {
			t.Error("Expected DEBUG=true")
		}
		if env[1].Name != "API_KEY" || *env[1].Secret != "my-secret" {
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

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Write invalid JSON
		if err := os.WriteFile(filepath.Join(configDir, AgentsConfigFile), []byte("not valid json"), 0644); err != nil {
			t.Fatalf("Failed to write agents.json: %v", err)
		}

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		_, err = loader.Load("claude")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}

		if !errors.Is(err, ErrInvalidAgentConfig) {
			t.Errorf("Expected ErrInvalidAgentConfig, got %v", err)
		}
	})

	t.Run("returns error for invalid configuration", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create config with both value and secret (invalid)
		agentsJSON := `{
  "claude": {
    "environment": [
      {
        "name": "BAD_VAR",
        "value": "value",
        "secret": "secret"
      }
    ]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, AgentsConfigFile), []byte(agentsJSON), 0644); err != nil {
			t.Fatalf("Failed to write agents.json: %v", err)
		}

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		_, err = loader.Load("claude")
		if err == nil {
			t.Error("Expected error for invalid configuration")
		}

		if !errors.Is(err, ErrInvalidAgentConfig) {
			t.Errorf("Expected ErrInvalidAgentConfig, got %v", err)
		}
	})

	t.Run("returns error for empty agent name", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		_, err = loader.Load("")
		if err == nil {
			t.Error("Expected error for empty agent name")
		}

		if !errors.Is(err, ErrInvalidAgentConfig) {
			t.Errorf("Expected ErrInvalidAgentConfig, got %v", err)
		}
	})

	t.Run("loads multiple agents from same file", func(t *testing.T) {
		t.Parallel()

		storageDir := t.TempDir()
		configDir := filepath.Join(storageDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		agentsJSON := `{
  "claude": {
    "environment": [{"name": "CLAUDE_VAR", "value": "claude-value"}]
  },
  "goose": {
    "environment": [{"name": "GOOSE_VAR", "value": "goose-value"}]
  }
}`
		if err := os.WriteFile(filepath.Join(configDir, AgentsConfigFile), []byte(agentsJSON), 0644); err != nil {
			t.Fatalf("Failed to write agents.json: %v", err)
		}

		loader, err := NewAgentConfigLoader(storageDir)
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		// Load claude
		claudeCfg, err := loader.Load("claude")
		if err != nil {
			t.Errorf("Failed to load claude config: %v", err)
		}
		if claudeCfg.Environment == nil || len(*claudeCfg.Environment) != 1 {
			t.Error("Expected claude environment")
		}

		// Load goose
		gooseCfg, err := loader.Load("goose")
		if err != nil {
			t.Errorf("Failed to load goose config: %v", err)
		}
		if gooseCfg.Environment == nil || len(*gooseCfg.Environment) != 1 {
			t.Error("Expected goose environment")
		}
	})
}

func TestAgentConfigLoader_ModuleDesignPattern(t *testing.T) {
	t.Parallel()

	t.Run("interface can be implemented", func(t *testing.T) {
		t.Parallel()

		var _ AgentConfigLoader = (*agentConfigLoader)(nil)
	})
}
