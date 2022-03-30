/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package resource

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// A service instance.
const (
	// ServiceNameKey is logical name of the service.
	//
	// Type: string
	// Required: Always
	// Stability: stable
	// Examples: 'shoppingcart'
	// Note: MUST be the same for all instances of horizontally scaled services. If
	// the value was not specified, SDKs MUST fallback to `unknown_service:`
	// concatenated with [`process.executable.name`](process.md#process), e.g.
	// `unknown_service:bash`. If `process.executable.name` is not available, the
	// value MUST be set to `unknown_service`.
	ServiceNameKey = attribute.Key("service.name")
	// ServiceNamespaceKey is the namespace for `service.name`.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'Shop'
	// Note: A string value having a meaning that helps to distinguish a group of
	// services, for example the team name that owns a group of services.
	// `service.name` is expected to be unique within the same namespace. If
	// `service.namespace` is not specified in the Resource then `service.name` is
	// expected to be unique for all services that have no explicit namespace defined
	// (so the empty/unspecified namespace is simply one more valid namespace). Zero-
	// length namespace string is assumed equal to unspecified namespace.
	ServiceNamespaceKey = attribute.Key("service.namespace")
	// ServiceInstanceIDKey is the string ID of the service instance.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: '627cc493-f310-47de-96bd-71410b7dec09'
	// Note: MUST be unique for each instance of the same
	// `service.namespace,service.name` pair (in other words
	// `service.namespace,service.name,service.instance.id` triplet MUST be globally
	// unique). The ID helps to distinguish instances of the same service that exist
	// at the same time (e.g. instances of a horizontally scaled service). It is
	// preferable for the ID to be persistent and stay the same for the lifetime of
	// the service instance, however it is acceptable that the ID is ephemeral and
	// changes during important lifetime events for the service (e.g. service
	// restarts). If the service has no inherent unique ID that can be used as the
	// value of this attribute it is recommended to generate a random Version 1 or
	// Version 4 RFC 4122 UUID (services aiming for reproducible UUIDs may also use
	// Version 5, see RFC 4122 for more recommendations).
	ServiceInstanceIDKey = attribute.Key("service.instance.id")
	// ServiceVersionKey sets the version string of the service API or implementation.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: '2.0.0'
	ServiceVersionKey = attribute.Key("service.version")
)

// WithAttributes adds attributes to the configured Resource.
func WithAttributes(attributes ...attribute.KeyValue) resource.Option {
	return resource.WithAttributes(attributes...)
}

// WithDetectors adds detectors to be evaluated for the configured resource.
func WithDetectors(detectors ...resource.Detector) resource.Option {
	return resource.WithDetectors(detectors...)
}

// WithFromEnv adds attributes from environment variables to the configured resource.
func WithFromEnv() resource.Option {
	return resource.WithFromEnv()
}

// WithHost adds attributes from the host to the configured resource.
func WithHost() resource.Option {
	return resource.WithHost()
}

// WithTelemetrySDK adds TelemetrySDK version info to the configured resource.
func WithTelemetrySDK() resource.Option {
	return resource.WithTelemetrySDK()
}

// WithSchemaURL sets the schema URL for the configured resource.
func WithSchemaURL(schemaURL string) resource.Option {
	return resource.WithSchemaURL(schemaURL)
}

// WithOS adds all the OS attributes to the configured Resource.
// See individual WithOS* functions to configure specific attributes.
func WithOS() resource.Option {
	return resource.WithOS()
}

// WithOSType adds an attribute with the operating system type to the configured Resource.
func WithOSType() resource.Option {
	return resource.WithOSType()
}

// WithOSDescription adds an attribute with the operating system description to the
// configured Resource. The formatted string is equivalent to the output of the
// `uname -snrvm` command.
func WithOSDescription() resource.Option {
	return resource.WithOSDescription()
}

// WithProcess adds all the Process attributes to the configured Resource.
// See individual WithProcess* functions to configure specific attributes.
func WithProcess() resource.Option {
	return resource.WithDetectors()
}

// WithProcessPID adds an attribute with the process identifier (PID) to the
// configured Resource.
func WithProcessPID() resource.Option {
	return resource.WithProcessPID()
}

// WithProcessExecutableName adds an attribute with the name of the process
// executable to the configured Resource.
func WithProcessExecutableName() resource.Option {
	return resource.WithProcessExecutableName()
}

// WithProcessExecutablePath adds an attribute with the full path to the process
// executable to the configured Resource.
func WithProcessExecutablePath() resource.Option {
	return resource.WithProcessExecutablePath()
}

// WithProcessCommandArgs adds an attribute with all the command arguments (including
// the command/executable itself) as received by the process the configured Resource.
func WithProcessCommandArgs() resource.Option {
	return resource.WithProcessCommandArgs()
}

// WithProcessOwner adds an attribute with the username of the user that owns the process
// to the configured Resource.
func WithProcessOwner() resource.Option {
	return resource.WithProcessOwner()
}

// WithProcessRuntimeName adds an attribute with the name of the runtime of this
// process to the configured Resource.
func WithProcessRuntimeName() resource.Option {
	return resource.WithProcessRuntimeName()
}

// WithProcessRuntimeVersion adds an attribute with the version of the runtime of
// this process to the configured Resource.
func WithProcessRuntimeVersion() resource.Option {
	return resource.WithProcessRuntimeVersion()
}

// WithProcessRuntimeDescription adds an attribute with an additional description
// about the runtime of the process to the configured Resource.
func WithProcessRuntimeDescription() resource.Option {
	return resource.WithProcessRuntimeDescription()
}
