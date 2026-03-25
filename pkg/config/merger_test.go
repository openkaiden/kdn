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

	workspace "github.com/kortex-hub/kortex-cli-api/workspace-configuration/go"
)

func TestMerger_Merge_NilInputs(t *testing.T) {
	t.Parallel()

	merger := NewMerger()

	t.Run("both nil", func(t *testing.T) {
		t.Parallel()

		result := merger.Merge(nil, nil)
		if result != nil {
			t.Error("Expected nil result when both inputs are nil")
		}
	})

	t.Run("base nil", func(t *testing.T) {
		t.Parallel()

		override := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "TEST", Value: strPtr("value")},
			},
		}

		result := merger.Merge(nil, override)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if result.Environment == nil || len(*result.Environment) != 1 {
			t.Error("Expected environment to be copied from override")
		}
	})

	t.Run("override nil", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "TEST", Value: strPtr("value")},
			},
		}

		result := merger.Merge(base, nil)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if result.Environment == nil || len(*result.Environment) != 1 {
			t.Error("Expected environment to be copied from base")
		}
	})
}

func TestMerger_Merge_Environment(t *testing.T) {
	t.Parallel()

	merger := NewMerger()

	t.Run("no overlap", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Value: strPtr("base1")},
				{Name: "VAR2", Value: strPtr("base2")},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR3", Value: strPtr("override3")},
				{Name: "VAR4", Value: strPtr("override4")},
			},
		}

		result := merger.Merge(base, override)

		if result.Environment == nil {
			t.Fatal("Expected environment to be set")
		}

		env := *result.Environment
		if len(env) != 4 {
			t.Errorf("Expected 4 environment variables, got %d", len(env))
		}

		// Check that all variables are present
		envMap := make(map[string]string)
		for _, e := range env {
			if e.Value != nil {
				envMap[e.Name] = *e.Value
			}
		}

		if envMap["VAR1"] != "base1" {
			t.Error("VAR1 not preserved from base")
		}
		if envMap["VAR2"] != "base2" {
			t.Error("VAR2 not preserved from base")
		}
		if envMap["VAR3"] != "override3" {
			t.Error("VAR3 not added from override")
		}
		if envMap["VAR4"] != "override4" {
			t.Error("VAR4 not added from override")
		}
	})

	t.Run("override takes precedence", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Value: strPtr("base-value")},
				{Name: "VAR2", Value: strPtr("keep-this")},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Value: strPtr("override-value")},
				{Name: "VAR3", Value: strPtr("new-var")},
			},
		}

		result := merger.Merge(base, override)

		env := *result.Environment
		if len(env) != 3 {
			t.Errorf("Expected 3 environment variables, got %d", len(env))
		}

		envMap := make(map[string]string)
		for _, e := range env {
			if e.Value != nil {
				envMap[e.Name] = *e.Value
			}
		}

		if envMap["VAR1"] != "override-value" {
			t.Errorf("Expected VAR1='override-value', got '%s'", envMap["VAR1"])
		}
		if envMap["VAR2"] != "keep-this" {
			t.Error("VAR2 should be preserved")
		}
		if envMap["VAR3"] != "new-var" {
			t.Error("VAR3 should be added")
		}
	})

	t.Run("value vs secret", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Value: strPtr("value1")},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Secret: strPtr("secret-ref")},
			},
		}

		result := merger.Merge(base, override)

		env := *result.Environment
		if len(env) != 1 {
			t.Fatalf("Expected 1 environment variable, got %d", len(env))
		}

		if env[0].Secret == nil || *env[0].Secret != "secret-ref" {
			t.Error("Expected secret to override value")
		}
		if env[0].Value != nil {
			t.Error("Expected value to be nil after secret override")
		}
	})

	t.Run("preserves order", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "A", Value: strPtr("a")},
				{Name: "B", Value: strPtr("b")},
				{Name: "C", Value: strPtr("c")},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "B", Value: strPtr("b-override")},
				{Name: "D", Value: strPtr("d")},
			},
		}

		result := merger.Merge(base, override)

		env := *result.Environment
		// Order should be: A (base), B (base position but override value), C (base), D (override)
		if len(env) != 4 {
			t.Fatalf("Expected 4 variables, got %d", len(env))
		}

		if env[0].Name != "A" {
			t.Errorf("Expected first variable to be A, got %s", env[0].Name)
		}
		if env[1].Name != "B" {
			t.Errorf("Expected second variable to be B, got %s", env[1].Name)
		}
		if env[2].Name != "C" {
			t.Errorf("Expected third variable to be C, got %s", env[2].Name)
		}
		if env[3].Name != "D" {
			t.Errorf("Expected fourth variable to be D, got %s", env[3].Name)
		}
	})

	t.Run("empty base", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{}
		override := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Value: strPtr("value1")},
			},
		}

		result := merger.Merge(base, override)

		if result.Environment == nil || len(*result.Environment) != 1 {
			t.Error("Expected environment from override")
		}
	})

	t.Run("empty override", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "VAR1", Value: strPtr("value1")},
			},
		}
		override := &workspace.WorkspaceConfiguration{}

		result := merger.Merge(base, override)

		if result.Environment == nil || len(*result.Environment) != 1 {
			t.Error("Expected environment from base")
		}
	})
}

func TestMerger_Merge_Mounts(t *testing.T) {
	t.Parallel()

	merger := NewMerger()

	t.Run("dependencies no overlap", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{"../dep1", "../dep2"},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{"../dep3", "../dep4"},
			},
		}

		result := merger.Merge(base, override)

		if result.Mounts == nil || result.Mounts.Dependencies == nil {
			t.Fatal("Expected dependencies to be set")
		}

		deps := *result.Mounts.Dependencies
		if len(deps) != 4 {
			t.Errorf("Expected 4 dependencies, got %d", len(deps))
		}
	})

	t.Run("dependencies deduplication", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{"../dep1", "../dep2"},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{"../dep2", "../dep3"},
			},
		}

		result := merger.Merge(base, override)

		deps := *result.Mounts.Dependencies
		if len(deps) != 3 {
			t.Errorf("Expected 3 unique dependencies, got %d", len(deps))
		}

		// Check order: dep1, dep2 (from base), dep3 (new from override)
		if deps[0] != "../dep1" || deps[1] != "../dep2" || deps[2] != "../dep3" {
			t.Errorf("Unexpected order: %v", deps)
		}
	})

	t.Run("configs no overlap", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Configs: &[]string{".gitconfig", ".ssh"},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Configs: &[]string{".kube", ".docker"},
			},
		}

		result := merger.Merge(base, override)

		if result.Mounts == nil || result.Mounts.Configs == nil {
			t.Fatal("Expected configs to be set")
		}

		configs := *result.Mounts.Configs
		if len(configs) != 4 {
			t.Errorf("Expected 4 configs, got %d", len(configs))
		}
	})

	t.Run("configs deduplication", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Configs: &[]string{".gitconfig", ".ssh"},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Configs: &[]string{".ssh", ".kube"},
			},
		}

		result := merger.Merge(base, override)

		configs := *result.Mounts.Configs
		if len(configs) != 3 {
			t.Errorf("Expected 3 unique configs, got %d", len(configs))
		}

		// Check order: .gitconfig, .ssh (from base), .kube (new from override)
		if configs[0] != ".gitconfig" || configs[1] != ".ssh" || configs[2] != ".kube" {
			t.Errorf("Unexpected order: %v", configs)
		}
	})

	t.Run("empty slices return nil", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{},
			},
		}

		override := &workspace.WorkspaceConfiguration{
			Mounts: &workspace.Mounts{
				Configs: &[]string{},
			},
		}

		result := merger.Merge(base, override)

		if result.Mounts != nil {
			t.Error("Expected mounts to be nil when all slices are empty")
		}
	})
}

func TestMerger_Merge_MultiLevel(t *testing.T) {
	t.Parallel()

	merger := NewMerger()

	t.Run("three level merge", func(t *testing.T) {
		t.Parallel()

		// Workspace level
		workspaceCfg := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "LEVEL", Value: strPtr("workspace")},
				{Name: "WORKSPACE_VAR", Value: strPtr("ws-value")},
			},
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{"../workspace-dep"},
			},
		}

		// Project level
		projectCfg := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "LEVEL", Value: strPtr("project")},
				{Name: "PROJECT_VAR", Value: strPtr("proj-value")},
			},
			Mounts: &workspace.Mounts{
				Dependencies: &[]string{"../project-dep"},
				Configs:      &[]string{".gitconfig"},
			},
		}

		// Agent level
		agentCfg := &workspace.WorkspaceConfiguration{
			Environment: &[]workspace.EnvironmentVariable{
				{Name: "LEVEL", Value: strPtr("agent")},
				{Name: "AGENT_VAR", Value: strPtr("agent-value")},
			},
			Mounts: &workspace.Mounts{
				Configs: &[]string{".claude"},
			},
		}

		// Merge: workspace -> project -> agent
		merged1 := merger.Merge(workspaceCfg, projectCfg)
		result := merger.Merge(merged1, agentCfg)

		// Check environment variables
		if result.Environment == nil {
			t.Fatal("Expected environment to be set")
		}

		env := *result.Environment
		envMap := make(map[string]string)
		for _, e := range env {
			if e.Value != nil {
				envMap[e.Name] = *e.Value
			}
		}

		// LEVEL should be from agent (highest precedence)
		if envMap["LEVEL"] != "agent" {
			t.Errorf("Expected LEVEL='agent', got '%s'", envMap["LEVEL"])
		}

		// All other vars should be present
		if envMap["WORKSPACE_VAR"] != "ws-value" {
			t.Error("WORKSPACE_VAR should be preserved")
		}
		if envMap["PROJECT_VAR"] != "proj-value" {
			t.Error("PROJECT_VAR should be preserved")
		}
		if envMap["AGENT_VAR"] != "agent-value" {
			t.Error("AGENT_VAR should be added")
		}

		// Check mounts
		if result.Mounts == nil {
			t.Fatal("Expected mounts to be set")
		}

		deps := *result.Mounts.Dependencies
		if len(deps) != 2 {
			t.Errorf("Expected 2 dependencies, got %d", len(deps))
		}

		configs := *result.Mounts.Configs
		if len(configs) != 2 {
			t.Errorf("Expected 2 configs, got %d", len(configs))
		}
	})
}

func TestMerger_Merge_EmptyConfigurations(t *testing.T) {
	t.Parallel()

	merger := NewMerger()

	t.Run("both empty", func(t *testing.T) {
		t.Parallel()

		base := &workspace.WorkspaceConfiguration{}
		override := &workspace.WorkspaceConfiguration{}

		result := merger.Merge(base, override)

		if result == nil {
			t.Error("Expected non-nil result")
		}

		if result.Environment != nil {
			t.Error("Expected environment to be nil")
		}

		if result.Mounts != nil {
			t.Error("Expected mounts to be nil")
		}
	})
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
