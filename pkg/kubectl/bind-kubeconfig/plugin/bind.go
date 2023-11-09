/*
Copyright 2022 The Kube Bind Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"

	"github.com/kube-bind/kube-bind/pkg/kubectl/base"
)

// BindAPIServiceOptions are the options for the kubectl-bind-apiservice command.
type BindAPIServiceOptions struct {
	Options *base.Options
	Logs    *logs.Options

	JSONYamlPrintFlags *genericclioptions.JSONYamlPrintFlags
	OutputFormat       string

	file string
}

// NewRancherBindOptions returns new BindAPIServiceOptions.
func NewRancherBindOptions(streams genericclioptions.IOStreams) *BindAPIServiceOptions {
	return &BindAPIServiceOptions{
		Options: base.NewOptions(streams),
		Logs:    logs.NewOptions(),
	}
}

// AddCmdFlags binds fields to cmd's flagset.
func (b *BindAPIServiceOptions) AddCmdFlags(cmd *cobra.Command) {
	b.Options.BindFlags(cmd)
	logsv1.AddFlags(b.Logs, cmd.Flags())

	cmd.Flags().StringVarP(&b.file, "file", "f", b.file, "A file with a GlobalRole manifest")
}

// Complete ensures all fields are initialized.
func (b *BindAPIServiceOptions) Complete(args []string) error {
	return b.Options.Complete()
}

// Validate validates the NewRancherBindOptions are complete and usable.
func (b *BindAPIServiceOptions) Validate() error {
	if b.file == "" {
		return errors.New("file is required")
	}

	return b.Options.Validate()
}

// Run starts the kubeconfig generation process.
func (b *BindAPIServiceOptions) Run(ctx context.Context) error {
	return nil
}
