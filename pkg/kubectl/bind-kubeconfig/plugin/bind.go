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
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"

	apis "github.com/Danil-Grigorev/rancher-bind/pkg/apis"
	managementv3 "github.com/Danil-Grigorev/rancher-bind/pkg/apis/rancher/management/v3"

	"github.com/kube-bind/kube-bind/pkg/kubectl/base"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	yaml "sigs.k8s.io/yaml"
)

// BindAPIServiceOptions are the options for the kubectl-rancher-bind command.
type BindAPIServiceOptions struct {
	Options *base.Options
	Logs    *logs.Options

	*runtime.Scheme

	file     string
	insecure bool
}

// NewRancherBindOptions returns new BindAPIServiceOptions.
func NewRancherBindOptions(streams genericclioptions.IOStreams) *BindAPIServiceOptions {
	options := &BindAPIServiceOptions{
		Options: base.NewOptions(streams),
		Logs:    logs.NewOptions(),
		Scheme:  runtime.NewScheme(),
	}

	utilruntime.Must(managementv3.AddToScheme(options.Scheme))

	return options
}

// AddCmdFlags binds fields to cmd's flagset.
func (b *BindAPIServiceOptions) AddCmdFlags(cmd *cobra.Command) {
	b.Options.BindFlags(cmd)
	logsv1.AddFlags(b.Logs, cmd.Flags())

	cmd.Flags().StringVarP(&b.file, "file", "f", b.file, "A file with a GlobalRole manifest")
	cmd.Flags().BoolVarP(&b.insecure, "insecure-skip-tls-verify", "i", b.insecure, "Sets the insecure-skip-tls-verify flag in the generated kubeconfig")
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
// - Fetch the setting pointing to the rancher url.
// - Create a GlobalRole resource.
// - Create a User resource, with a generated password.
// - Create a global role binding with sufficient permissions to obtain the token.
// - Authenticate as the user.
// - Collect the kubeconfig generated from the given token.
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

	role, err := CreateClusterRole(ctx, cl, user)
	if err != nil {
		return err
	}
	defer delete(ctx, cl, role)

	binding, err := CreateRoleBinding(ctx, cl, user)
	if err != nil {
		return err
	}
	defer delete(ctx, cl, binding)

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

	if err := ApplyUserGlobalRole(ctx, cl, user.Username, b.file); err != nil {
		return err
	}

	if err := b.DisplayKubeconfig(config); err != nil {
		return err
	}

	return nil
}

func (b *BindAPIServiceOptions) GetClient() (client.Client, error) {
	config, err := b.Options.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{Scheme: b.Scheme})
}

func (b *BindAPIServiceOptions) DisplayKubeconfig(config *apis.ConfigResponse) error {
	cfg := &clientcmdapiv1.Config{}

	decoder := apiyaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(config.Config)), 1000)
	if err := decoder.Decode(cfg); err != nil {
		return fmt.Errorf("unable to decode generated kubeconfig: %w", err)
	}

	if b.insecure {
		for i := range cfg.Clusters {
			cfg.Clusters[i].Cluster.InsecureSkipTLSVerify = true
			cfg.Clusters[i].Cluster.CertificateAuthorityData = nil
		}
	}

	if result, err := yaml.Marshal(cfg); err != nil {
		return fmt.Errorf("unable to display generated kubeconfig: %w", err)
	} else {
		fmt.Printf("%s", result)
	}

	return nil
}
