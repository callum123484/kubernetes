/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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
	"io"
	"os"
	"strconv"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	cmdutil "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/resource"

	"github.com/daviddengcn/go-colortext"
	"github.com/spf13/cobra"
)

func NewCmdClusterInfo(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "cluster-info",
		// clusterinfo is deprecated.
		Aliases: []string{"clusterinfo"},
		Short:   "Display cluster info",
		Long:    "Display addresses of the master and services with label kubernetes.io/cluster-service=true",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunClusterInfo(f, out, cmd)
			cmdutil.CheckErr(err)
		},
	}
	return cmd
}

func RunClusterInfo(factory *cmdutil.Factory, out io.Writer, cmd *cobra.Command) error {
	if os.Args[1] == "clusterinfo" {
		printDeprecationWarning("cluster-info", "clusterinfo")
	}

	client, err := factory.ClientConfig()
	if err != nil {
		return err
	}
	printService(out, "Kubernetes master", client.Host)

	mapper, typer := factory.Object()
	cmdNamespace, err := factory.DefaultNamespace()
	if err != nil {
		return err
	}

	// TODO use generalized labels once they are implemented (#341)
	b := resource.NewBuilder(mapper, typer, factory.ClientMapperForCommand()).
		NamespaceParam(cmdNamespace).DefaultNamespace().
		SelectorParam("kubernetes.io/cluster-service=true").
		ResourceTypeOrNameArgs(false, []string{"services"}...).
		Latest()
	b.Do().Visit(func(r *resource.Info) error {
		services := r.Object.(*api.ServiceList).Items
		for _, service := range services {
			var link string
			if len(service.Spec.PublicIPs) > 0 {
				for _, port := range service.Spec.Ports {
					link += "http://" + service.Spec.PublicIPs[0] + ":" + strconv.Itoa(port.Port) + " "
				}
			} else {
				link = client.Host + "/api/v1beta3/proxy/namespaces/" + service.ObjectMeta.Namespace + "/services/" + service.ObjectMeta.Name
			}
			printService(out, service.ObjectMeta.Labels["kubernetes.io/name"], link)
		}
		return nil
	})
	return nil

	// TODO consider printing more information about cluster
}

func printService(out io.Writer, name, link string) {
	ct.ChangeColor(ct.Green, false, ct.None, false)
	fmt.Fprint(out, name)
	ct.ResetColor()
	fmt.Fprintf(out, " is running at ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Fprint(out, link)
	ct.ResetColor()
	fmt.Fprintln(out, "")
}
