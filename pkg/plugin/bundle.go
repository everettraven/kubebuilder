/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
)

type bundle struct {
	name    string
	version Version
	plugins []Plugin

	supportedProjectVersions []config.Version
}

// NewBundle creates a new Bundle with the provided name and version, and that wraps the provided plugins.
// The list of supported project versions is computed from the provided plugins.
func NewBundle(name string, version Version, plugins ...Plugin) (Bundle, error) {
	supportedProjectVersions := CommonSupportedProjectVersions(plugins...)
	if len(supportedProjectVersions) == 0 {
		return nil, fmt.Errorf("in order to bundle plugins, they must all support at least one common project version")
	}

	// Plugins may be bundles themselves, so unbundle here
	// NOTE(Adirio): unbundling here ensures that Bundle.Plugin always returns a flat list of Plugins instead of also
	//               including Bundles, and therefore we don't have to use a recursive algorithm when resolving.
	allPlugins := make([]Plugin, 0, len(plugins))
	for _, plugin := range plugins {
		if pluginBundle, isBundle := plugin.(Bundle); isBundle {
			allPlugins = append(allPlugins, pluginBundle.Plugins()...)
		} else {
			allPlugins = append(allPlugins, plugin)
		}
	}

	return bundle{
		name:                     name,
		version:                  version,
		plugins:                  allPlugins,
		supportedProjectVersions: supportedProjectVersions,
	}, nil
}

// Name implements Plugin
func (b bundle) Name() string {
	return b.name
}

// Version implements Plugin
func (b bundle) Version() Version {
	return b.version
}

// SupportedProjectVersions implements Plugin
func (b bundle) SupportedProjectVersions() []config.Version {
	return b.supportedProjectVersions
}

// Plugins implements Bundle
func (b bundle) Plugins() []Plugin {
	return b.plugins
}

// POC: Dynamic Bundles
// -----
type dynamicBundle struct {
	bundle

	beforePlugins []Plugin
	afterPlugins  []Plugin
}

// Name implements Plugin
func (db dynamicBundle) Name() string {
	return db.bundle.name
}

// Version implements Plugin
func (db dynamicBundle) Version() Version {
	return db.bundle.version
}

// SupportedProjectVersions implements Plugin
func (db dynamicBundle) SupportedProjectVersions() []config.Version {
	return db.bundle.supportedProjectVersions
}

// Plugins implements Bundle
func (db dynamicBundle) Plugins() []Plugin {
	return append(db.beforePlugins, append(db.bundle.plugins, db.afterPlugins...)...)
}

func (db dynamicBundle) InjectPlugins(plugins []Plugin) {
	var pluginstring string
	for _, plugin := range plugins {
		pluginstring += fmt.Sprintf("PLUGIN NAME: %s, PLUGIN VERSION: %s, SUPPROJV: %s, PLUGIN_KEY: %s | ", plugin.Name(), plugin.Version(), plugin.SupportedProjectVersions(), KeyFor(plugin))
	}
	fmt.Printf("INJECTING PLUGINS %s\n", pluginstring)
	db.bundle.plugins = plugins
}

func NewDynamicBundle(name string, version Version, beforePlugins []Plugin, injectedPlugins []Plugin, afterPlugins []Plugin) (DynamicBundle, error) {
	supportedProjectVersions := CommonSupportedProjectVersions(append(beforePlugins, (append(injectedPlugins, afterPlugins...))...)...)
	if len(supportedProjectVersions) == 0 {
		return nil, fmt.Errorf("in order to bundle plugins, they must all support at least one common project version")
	}

	return newDynamicBundle(name, version, supportedProjectVersions, beforePlugins, injectedPlugins, afterPlugins), nil
}

func newDynamicBundle(name string, version Version, spv []config.Version, bp []Plugin, ip []Plugin, ap []Plugin) dynamicBundle {
	var db dynamicBundle
	db.bundle.name = name
	db.bundle.version = version
	db.bundle.supportedProjectVersions = spv
	db.bundle.plugins = append(db.bundle.plugins, ip...)

	db.beforePlugins = append(db.beforePlugins, bp...)
	db.afterPlugins = append(db.afterPlugins, ap...)

	return db
}

func PrintDynamicBundle(db DynamicBundle) string {
	var plugins string
	for _, plugin := range db.Plugins() {
		plugins += fmt.Sprintf("PLUGIN NAME: %s, PLUGIN VERSION: %s, SUPPROJV: %s, PLUGIN_KEY: %s | ", plugin.Name(), plugin.Version(), plugin.SupportedProjectVersions(), KeyFor(plugin))
	}

	return fmt.Sprintf("DB NAME: %s\nDB VERSION: %s\nDB SUPPRV: %s\nPLUGINS: %s", db.Name(), db.Version(), db.SupportedProjectVersions(), plugins)
}

//-----
