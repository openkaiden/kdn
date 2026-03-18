---
name: add-runtime
description: Add a new runtime implementation to the kortex-cli runtime system
argument-hint: <runtime-name>
---

# Add Runtime Skill

This skill guides you through adding a new runtime implementation to the kortex-cli runtime system.

## What are Runtimes?

Runtimes provide the execution environment for workspaces on different container/VM platforms:
- **Podman**: Container-based workspaces
- **MicroVM**: Lightweight VM-based workspaces
- **Kubernetes**: Kubernetes pod-based workspaces
- **fake**: Test runtime for development

## Steps to Add a New Runtime

### 1. Create Runtime Package

Create a new directory: `pkg/runtime/<runtime-name>/`

Example: `pkg/runtime/podman/`

### 2. Implement the Runtime Interface

Create `pkg/runtime/<runtime-name>/<runtime-name>.go` with:

```go
package <runtime-name>

import (
    "context"
    "github.com/kortex-hub/kortex-cli/pkg/runtime"
)

type <runtime-name>Runtime struct {
    storageDir string
}

// Ensure implementation of runtime.Runtime at compile time
var _ runtime.Runtime = (*<runtime-name>Runtime)(nil)

// Ensure implementation of runtime.StorageAware at compile time (optional)
var _ runtime.StorageAware = (*<runtime-name>Runtime)(nil)

// New creates a new runtime instance
func New() runtime.Runtime {
    return &<runtime-name>Runtime{}
}

// Type returns the runtime type identifier
func (r *<runtime-name>Runtime) Type() string {
    return "<runtime-name>"
}

// Initialize implements runtime.StorageAware (optional)
func (r *<runtime-name>Runtime) Initialize(storageDir string) error {
    r.storageDir = storageDir
    // Optional: create subdirectories, load state, etc.
    return nil
}

// Available implements runtimesetup.Available (optional)
func (r *<runtime-name>Runtime) Available() bool {
    // Check if the runtime is available on this system
    // Example: check if CLI tool is installed
    _, err := exec.LookPath("<runtime-cli-tool>")
    return err == nil
}

// Create creates a new runtime instance
func (r *<runtime-name>Runtime) Create(ctx context.Context, params runtime.CreateParams) (runtime.RuntimeInfo, error) {
    // Implementation: create workspace on the platform
    // Use params.Name, params.SourcePath, params.ConfigPath
    return runtime.RuntimeInfo{}, nil
}

// Start starts a runtime instance
func (r *<runtime-name>Runtime) Start(ctx context.Context, id string) (runtime.RuntimeInfo, error) {
    // Implementation: start the workspace
    return runtime.RuntimeInfo{}, nil
}

// Stop stops a runtime instance
func (r *<runtime-name>Runtime) Stop(ctx context.Context, id string) error {
    // Implementation: stop the workspace
    return nil
}

// Remove removes a runtime instance
func (r *<runtime-name>Runtime) Remove(ctx context.Context, id string) error {
    // Implementation: remove the workspace
    return nil
}

// Info retrieves information about a runtime instance
func (r *<runtime-name>Runtime) Info(ctx context.Context, id string) (runtime.RuntimeInfo, error) {
    // Implementation: get workspace info
    return runtime.RuntimeInfo{}, nil
}
```

### 3. Register the Runtime

Edit `pkg/runtimesetup/register.go`:

1. Add import:
```go
import (
    "github.com/kortex-hub/kortex-cli/pkg/runtime"
    "github.com/kortex-hub/kortex-cli/pkg/runtime/fake"
    "github.com/kortex-hub/kortex-cli/pkg/runtime/<runtime-name>"  // Add this
)
```

2. Add to `availableRuntimes` slice:
```go
var availableRuntimes = []runtimeFactory{
    fake.New,
    <runtime-name>.New,  // Add this
}
```

### 4. Add Tests

Create `pkg/runtime/<runtime-name>/<runtime-name>_test.go`:

```go
package <runtime-name>

import (
    "context"
    "testing"
)

func TestNew(t *testing.T) {
    t.Parallel()

    rt := New()
    if rt == nil {
        t.Fatal("New() returned nil")
    }

    if rt.Type() != "<runtime-name>" {
        t.Errorf("Expected type '<runtime-name>', got %s", rt.Type())
    }
}

func TestCreate(t *testing.T) {
    t.Parallel()

    // Add tests for Create method
}

// Add tests for other methods...
```

### 5. Update Copyright Headers

Run the copyright headers skill:
```bash
/copyright-headers
```

### 6. Test the Runtime

```bash
# Run tests
make test

# Build
make build

# Test with CLI (if runtime is available on your system)
./kortex-cli init --runtime <runtime-name>
```

## Required Interfaces

### Runtime Interface (required)

All runtimes MUST implement:

```go
type Runtime interface {
    Type() string
    Create(ctx context.Context, params CreateParams) (RuntimeInfo, error)
    Start(ctx context.Context, id string) (RuntimeInfo, error)
    Stop(ctx context.Context, id string) error
    Remove(ctx context.Context, id string) error
    Info(ctx context.Context, id string) (RuntimeInfo, error)
}
```

### StorageAware Interface (optional)

Implement if the runtime needs persistent storage:

```go
type StorageAware interface {
    Initialize(storageDir string) error
}
```

When implemented, the registry will:
1. Create a directory at `REGISTRY_STORAGE/<runtime-type>`
2. Call `Initialize()` with the path
3. The runtime can use this directory to persist data

### Available Interface (optional)

Implement to control runtime availability:

```go
type Available interface {
    Available() bool
}
```

Use this to:
- Check if required CLI tools are installed
- Check OS compatibility
- Check configuration prerequisites
- Check license/permission requirements

## Reference Implementation

See `pkg/runtime/fake/` for a complete reference implementation that demonstrates:
- All required Runtime interface methods
- StorageAware implementation for persistence
- Proper error handling and state management
- Comprehensive tests

## Common Patterns

### Error Handling

Use the predefined errors from `pkg/runtime`:

```go
import "github.com/kortex-hub/kortex-cli/pkg/runtime"

// Instance not found
return runtime.RuntimeInfo{}, fmt.Errorf("%w: %s", runtime.ErrInstanceNotFound, id)

// Invalid parameters
return runtime.RuntimeInfo{}, fmt.Errorf("%w: name is required", runtime.ErrInvalidParams)
```

### Persistence

If using StorageAware:

```go
func (r *myRuntime) Initialize(storageDir string) error {
    r.storageDir = storageDir
    r.storageFile = filepath.Join(storageDir, "instances.json")

    // Load existing state
    return r.loadFromDisk()
}

func (r *myRuntime) Create(...) {
    // ... create instance

    // Save to disk
    if err := r.saveToDisk(); err != nil {
        return runtime.RuntimeInfo{}, fmt.Errorf("failed to persist instance: %w", err)
    }
}
```

## Usage Example

After implementing a Podman runtime:

```bash
# Initialize workspace with Podman runtime
./kortex-cli init --runtime podman

# Start workspace
./kortex-cli workspace start <workspace-id>

# Stop workspace
./kortex-cli workspace stop <workspace-id>

# Remove workspace
./kortex-cli workspace remove <workspace-id>
```

## Notes

- Runtime names should be lowercase (e.g., `podman`, `microvm`, `k8s`)
- Use the `fake` runtime as a reference implementation
- All runtimes are registered automatically via `runtimesetup.RegisterAll()`
- Commands don't need to be modified when adding new runtimes
- Only available runtimes (those with `Available() == true`) will be registered
