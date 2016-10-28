/*
Copyright 2014 The Kubernetes Authors.

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

package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/kolibox/koli/pkg/cli/version"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

// NewCmdVersion command for printing the version of client and server
func NewCmdVersion(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the client and server version information",
		Run: func(cmd *cobra.Command, args []string) {
			err := runVersion(f, out, cmd)
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().Bool("client", false, "Client version only (no server required).")
	return cmd
}

func runVersion(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command) error {
	fmt.Fprintf(out, "Client Version: %#v\n", version.Get())
	if cmdutil.GetFlagBool(cmd, "client") {
		return nil
	}

	client, err := f.Client()
	if err != nil {
		return err
	}

	kubectl.GetServerVersion(out, client)
	return nil
}
