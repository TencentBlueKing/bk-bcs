/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package secret defines the secret command
package secret

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/secret"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

var (
	project          string
	getVersion       int
	secretConfigfile string
	secretKV         = &([]string{})
	overrideKV       = &([]string{})
	secretConfirm    bool
)

// NewSecretCmd create secret command
func NewSecretCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "secret",
		Short: "Mange secrets stored in gitops",
	}
	command.AddCommand(list())
	command.AddCommand(create())
	command.AddCommand(metadata())
	command.AddCommand(get())
	command.AddCommand(del())
	command.AddCommand(updateKeys())
	return command
}

func list() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Run: func(cmd *cobra.Command, args []string) {
			if project == "" {
				utils.ExitError("'project' param must set")
			}
			h := secret.NewHandler()
			h.List(cmd.Context(), project)
		},
	}
	c.Flags().StringVarP(&project, "project", "p", "", "Filter by project name")
	return c
}

// create return the create command
// nolint
func create() *cobra.Command {
	c := &cobra.Command{
		Use:   "create NAME",
		Short: "Create or update exist secret",
		Example: `  # Create an empty secret
  bcs-gitops secret create test-secret -p test

  # Create secret with configuration file
  powerapp secret create test-secret -p test -f ./secret.properties

  # Create secret with key-value set in command params
  powerapp secret create test-secret -p test --with-kv key1="value1" key2="value2"
`,
		Run: func(cmd *cobra.Command, args []string) {
			if project == "" {
				utils.ExitError("'project' param must set")
			}
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]
			data := parseSecretFile()
			if secretConfirm {
				fmt.Println("==> Print the secret data: ")
				for k, v := range data {
					fmt.Printf("  %s: %s\n", k, v)
				}
				fmt.Printf("Total values: %d\n", len(data))
				fmt.Printf("Are you sure you want to create secret '%s'? (Y/n): ", name)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if t := strings.ToLower(input); t == "yes" || t == "y" || t == "" {
					h := secret.NewHandler()
					h.Create(cmd.Context(), project, name, data)
				} else {
					// nolint
					fmt.Println("Cancelled.")
				}
			} else {
				h := secret.NewHandler()
				h.Create(cmd.Context(), project, name, data)
			}
		},
	}
	c.Flags().StringVarP(&project, "project", "p", "", "Filter by project name")
	c.Flags().StringVarP(&secretConfigfile, "file", "f", "",
		"The files that contain the secret configurations")
	secretKV = c.Flags().StringSliceP("with-kv", "d", nil,
		"Secret key-value (exclusion with '--file')")
	c.Flags().BoolVar(&secretConfirm, "confirm", false, "Second confirmation")
	return c
}

// metadata return metadata command
func metadata() *cobra.Command {
	c := &cobra.Command{
		Use:   "metadata NAME",
		Short: "List versions of secret",
		Run: func(cmd *cobra.Command, args []string) {
			if project == "" {
				utils.ExitError("'project' param must set")
			}
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]
			h := secret.NewHandler()
			h.GetMetadata(cmd.Context(), project, name)
		},
	}
	c.Flags().StringVarP(&project, "project", "p", "", "Filter by project name")
	return c
}

// return get command
func get() *cobra.Command {
	c := &cobra.Command{
		Use:   "get NAME",
		Short: "Get secret details with specified version",
		Run: func(cmd *cobra.Command, args []string) {
			if project == "" {
				utils.ExitError("'project' param must set")
			}
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]
			h := secret.NewHandler()
			h.GetVersion(cmd.Context(), project, name, getVersion)
		},
	}
	c.Flags().StringVarP(&project, "project", "p", "", "Filter by project name")
	c.Flags().IntVar(&getVersion, "version", 0, "Version of secret (default use latest version)")
	return c
}

// del return del command
func del() *cobra.Command {
	c := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete secret",
		Run: func(cmd *cobra.Command, args []string) {
			if project == "" {
				utils.ExitError("'project' param must set")
			}
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]
			h := secret.NewHandler()
			h.Delete(cmd.Context(), project, name)
		},
	}
	c.Flags().StringVarP(&project, "project", "p", "", "Filter by project name")
	return c
}

// nolint
func updateKeys() *cobra.Command {
	c := &cobra.Command{
		Use:   "update-keys NAME",
		Short: "Update secret with keys (update value when key exist, auto-create when key not exist)",
		Example: `  # Update secret with specified keys
  powerapp secret -p test update-keys -d k1=t1 -d k2=t2`,
		Run: func(cmd *cobra.Command, args []string) {
			if project == "" {
				utils.ExitError("'project' param must set")
			}
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name := args[0]
			if len(*overrideKV) == 0 {
				utils.ExitError("'--set-kv' must set at least one item")
			}
			h := secret.NewHandler()
			original := h.GetLatestVersion(cmd.Context(), project, name)
			result, changed := parseUpdateKey(original)
			if len(changed) == 0 {
				utils.ExitError("there not have changed keys, no need to update")
			}
			fmt.Println("==> Print the secret data(with diff): ")
			for k, v := range result {
				if _, ok1 := changed[k]; ok1 {
					if v2, ok2 := original[k]; ok2 {
						color.Red("- %s: %s", k, v2)
						color.Green("+ %s: %s", k, v)
					} else {
						color.Green("+ %s: %s", k, v)
					}
				} else {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
			fmt.Printf("Total values: %d\n", len(result))
			fmt.Printf("Are you sure you want to update secret '%s'? (Y/n): ", name)
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if t := strings.ToLower(input); t == "yes" || t == "y" || t == "" {
				h := secret.NewHandler()
				h.Create(cmd.Context(), project, name, result)
			} else {
				fmt.Println("Cancelled.")
			}
		},
	}
	c.Flags().StringVarP(&project, "project", "p", "", "Filter by project name")
	overrideKV = c.Flags().StringSliceP("set-kv", "d", nil,
		"Secret key-value (update when key exist, add when key not exist)")
	return c
}

// parseSecretFile read secret file
func parseSecretFile() map[string]string {
	if secretConfigfile == "" && len(*secretKV) == 0 {
		return make(map[string]string)
	}
	if secretConfigfile != "" && len(*secretKV) != 0 {
		fmt.Println("Warning: '--file' and '--with-kv' are both declared, just use '--file' param")
		secretKV = &([]string{})
	}
	data := make(map[string]string)
	if secretConfigfile != "" {
		bs, err := os.ReadFile(secretConfigfile)
		if err != nil {
			utils.ExitError(fmt.Sprintf("read file '%s' failed: %s", secretConfigfile, err.Error()))
		}
		slice := strings.Split(string(bs), "\n")
		for i := range slice {
			t := strings.TrimSpace(slice[i])
			if t == "" {
				continue
			}
			items := strings.SplitN(t, ":", 2)
			if len(items) != 2 {
				utils.ExitError(fmt.Sprintf("file '%s' line %d split with ':' length not 2",
					secretConfigfile, i+1))
			}
			key := items[0]
			value := items[1]
			data[key] = value
		}
		return data
	}
	if len(*secretKV) != 0 {
		kv := *secretKV
		for i := range kv {
			v := strings.TrimSpace(kv[i])
			keyValue := strings.SplitN(v, "=", 2)
			if len(keyValue) != 2 {
				utils.ExitError(fmt.Sprintf("key-value '%s' invalid input format", v))
			}
			data[keyValue[0]] = utils.TrimLeadAndTrailQuotes(keyValue[1])
		}
		return data
	}
	return data
}

// parseUpdateKey pase update key from command parameters
func parseUpdateKey(originalData map[string]string) (map[string]string, map[string]string) {
	changedData := make(map[string]string)
	kv := *overrideKV
	for i := range kv {
		v := strings.TrimSpace(kv[i])
		keyValue := strings.SplitN(v, "=", 2)
		if len(keyValue) != 2 {
			utils.ExitError(fmt.Sprintf("key-value '%s' invalid input format", v))
		}
		changedData[keyValue[0]] = utils.TrimLeadAndTrailQuotes(keyValue[1])
	}

	resultData := make(map[string]string)
	for k, v := range originalData {
		resultData[k] = v
	}
	for k, v := range changedData {
		ov, ok := resultData[k]
		if !ok || ov != v {
			resultData[k] = v
			continue
		}
		delete(changedData, k)
	}
	return resultData, changedData
}
