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

package cmd

import (
	"testing"
)

func TestRemoveCmd(t *testing.T) {
	t.Parallel()

	cmd := NewRemoveCmd()
	if cmd == nil {
		t.Fatal("NewRemoveCmd() returned nil")
	}

	if cmd.Use != "remove NAME|ID" {
		t.Errorf("Expected Use to be 'remove NAME|ID', got '%s'", cmd.Use)
	}

	// Verify it has the same behavior as workspace remove
	workspaceRemoveCmd := NewWorkspaceRemoveCmd()
	if cmd.Short != workspaceRemoveCmd.Short {
		t.Errorf("Expected Short to match workspace remove, got '%s'", cmd.Short)
	}
}
