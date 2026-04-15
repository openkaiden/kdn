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

package secretservicesetup

import (
	"errors"
	"testing"

	"github.com/openkaiden/kdn/pkg/secretservice"
)

// fakeSecretService is a test implementation of the SecretService interface
type fakeSecretService struct {
	name           string
	hostPattern    string
	path           string
	envVars        []string
	headerName     string
	headerTemplate string
}

func (f *fakeSecretService) Name() string           { return f.name }
func (f *fakeSecretService) HostPattern() string    { return f.hostPattern }
func (f *fakeSecretService) Path() string           { return f.path }
func (f *fakeSecretService) EnvVars() []string      { return f.envVars }
func (f *fakeSecretService) HeaderName() string     { return f.headerName }
func (f *fakeSecretService) HeaderTemplate() string { return f.headerTemplate }

// fakeRegistrar implements SecretServiceRegistrar for testing
type fakeRegistrar struct {
	registered map[string]secretservice.SecretService
	failOn     string // service name to fail registration on
}

func newFakeRegistrar() *fakeRegistrar {
	return &fakeRegistrar{
		registered: make(map[string]secretservice.SecretService),
	}
}

func (f *fakeRegistrar) RegisterSecretService(service secretservice.SecretService) error {
	if service.Name() == f.failOn {
		return errors.New("registration failed")
	}
	f.registered[service.Name()] = service
	return nil
}

func TestRegisterAll(t *testing.T) {
	t.Parallel()

	t.Run("registers all secret services successfully", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()

		err := RegisterAll(registrar)
		if err != nil {
			t.Errorf("RegisterAll() error = %v, want nil", err)
		}

		// No secret services are registered by default
		if len(registrar.registered) != 0 {
			t.Errorf("registered %d secret services, want 0", len(registrar.registered))
		}
	})
}

func TestRegisterAllWithFactories(t *testing.T) {
	t.Parallel()

	t.Run("registers secret services from custom factories", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()

		factories := []secretServiceFactory{
			func() secretservice.SecretService {
				return &fakeSecretService{name: "github", hostPattern: `github\.com`, headerName: "Authorization"}
			},
		}

		err := registerAllWithFactories(registrar, factories)
		if err != nil {
			t.Errorf("registerAllWithFactories() error = %v, want nil", err)
		}

		if len(registrar.registered) != 1 {
			t.Errorf("registered %d secret services, want 1", len(registrar.registered))
		}

		if _, exists := registrar.registered["github"]; !exists {
			t.Error("github secret service was not registered")
		}
	})

	t.Run("handles empty factory list", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()
		factories := []secretServiceFactory{}

		err := registerAllWithFactories(registrar, factories)
		if err != nil {
			t.Errorf("registerAllWithFactories() with empty list error = %v, want nil", err)
		}

		if len(registrar.registered) != 0 {
			t.Errorf("registered %d secret services, want 0", len(registrar.registered))
		}
	})

	t.Run("stops on first registration error", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()
		registrar.failOn = "github"

		factories := []secretServiceFactory{
			func() secretservice.SecretService {
				return &fakeSecretService{name: "github", headerName: "Authorization"}
			},
		}

		err := registerAllWithFactories(registrar, factories)
		if err == nil {
			t.Error("registerAllWithFactories() should return error when registration fails")
		}
	})

	t.Run("returns error for nil factory result", func(t *testing.T) {
		t.Parallel()

		registrar := newFakeRegistrar()

		factories := []secretServiceFactory{
			func() secretservice.SecretService {
				return nil
			},
		}

		err := registerAllWithFactories(registrar, factories)
		if err == nil {
			t.Error("registerAllWithFactories() should return error when factory returns nil")
		}
	})
}

func TestAvailableSecretServicesEmpty(t *testing.T) {
	t.Parallel()

	if len(availableSecretServices) != 0 {
		t.Errorf("availableSecretServices should be empty, got %d entries", len(availableSecretServices))
	}
}
