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
	"testing"
)

func TestValidateImageConfig(t *testing.T) {
	t.Parallel()

	t.Run("accepts valid config", func(t *testing.T) {
		t.Parallel()

		cfg := &ImageConfig{
			Version:     "latest",
			Packages:    []string{"package1", "package2"},
			Sudo:        []string{"/usr/bin/dnf", "/bin/kill"},
			RunCommands: []string{"echo test"},
		}

		err := validateImageConfig(cfg)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("rejects nil config", func(t *testing.T) {
		t.Parallel()

		err := validateImageConfig(nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("rejects empty version", func(t *testing.T) {
		t.Parallel()

		cfg := &ImageConfig{
			Version:  "",
			Packages: []string{},
			Sudo:     []string{},
		}

		err := validateImageConfig(cfg)
		if err == nil {
			t.Error("Expected error for empty version")
		}
	})

	t.Run("accepts empty packages", func(t *testing.T) {
		t.Parallel()

		cfg := &ImageConfig{
			Version:  "latest",
			Packages: []string{},
			Sudo:     []string{},
		}

		err := validateImageConfig(cfg)
		if err != nil {
			t.Errorf("Expected empty packages to be valid, got error: %v", err)
		}
	})

	t.Run("rejects non-absolute sudo paths", func(t *testing.T) {
		t.Parallel()

		cfg := &ImageConfig{
			Version:  "latest",
			Packages: []string{},
			Sudo:     []string{"relative/path"},
		}

		err := validateImageConfig(cfg)
		if err == nil {
			t.Error("Expected error for non-absolute sudo path")
		}
	})

	t.Run("accepts absolute sudo paths", func(t *testing.T) {
		t.Parallel()

		cfg := &ImageConfig{
			Version:  "latest",
			Packages: []string{},
			Sudo:     []string{"/usr/bin/dnf", "/bin/kill"},
		}

		err := validateImageConfig(cfg)
		if err != nil {
			t.Errorf("Expected absolute sudo paths to be valid, got error: %v", err)
		}
	})

	t.Run("accepts empty run commands", func(t *testing.T) {
		t.Parallel()

		cfg := &ImageConfig{
			Version:     "latest",
			Packages:    []string{},
			Sudo:        []string{},
			RunCommands: []string{},
		}

		err := validateImageConfig(cfg)
		if err != nil {
			t.Errorf("Expected empty run commands to be valid, got error: %v", err)
		}
	})
}

func TestValidateAgentConfig(t *testing.T) {
	t.Parallel()

	t.Run("accepts valid config", func(t *testing.T) {
		t.Parallel()

		cfg := &AgentConfig{
			Packages:        []string{"package1"},
			RunCommands:     []string{"echo test"},
			TerminalCommand: []string{"claude"},
		}

		err := validateAgentConfig(cfg)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("rejects nil config", func(t *testing.T) {
		t.Parallel()

		err := validateAgentConfig(nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("accepts empty packages", func(t *testing.T) {
		t.Parallel()

		cfg := &AgentConfig{
			Packages:        []string{},
			RunCommands:     []string{},
			TerminalCommand: []string{"claude"},
		}

		err := validateAgentConfig(cfg)
		if err != nil {
			t.Errorf("Expected empty packages to be valid, got error: %v", err)
		}
	})

	t.Run("accepts empty run commands", func(t *testing.T) {
		t.Parallel()

		cfg := &AgentConfig{
			Packages:        []string{},
			RunCommands:     []string{},
			TerminalCommand: []string{"claude"},
		}

		err := validateAgentConfig(cfg)
		if err != nil {
			t.Errorf("Expected empty run commands to be valid, got error: %v", err)
		}
	})

	t.Run("rejects empty terminal command", func(t *testing.T) {
		t.Parallel()

		cfg := &AgentConfig{
			Packages:        []string{},
			RunCommands:     []string{},
			TerminalCommand: []string{},
		}

		err := validateAgentConfig(cfg)
		if err == nil {
			t.Error("Expected error for empty terminal command")
		}
	})

	t.Run("accepts terminal command with multiple elements", func(t *testing.T) {
		t.Parallel()

		cfg := &AgentConfig{
			Packages:        []string{},
			RunCommands:     []string{},
			TerminalCommand: []string{"claude", "--verbose"},
		}

		err := validateAgentConfig(cfg)
		if err != nil {
			t.Errorf("Expected multi-element terminal command to be valid, got error: %v", err)
		}
	})
}
