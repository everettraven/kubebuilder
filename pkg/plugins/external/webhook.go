/*
Copyright 2021 The Kubernetes Authors.

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

package external

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/external"
)

var _ plugin.CreateWebhookSubcommand = &createWebhookSubcommand{}

type createWebhookSubcommand struct {
	Path string
	Args []string
}

func (p *createWebhookSubcommand) InjectResource(*resource.Resource) error {
	// Do nothing since resource flags are passed to the external plugin directly.
	return nil
}

func (p *createWebhookSubcommand) Scaffold(fs machinery.Filesystem) error {
	req := external.PluginRequest{
		APIVersion: defaultAPIVersion,
		Command:    "create webhook",
		Args:       p.Args,
	}

	req.Universe = map[string]string{}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	out, err := outputGetter.GetExecOutput(reqBytes, p.Path)
	if err != nil {
		return err
	}

	res := external.PluginResponse{}
	if err := json.Unmarshal(out, &res); err != nil {
		return err
	}

	// Error if the plugin failed.
	if res.Error {
		return fmt.Errorf(strings.Join(res.ErrorMsgs, "\n"))
	}

	currentDir, err := currentDirGetter.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	for filename, data := range res.Universe {
		f, err := fs.FS.Create(filepath.Join(currentDir, filename))
		if err != nil {
			return err
		}

		defer f.Close()

		if _, err := f.Write([]byte(data)); err != nil {
			return err
		}
	}

	return nil
}
