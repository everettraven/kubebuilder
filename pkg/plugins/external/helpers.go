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
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

var outputGetter ExecOutputGetter = &execOutputGetter{}

// ExecOutputGetter is an interface that implements the exec output method.
type ExecOutputGetter interface {
	GetExecOutput(req []byte, path string) ([]byte, error)
}

type execOutputGetter struct{}

func (e *execOutputGetter) GetExecOutput(request []byte, path string) ([]byte, error) {
	cmd := exec.Command(path)
	cmd.Stdin = bytes.NewBuffer(request)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

var currentDirGetter OsWdGetter = &osWdGetter{}

// OsWdGetter is an interface that implements the get current directory method.
type OsWdGetter interface {
	GetCurrentDir() (string, error)
}

type osWdGetter struct{}

func (o *osWdGetter) GetCurrentDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %v", err)
	}

	return currentDir, nil
}
