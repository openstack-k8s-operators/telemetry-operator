/*
Copyright 2024.

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

package telemetry

import "errors"

// ErrNetworkAttachmentDefinitionNotFound indicates that a network-attachment-definition was not found.
var ErrNetworkAttachmentDefinitionNotFound = errors.New("network-attachment-definition not found")

// ErrNodeIPOrHostnameNotFound indicates that an IP or HostName could not be found for a node.
var ErrNodeIPOrHostnameNotFound = errors.New("failed to find an IP or HostName for node")

// ErrMonitoringStackNil indicates that the monitoringStack is set to nil.
var ErrMonitoringStackNil = errors.New("monitoringStack is set to nil")

// ErrPersistentStorageConfigNil indicates that a nil value was received in persistent storage config.
var ErrPersistentStorageConfigNil = errors.New("received a nil value in persistent storage config")

// ErrInvalidPortNumber indicates that an invalid port number was provided.
var ErrInvalidPortNumber = errors.New("invalid port number, must be between 0 and 65535")
