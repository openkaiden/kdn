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

package secretservice

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrSecretServiceNotFound is returned when a secret service is not found in the registry.
	ErrSecretServiceNotFound = errors.New("secret service not found")
)

// Registry manages secret service implementations.
type Registry interface {
	// Register registers a secret service implementation.
	// Returns an error if a secret service with the same name is already registered.
	Register(service SecretService) error
	// Get retrieves a secret service implementation by name.
	// Returns ErrSecretServiceNotFound if the secret service is not registered.
	Get(name string) (SecretService, error)
	// List returns all registered secret service names.
	List() []string
}

// registry is the internal implementation of Registry.
type registry struct {
	mu       sync.RWMutex
	services map[string]SecretService
}

// Compile-time check to ensure registry implements Registry interface
var _ Registry = (*registry)(nil)

// NewRegistry creates a new secret service registry.
func NewRegistry() Registry {
	return &registry{
		services: make(map[string]SecretService),
	}
}

// Register registers a secret service implementation.
func (r *registry) Register(service SecretService) error {
	if service == nil {
		return errors.New("secret service cannot be nil")
	}

	name := service.Name()
	if name == "" {
		return errors.New("secret service name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("secret service %q is already registered", name)
	}

	r.services[name] = service
	return nil
}

// Get retrieves a secret service implementation by name.
func (r *registry) Get(name string) (SecretService, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, ErrSecretServiceNotFound
	}

	return service, nil
}

// List returns all registered secret service names.
func (r *registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
