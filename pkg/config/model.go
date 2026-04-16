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

import "strings"

// ParseModelID splits a model ID encoded as "provider::model::baseURL" into its
// components. Returns (provider, model, baseURL). For plain model IDs,
// provider and baseURL are empty.
func ParseModelID(modelID string) (provider, model, baseURL string) {
	parts := strings.SplitN(modelID, "::", 3)
	switch len(parts) {
	case 3:
		return parts[0], parts[1], parts[2]
	case 2:
		return parts[0], parts[1], ""
	default:
		return "", modelID, ""
	}
}

// DisplayModelName returns just the model name, stripping the provider:: and
// ::baseURL encoding for human-readable display.
func DisplayModelName(modelID string) string {
	_, name, _ := ParseModelID(modelID)
	return name
}
