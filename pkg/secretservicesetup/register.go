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

// Package secretservicesetup provides centralized registration of all available secret service implementations.
package secretservicesetup

import (
	"fmt"

	"github.com/openkaiden/kdn/pkg/secretservice"
)

// SecretServiceRegistrar is an interface for types that can register secret services.
// This is implemented by instances.Manager.
type SecretServiceRegistrar interface {
	RegisterSecretService(service secretservice.SecretService) error
}

// secretServiceFactory is a function that creates a new secret service instance.
type secretServiceFactory func() secretservice.SecretService

// availableSecretServices is the list of all secret services that can be registered.
// Add new secret services here to make them available for automatic registration.
var availableSecretServices = []secretServiceFactory{}

// RegisterAll registers all available secret service implementations to the given registrar.
// Returns an error if any secret service fails to register.
func RegisterAll(registrar SecretServiceRegistrar) error {
	return registerAllWithFactories(registrar, availableSecretServices)
}

// registerAllWithFactories registers the given secret services to the registrar.
// This function is internal and used for testing with custom secret service lists.
func registerAllWithFactories(registrar SecretServiceRegistrar, factories []secretServiceFactory) error {
	for _, factory := range factories {
		svc := factory()
		if svc == nil {
			return fmt.Errorf("secret service factory returned nil")
		}
		if err := registrar.RegisterSecretService(svc); err != nil {
			return fmt.Errorf("failed to register secret service %q: %w", svc.Name(), err)
		}
	}

	return nil
}
