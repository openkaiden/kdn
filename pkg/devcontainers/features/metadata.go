/**********************************************************************
 * Copyright (C) 2026 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 **********************************************************************/

package features

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// devcontainerFeatureJSON is the raw on-disk structure of devcontainer-feature.json.
type devcontainerFeatureJSON struct {
	ContainerEnv  map[string]string            `json:"containerEnv"`
	Options       map[string]featureOptionSpec `json:"options"`
	InstallsAfter []string                     `json:"installsAfter"`
}

type featureOptionSpec struct {
	Type    string      `json:"type"`
	Default interface{} `json:"default"`
	Enum    []string    `json:"enum"`
}

// featureMetadata implements FeatureMetadata.
type featureMetadata struct {
	containerEnv  map[string]string
	options       FeatureOptions
	installsAfter []string
}

var _ FeatureMetadata = (*featureMetadata)(nil)

func (m *featureMetadata) ContainerEnv() map[string]string { return m.containerEnv }
func (m *featureMetadata) Options() FeatureOptions         { return m.options }
func (m *featureMetadata) InstallsAfter() []string         { return m.installsAfter }

// featureOptions implements FeatureOptions.
type featureOptions struct {
	specs map[string]featureOptionSpec
}

var _ FeatureOptions = (*featureOptions)(nil)

var nonAlphanumRE = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// normalizeKey converts an option key to a normalized env var name:
// uppercased, runs of non-alphanumeric characters replaced with '_'.
func normalizeKey(k string) string {
	return nonAlphanumRE.ReplaceAllString(strings.ToUpper(k), "_")
}

func (o *featureOptions) Merge(userOptions map[string]interface{}) (map[string]string, error) {
	result := make(map[string]string, len(o.specs))

	// Apply defaults first.
	for key, spec := range o.specs {
		if spec.Default != nil {
			result[normalizeKey(key)] = fmt.Sprintf("%v", spec.Default)
		}
	}

	// Apply and validate user-supplied options.
	for key, val := range userOptions {
		spec, ok := o.specs[key]
		if !ok {
			return nil, fmt.Errorf("unknown option: %s", key)
		}
		norm := normalizeKey(key)

		switch spec.Type {
		case "boolean":
			switch v := val.(type) {
			case bool:
				if v {
					result[norm] = "true"
				} else {
					result[norm] = "false"
				}
			case string:
				s := strings.ToLower(v)
				if s != "true" && s != "false" {
					return nil, fmt.Errorf("option %s: expected boolean, got %q", key, v)
				}
				result[norm] = s
			default:
				return nil, fmt.Errorf("option %s: expected boolean, got %T", key, val)
			}
		case "string", "":
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("option %s: expected string, got %T", key, val)
			}
			if len(spec.Enum) > 0 {
				valid := false
				for _, e := range spec.Enum {
					if e == s {
						valid = true
						break
					}
				}
				if !valid {
					return nil, fmt.Errorf("option %s: value %q is not in enum %v", key, s, spec.Enum)
				}
			}
			result[norm] = s
		default:
			return nil, fmt.Errorf("option %s: unsupported type %q", key, spec.Type)
		}
	}

	return result, nil
}

// parseMetadata reads and parses devcontainer-feature.json from dir.
func parseMetadata(dir string) (FeatureMetadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, "devcontainer-feature.json"))
	if err != nil {
		return nil, fmt.Errorf("reading devcontainer-feature.json: %w", err)
	}

	var raw devcontainerFeatureJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing devcontainer-feature.json: %w", err)
	}

	if raw.Options == nil {
		raw.Options = map[string]featureOptionSpec{}
	}

	return &featureMetadata{
		containerEnv:  raw.ContainerEnv,
		options:       &featureOptions{specs: raw.Options},
		installsAfter: raw.InstallsAfter,
	}, nil
}
