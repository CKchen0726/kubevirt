/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 *
 */

package libvmi

import (
	kvirtv1 "kubevirt.io/api/core/v1"

	cd "kubevirt.io/kubevirt/tests/containerdisk"
)

// Default VMI values
const (
	DefaultTestGracePeriod int64 = 0
	DefaultVmiName               = "testvmi"
)

// NewFedora instantiates a new Fedora based VMI configuration,
// building its extra properties based on the specified With* options.
func NewFedora(opts ...Option) *kvirtv1.VirtualMachineInstance {
	return NewTestToolingFedora(opts...)
}

// NewTestToolingFedora instantiates a new Fedora based VMI configuration,
// building its extra properties based on the specified With* options.
// This image has tooling for the guest agent, stress, and more
func NewTestToolingFedora(opts ...Option) *kvirtv1.VirtualMachineInstance {
	return newFedora(cd.ContainerDiskFedoraTestTooling, opts...)
}

// NewSriovFedora instantiates a new Fedora based VMI configuration,
// building its extra properties based on the specified With* options, the
// image used include Guest Agent and some moduled needed by SRIOV.
func NewSriovFedora(opts ...Option) *kvirtv1.VirtualMachineInstance {
	return newFedora(cd.ContainerDiskFedoraTestTooling, opts...)
}

// NewSEVFedora instantiates a new Fedora based VMI configuration,
// building its extra properties based on the specified With* options, the
// image used is configured for UEFI boot and it supports AMD SEV.
func NewSEVFedora(opts ...Option) *kvirtv1.VirtualMachineInstance {
	const secureBoot = false
	sevOptions := []Option{
		WithUefi(secureBoot),
		WithSEV(),
	}
	opts = append(sevOptions, opts...)
	return newFedora(cd.ContainerDiskFedoraTestTooling, opts...)
}

// NewFedora instantiates a new Fedora based VMI configuration with specified
// containerDisk, building its extra properties based on the specified With*
// options.
func newFedora(containerDisk cd.ContainerDisk, opts ...Option) *kvirtv1.VirtualMachineInstance {
	fedoraOptions := []Option{
		WithTerminationGracePeriod(DefaultTestGracePeriod),
		WithResourceMemory("512M"),
		WithRng(),
		WithContainerImage(cd.ContainerDiskFor(containerDisk)),
	}
	opts = append(fedoraOptions, opts...)
	return New(RandName(DefaultVmiName), opts...)
}

// NewCirros instantiates a new CirrOS based VMI configuration
func NewCirros(opts ...Option) *kvirtv1.VirtualMachineInstance {
	cirrosOpts := []Option{
		WithContainerImage(cd.ContainerDiskFor(cd.ContainerDiskCirros)),
		WithCloudInitNoCloudUserData("#!/bin/bash\necho 'hello'\n", true),
		WithResourceMemory("128Mi"),
		WithTerminationGracePeriod(DefaultTestGracePeriod),
	}
	cirrosOpts = append(cirrosOpts, opts...)
	return New(RandName(DefaultVmiName), cirrosOpts...)
}

// NewAlpine instantiates a new Alpine based VMI configuration
func NewAlpine(opts ...Option) *kvirtv1.VirtualMachineInstance {
	alpineOpts := []Option{
		WithContainerImage(cd.ContainerDiskFor(cd.ContainerDiskAlpine)),
		WithCloudInitNoCloudUserData("#!/bin/bash\necho 'hello'\n", true),
		WithResourceMemory("128Mi"),
		WithRng(),
		WithTerminationGracePeriod(DefaultTestGracePeriod),
	}
	alpineOpts = append(alpineOpts, opts...)
	return New(RandName(DefaultVmiName), alpineOpts...)
}
