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
	workspace "github.com/kortex-hub/kortex-cli-api/workspace-configuration/go"
)

// Merger merges multiple WorkspaceConfiguration objects with proper precedence rules.
// When merging:
// - Environment variables: Later configs override earlier ones (by name)
// - Mount dependencies: Deduplicated (preserves order, no duplicates)
// - Mount configs: Deduplicated (preserves order, no duplicates)
type Merger interface {
	// Merge combines two WorkspaceConfiguration objects.
	// The override config takes precedence over the base config.
	// Returns a new merged configuration without modifying the inputs.
	Merge(base, override *workspace.WorkspaceConfiguration) *workspace.WorkspaceConfiguration
}

// merger is the internal implementation of Merger
type merger struct{}

// Compile-time check to ensure merger implements Merger interface
var _ Merger = (*merger)(nil)

// NewMerger creates a new configuration merger
func NewMerger() Merger {
	return &merger{}
}

// Merge combines two WorkspaceConfiguration objects with override taking precedence
func (m *merger) Merge(base, override *workspace.WorkspaceConfiguration) *workspace.WorkspaceConfiguration {
	// If both are nil, return nil
	if base == nil && override == nil {
		return nil
	}

	// If only base is nil, return a copy of override
	if base == nil {
		return copyConfig(override)
	}

	// If only override is nil, return a copy of base
	if override == nil {
		return copyConfig(base)
	}

	// Merge both configurations
	result := &workspace.WorkspaceConfiguration{}

	// Merge environment variables
	result.Environment = mergeEnvironment(base.Environment, override.Environment)

	// Merge mounts
	result.Mounts = mergeMounts(base.Mounts, override.Mounts)

	return result
}

// mergeEnvironment merges environment variables, with override taking precedence by name
func mergeEnvironment(base, override *[]workspace.EnvironmentVariable) *[]workspace.EnvironmentVariable {
	if base == nil && override == nil {
		return nil
	}

	// Create a map to track variables by name
	envMap := make(map[string]workspace.EnvironmentVariable)
	var order []string

	// Add base environment variables
	if base != nil {
		for _, env := range *base {
			envMap[env.Name] = env
			order = append(order, env.Name)
		}
	}

	// Override with variables from override config
	if override != nil {
		for _, env := range *override {
			if _, exists := envMap[env.Name]; !exists {
				// New variable, add to order
				order = append(order, env.Name)
			}
			// Override or add the variable
			envMap[env.Name] = env
		}
	}

	// Build result array preserving order
	if len(envMap) == 0 {
		return nil
	}

	result := make([]workspace.EnvironmentVariable, 0, len(order))
	for _, name := range order {
		result = append(result, envMap[name])
	}

	return &result
}

// mergeMounts merges mount configurations, deduplicating paths
func mergeMounts(base, override *workspace.Mounts) *workspace.Mounts {
	if base == nil && override == nil {
		return nil
	}

	result := &workspace.Mounts{}

	// Merge dependencies
	var baseDeps, overrideDeps *[]string
	if base != nil {
		baseDeps = base.Dependencies
	}
	if override != nil {
		overrideDeps = override.Dependencies
	}
	result.Dependencies = mergeStringSlices(baseDeps, overrideDeps)

	// Merge configs
	var baseConfigs, overrideConfigs *[]string
	if base != nil {
		baseConfigs = base.Configs
	}
	if override != nil {
		overrideConfigs = override.Configs
	}
	result.Configs = mergeStringSlices(baseConfigs, overrideConfigs)

	// Return nil if both are empty
	if result.Dependencies == nil && result.Configs == nil {
		return nil
	}

	return result
}

// mergeStringSlices merges two string slices, deduplicating while preserving order
func mergeStringSlices(base, override *[]string) *[]string {
	if base == nil && override == nil {
		return nil
	}

	// Use a map to track seen values
	seen := make(map[string]bool)
	var result []string

	// Add base values
	if base != nil {
		for _, value := range *base {
			if !seen[value] {
				seen[value] = true
				result = append(result, value)
			}
		}
	}

	// Add override values (deduplicating)
	if override != nil {
		for _, value := range *override {
			if !seen[value] {
				seen[value] = true
				result = append(result, value)
			}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return &result
}

// copyConfig creates a deep copy of a WorkspaceConfiguration
func copyConfig(cfg *workspace.WorkspaceConfiguration) *workspace.WorkspaceConfiguration {
	if cfg == nil {
		return nil
	}

	result := &workspace.WorkspaceConfiguration{}

	// Copy environment variables
	if cfg.Environment != nil {
		envCopy := make([]workspace.EnvironmentVariable, len(*cfg.Environment))
		copy(envCopy, *cfg.Environment)
		result.Environment = &envCopy
	}

	// Copy mounts
	if cfg.Mounts != nil {
		result.Mounts = &workspace.Mounts{}

		if cfg.Mounts.Dependencies != nil {
			depsCopy := make([]string, len(*cfg.Mounts.Dependencies))
			copy(depsCopy, *cfg.Mounts.Dependencies)
			result.Mounts.Dependencies = &depsCopy
		}

		if cfg.Mounts.Configs != nil {
			configsCopy := make([]string, len(*cfg.Mounts.Configs))
			copy(configsCopy, *cfg.Mounts.Configs)
			result.Mounts.Configs = &configsCopy
		}
	}

	return result
}
