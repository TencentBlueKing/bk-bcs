/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package main provides the drplan-gen CLI tool entry point.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/generator"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	var (
		name      string
		namespace string
		inputFile string
		outputDir string
		showVer   bool
	)

	rootCmd := &cobra.Command{
		Use:   "drplan-gen",
		Short: "Generate DRPlan from Helm template output",
		Long: "drplan-gen reads rendered Kubernetes YAML (from helm template, helmfile template, etc.)\n" +
			"and generates DRPlan, DRWorkflow, and DRPlanExecution YAML files.\n\n" +
			"Helm hook annotations (helm.sh/hook, helm.sh/hook-weight, helm.sh/hook-delete-policy)\n" +
			"are automatically recognized and mapped into a single generated workflow.\n" +
			"Hook Subscription actions are automatically generated with waitReady: true and when: mode == \"install|upgrade\".\n" +
			"By default, drplan-gen generates one stage and one workflow: drplan.yaml + workflow-install.yaml.\n\n" +
			"Examples:\n" +
			"\thelm template my-app ./my-chart | drplan-gen --name my-app --namespace default\n" +
			"\tdrplan-gen --name my-app --namespace default -f rendered.yaml\n" +
			"\tdrplan-gen --name my-app -f rendered.yaml -o ./output/",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showVer {
				fmt.Fprintln(cmd.OutOrStdout(), formatVersion())
				return nil
			}
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("required flag \"name\" not set")
			}
			return run(name, namespace, inputFile, outputDir)
		},
	}

	rootCmd.Flags().StringVar(&name, "name", "", "Release name (required)")
	rootCmd.Flags().StringVar(&namespace, "namespace", "default", "Target namespace")
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input YAML file (reads from stdin if not set)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory")
	rootCmd.Flags().BoolVar(&showVer, "version", false, "Print version information and exit")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), formatVersion())
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func formatVersion() string {
	return fmt.Sprintf("drplan-gen %s (commit=%s, buildDate=%s)", version, gitCommit, buildDate)
}

func run(name, namespace, inputFile, outputDir string) error {
	var reader io.Reader

	if inputFile != "" {
		f, err := os.Open(filepath.Clean(inputFile)) // #nosec G304
		if err != nil {
			return fmt.Errorf("opening input file: %w", err)
		}
		defer func() { _ = f.Close() }()
		reader = f
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input: provide -f flag or pipe YAML via stdin")
		}
		reader = os.Stdin
	}

	resources, err := generator.ParseYAML(reader)
	if err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	analysis := generator.Classify(resources)

	config := generator.GenerateConfig{
		ReleaseName: name,
		Namespace:   namespace,
		OutputDir:   outputDir,
	}

	result, err := generator.GeneratePlan(analysis, config)
	if err != nil {
		return fmt.Errorf("generating plan: %w", err)
	}

	if err := generator.WriteOutput(result, outputDir); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Generated DRPlan files in %s/\n", outputDir)
	fmt.Fprintf(os.Stderr, "  drplan.yaml\n")
	for filename := range result.WorkflowYAMLs {
		fmt.Fprintf(os.Stderr, "  %s\n", filename)
	}
	for filename := range result.ExecutionYAMLs {
		fmt.Fprintf(os.Stderr, "  %s\n", filename)
	}

	return nil
}
