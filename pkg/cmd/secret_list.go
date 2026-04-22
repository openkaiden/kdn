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
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/openkaiden/kdn/pkg/secret"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

type secretListCmd struct {
	store secret.Store
}

func (s *secretListCmd) preRun(cmd *cobra.Command, args []string) error {
	storageDir, err := cmd.Flags().GetString("storage")
	if err != nil {
		return fmt.Errorf("failed to read --storage flag: %w", err)
	}
	absStorageDir, err := filepath.Abs(storageDir)
	if err != nil {
		return fmt.Errorf("failed to resolve storage directory path: %w", err)
	}

	s.store = secret.NewStore(absStorageDir)
	return nil
}

func (s *secretListCmd) run(cmd *cobra.Command, args []string) error {
	items, err := s.store.List()
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(items) == 0 {
		fmt.Fprintln(out, "No secrets found")
		return nil
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("NAME", "TYPE", "DESCRIPTION")
	tbl.WithWriter(out)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, item := range items {
		tbl.AddRow(item.Name, item.Type, item.Description)
	}

	tbl.Print()
	return nil
}

func NewSecretListCmd() *cobra.Command {
	c := &secretListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all secrets",
		Long:  "List all secrets stored in the kdn storage directory",
		Example: `# List all secrets
kdn secret list`,
		Args:    cobra.NoArgs,
		PreRunE: c.preRun,
		RunE:    c.run,
	}

	return cmd
}
