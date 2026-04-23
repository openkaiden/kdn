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

package onecli

import (
	"fmt"
	"strings"

	"github.com/openkaiden/kdn/pkg/secret"
	"github.com/openkaiden/kdn/pkg/secretservice"
)

const secretTypeOther = "other"

// SecretMapper converts stored secrets to OneCLI CreateSecretInput values.
type SecretMapper interface {
	Map(item secret.ListItem, value string) (CreateSecretInput, error)
}

type secretMapper struct {
	registry secretservice.Registry
}

var _ SecretMapper = (*secretMapper)(nil)

// NewSecretMapper creates a SecretMapper that uses the given registry to look up
// secret service metadata for known secret types.
func NewSecretMapper(registry secretservice.Registry) SecretMapper {
	return &secretMapper{registry: registry}
}

// Map converts a stored secret item and its value to a CreateSecretInput.
// For type "other", the item's own fields are used directly.
// For all other types, the SecretService registry provides host pattern, header, and template.
func (m *secretMapper) Map(item secret.ListItem, value string) (CreateSecretInput, error) {
	if item.Type == secretTypeOther {
		return m.mapOtherSecret(item, value)
	}
	return m.mapKnownSecret(item, value)
}

func (m *secretMapper) mapKnownSecret(item secret.ListItem, value string) (CreateSecretInput, error) {
	svc, err := m.registry.Get(item.Type)
	if err != nil {
		return CreateSecretInput{}, fmt.Errorf("unknown secret type %q: %w", item.Type, err)
	}

	input := CreateSecretInput{
		Name:        item.Name,
		Type:        "generic",
		Value:       value,
		HostPattern: svc.HostPattern(),
		PathPattern: svc.Path(),
	}

	if headerName := svc.HeaderName(); headerName != "" {
		input.InjectionConfig = &InjectionConfig{
			HeaderName:  headerName,
			ValueFormat: convertTemplate(svc.HeaderTemplate()),
		}
	}

	return input, nil
}

func (m *secretMapper) mapOtherSecret(item secret.ListItem, value string) (CreateSecretInput, error) {
	if len(item.Hosts) > 1 {
		return CreateSecretInput{}, fmt.Errorf("secret type %q supports only one host per secret; declare one secret per host (got %d hosts)", secretTypeOther, len(item.Hosts))
	}

	hostPattern := "*"
	if len(item.Hosts) > 0 {
		hostPattern = item.Hosts[0]
	}

	input := CreateSecretInput{
		Name:        item.Name,
		Type:        "generic",
		Value:       value,
		HostPattern: hostPattern,
		PathPattern: item.Path,
	}

	if item.Header != "" {
		input.InjectionConfig = &InjectionConfig{
			HeaderName:  item.Header,
			ValueFormat: convertTemplate(item.HeaderTemplate),
		}
	}

	return input, nil
}

// convertTemplate converts kdn's ${value} placeholder to OneCLI's {value} format.
func convertTemplate(tmpl string) string {
	return strings.ReplaceAll(tmpl, "${value}", "{value}")
}
