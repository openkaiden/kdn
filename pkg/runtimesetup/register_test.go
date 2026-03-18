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

package runtimesetup

import (
	"context"
	"fmt"
	"testing"

	"github.com/kortex-hub/kortex-cli/pkg/runtime"
)

// fakeRegistrar is a test implementation of Registrar
type fakeRegistrar struct {
	registered []runtime.Runtime
	failNext   bool
}

func (f *fakeRegistrar) RegisterRuntime(rt runtime.Runtime) error {
	if f.failNext {
		f.failNext = false
		return runtime.ErrRuntimeNotFound // reusing an error for testing
	}
	f.registered = append(f.registered, rt)
	return nil
}

// testRuntime is a simple test runtime implementation
type testRuntime struct {
	runtimeType string
	available   bool
}

func (t *testRuntime) Type() string { return t.runtimeType }

func (t *testRuntime) Create(ctx context.Context, params runtime.CreateParams) (runtime.RuntimeInfo, error) {
	return runtime.RuntimeInfo{}, fmt.Errorf("not implemented")
}

func (t *testRuntime) Start(ctx context.Context, id string) (runtime.RuntimeInfo, error) {
	return runtime.RuntimeInfo{}, fmt.Errorf("not implemented")
}

func (t *testRuntime) Stop(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (t *testRuntime) Remove(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (t *testRuntime) Info(ctx context.Context, id string) (runtime.RuntimeInfo, error) {
	return runtime.RuntimeInfo{}, fmt.Errorf("not implemented")
}

func (t *testRuntime) Available() bool {
	return t.available
}

func TestRegisterAll(t *testing.T) {
	t.Parallel()

	t.Run("registers all runtimes", func(t *testing.T) {
		t.Parallel()

		registrar := &fakeRegistrar{}

		// Create test runtimes
		testFactories := []runtimeFactory{
			func() runtime.Runtime { return &testRuntime{runtimeType: "test1", available: true} },
			func() runtime.Runtime { return &testRuntime{runtimeType: "test2", available: true} },
		}

		err := registerAllWithAvailable(registrar, testFactories)
		if err != nil {
			t.Fatalf("registerAllWithAvailable() failed: %v", err)
		}

		// We should have registered 2 test runtimes
		if len(registrar.registered) != 2 {
			t.Errorf("Expected 2 runtimes to be registered, got %d", len(registrar.registered))
		}

		// Check that both types are present
		types := make(map[string]bool)
		for _, rt := range registrar.registered {
			types[rt.Type()] = true
		}

		if !types["test1"] {
			t.Error("Expected 'test1' runtime to be registered")
		}
		if !types["test2"] {
			t.Error("Expected 'test2' runtime to be registered")
		}
	})

	t.Run("skips unavailable runtimes", func(t *testing.T) {
		t.Parallel()

		registrar := &fakeRegistrar{}

		// Create test runtimes with one unavailable
		testFactories := []runtimeFactory{
			func() runtime.Runtime { return &testRuntime{runtimeType: "test1", available: true} },
			func() runtime.Runtime { return &testRuntime{runtimeType: "test2", available: false} },
			func() runtime.Runtime { return &testRuntime{runtimeType: "test3", available: true} },
		}

		err := registerAllWithAvailable(registrar, testFactories)
		if err != nil {
			t.Fatalf("registerAllWithAvailable() failed: %v", err)
		}

		// We should have registered only 2 runtimes (test2 is unavailable)
		if len(registrar.registered) != 2 {
			t.Errorf("Expected 2 runtimes to be registered, got %d", len(registrar.registered))
		}

		// Check that only available runtimes are present
		types := make(map[string]bool)
		for _, rt := range registrar.registered {
			types[rt.Type()] = true
		}

		if !types["test1"] {
			t.Error("Expected 'test1' runtime to be registered")
		}
		if types["test2"] {
			t.Error("Did not expect 'test2' runtime to be registered (unavailable)")
		}
		if !types["test3"] {
			t.Error("Expected 'test3' runtime to be registered")
		}
	})

	t.Run("returns error on registration failure", func(t *testing.T) {
		t.Parallel()

		registrar := &fakeRegistrar{failNext: true}

		testFactories := []runtimeFactory{
			func() runtime.Runtime { return &testRuntime{runtimeType: "test1", available: true} },
		}

		err := registerAllWithAvailable(registrar, testFactories)
		if err == nil {
			t.Fatal("Expected error when registration fails, got nil")
		}
	})
}
