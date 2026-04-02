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

//go:build windows

package system

import (
	"os/exec"
	"strconv"
	"strings"
)

// podmanMachineUser returns the SSH username configured for the podman machine.
func podmanMachineUser() string {
	cmd := exec.Command("podman", "machine", "inspect", "--format", "{{ .SSHConfig.RemoteUsername }}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// podmanMachineID runs "id" inside the podman machine via SSH and returns the result.
func podmanMachineID(flag string) (int, bool) {
	user := podmanMachineUser()
	if user == "" {
		return 0, false
	}
	cmd := exec.Command("podman", "machine", "ssh", "--username", user, "id", flag)
	output, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	id, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}

// Getuid returns the numeric user ID of the caller.
// On Windows, this queries the podman machine for the actual UID and falls back to 1000 if unavailable.
func (s *systemImpl) Getuid() int {
	if uid, ok := podmanMachineID("-u"); ok {
		return uid
	}
	return 1000
}

// Getgid returns the numeric group ID of the caller.
// On Windows, this queries the podman machine for the actual GID and falls back to 1000 if unavailable.
func (s *systemImpl) Getgid() int {
	if gid, ok := podmanMachineID("-g"); ok {
		return gid
	}
	return 1000
}
