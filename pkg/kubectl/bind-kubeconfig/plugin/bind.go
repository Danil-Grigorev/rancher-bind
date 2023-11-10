/*
Copyright 2023 SUSE.

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
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"

	apis "github.com/Danil-Grigorev/rancher-bind/pkg/apis"

	"github.com/kube-bind/kube-bind/pkg/kubectl/base"
)

// BindAPIServiceOptions are the options for the kubectl-rancher-bind command.
type BindAPIServiceOptions struct {
	Options *base.Options
	Logs    *logs.Options

	*runtime.Scheme

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
//
// Flow:
// - Fetch the setting pointing to the rancher url
// - Create a GlobalRole resource.
// - Create a User resource, with a generated password.
// - Create a global role binding with sufficient permissions to obtain the token.
//
// - Authenticate as the user, using https://<rancher-url>/v3-public/localProviders/local?action=login, collect the new user token.
// - Collect the kubeconfig generated from the given token: `curl -s -u <token> https://<rancher-url>/v3/clusters/local?action=generateKubeconfig -X POST -H 'content-type: application/json' --insecure | jq -r .config`
// - Remove the user password to prevent further authentication.
// - Remove the temporary GlobalRole and binding.
// - Create the provided ClusterRole from file, add a role binding.
func (b *BindAPIServiceOptions) Run(ctx context.Context) error {
	cl, err := b.GetClient()
	if err != nil {
		return err
	}

	serverUrl, err := GetServer(ctx, cl)
	if err != nil {
		return err
	}

	password, hash, err := GenerateRandomPassword()
	if err != nil {
		return err
	}

	user, err := CreateUser(ctx, cl, hash)
	if err != nil {
		return err
	}

	_, err = CreateClusterRole(ctx, cl, user)
	if err != nil {
		return err
	}
	// defer delete(role)

	_, err = CreateRoleBinding(ctx, cl, user)
	if err != nil {
		return err
	}
	// defer delete(binding)

	token, err := AuthenticateUser(serverUrl, &apis.Login{
		Username: user.Username,
		Password: password,
	})
	if err != nil {
		return err
	}

	config, err := CollectKubeconfig(serverUrl, token.Token)
	if err != nil {
		return err
	}

	fmt.Printf("%s", config.Config)
	return nil
}

func (b *BindAPIServiceOptions) GetClient() (client.Client, error) {
	config, err := b.Options.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{Scheme: b.Scheme})
}
