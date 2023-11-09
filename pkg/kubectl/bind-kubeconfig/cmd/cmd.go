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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	logsv1 "k8s.io/component-base/logs/api/v1"

	"github.com/Danil-Grigorev/rancher-bind/pkg/kubectl/bind-kubeconfig/plugin"
)

var (
	bindAPIServiceExampleUses = `
	# generate a kubeconfig to access rancher cluster using provided GlobalRole resource
	%[1]s -f <global-role.yaml>
	`
)

func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	opts := plugin.NewRancherBindOptions(streams)
	cmd := &cobra.Command{
		Use:     "-f <file-with-a-GlobalRole>",
		Short:   "Generate a kubeconfig for a newly created user matching the provided role",
		Example: fmt.Sprintf(bindAPIServiceExampleUses, "kubectl rancher-bind"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := logsv1.ValidateAndApply(opts.Logs, nil); err != nil {
				return err
			}

			if len(args) > 1 {
				return cmd.Help()
			}
			if err := opts.Complete(args); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}

			return opts.Run(cmd.Context())
		},
	}
	opts.AddCmdFlags(cmd)

	return cmd, nil
}
