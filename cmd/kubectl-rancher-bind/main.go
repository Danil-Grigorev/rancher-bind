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

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	kubeconfigcmd "github.com/Danil-Grigorev/rancher-bind/pkg/kubectl/bind-kubeconfig/cmd"
)

func init() {
	rand.Seed(time.Now().UnixMilli())
}

func main() {
	flags := pflag.NewFlagSet("kubectl-rancher-bind", pflag.ExitOnError)
	pflag.CommandLine = flags

	kubeconfigCmd, err := kubeconfigcmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	if err := kubeconfigCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
