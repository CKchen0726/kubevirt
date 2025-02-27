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
 * Copyright 2017 Red Hat, Inc.
 *
 */

package tests_test

import (
	"time"

	expect "github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"

	"kubevirt.io/kubevirt/tests/util"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/console"
	cd "kubevirt.io/kubevirt/tests/containerdisk"
	"kubevirt.io/kubevirt/tests/libvmi"
)

var _ = Describe("[rfe_id:127][posneg:negative][crit:medium][vendor:cnv-qe@redhat.com][level:component][sig-compute]Console", func() {

	var virtClient kubecli.KubevirtClient

	BeforeEach(func() {
		var err error

		virtClient, err = kubecli.GetKubevirtClient()
		util.PanicOnError(err)

		tests.BeforeTestCleanup()
	})

	ExpectConsoleOutput := func(vmi *v1.VirtualMachineInstance, expected string) {
		By("Checking that the console output equals to expected one")
		Expect(console.SafeExpectBatch(vmi, []expect.Batcher{
			&expect.BSnd{S: "\n"},
			&expect.BExp{R: expected},
		}, 120)).To(Succeed())
	}

	OpenConsole := func(vmi *v1.VirtualMachineInstance) (expect.Expecter, <-chan error) {
		By("Expecting the VirtualMachineInstance console")
		expecter, errChan, err := console.NewExpecter(virtClient, vmi, 30*time.Second)
		Expect(err).ToNot(HaveOccurred())
		return expecter, errChan
	}

	Describe("[rfe_id:127][posneg:negative][crit:medium][vendor:cnv-qe@redhat.com][level:component]A new VirtualMachineInstance", func() {
		Context("with a serial console", func() {
			Context("with a cirros image", func() {

				It("[test_id:1588]should return that we are running cirros", func() {
					vmi := libvmi.NewCirros()
					vmi = tests.RunVMIAndExpectLaunch(vmi, 30)
					ExpectConsoleOutput(
						vmi,
						"login as 'cirros' user",
					)
				})
			})

			Context("with a fedora image", func() {
				It("[sig-compute][test_id:1589]should return that we are running fedora", func() {
					vmi := libvmi.NewFedora()
					vmi = tests.RunVMIAndExpectLaunch(vmi, 30)
					ExpectConsoleOutput(
						vmi,
						"Welcome to",
					)
				})
			})

			Context("with an alpine image", func() {
				type vmiBuilder func() *v1.VirtualMachineInstance

				newVirtualMachineInstanceWithAlpineFileDisk := func() *v1.VirtualMachineInstance {
					vmi, _ := tests.NewRandomVirtualMachineInstanceWithFileDisk(cd.DataVolumeImportUrlForContainerDisk(cd.ContainerDiskAlpine), util.NamespaceTestDefault, k8sv1.ReadWriteOnce)
					return vmi
				}

				newVirtualMachineInstanceWithAlpineBlockDisk := func() *v1.VirtualMachineInstance {
					vmi, _ := tests.NewRandomVirtualMachineInstanceWithBlockDisk(cd.DataVolumeImportUrlForContainerDisk(cd.ContainerDiskAlpine), util.NamespaceTestDefault, k8sv1.ReadWriteOnce)
					return vmi
				}

				table.DescribeTable("should return that we are running alpine", func(createVMI vmiBuilder) {
					vmi := createVMI()
					vmi = tests.RunVMIAndExpectLaunch(vmi, 120)
					ExpectConsoleOutput(vmi, "login")
				},
					table.Entry("[test_id:4637][storage-req]with Filesystem Disk", newVirtualMachineInstanceWithAlpineFileDisk),
					table.Entry("[test_id:4638][storage-req]with Block Disk", newVirtualMachineInstanceWithAlpineBlockDisk),
				)
			})

			It("[test_id:1590]should be able to reconnect to console multiple times", func() {
				vmi := libvmi.NewAlpine()
				vmi = tests.RunVMIAndExpectLaunch(vmi, 30)

				for i := 0; i < 5; i++ {
					ExpectConsoleOutput(vmi, "login")
				}
			})

			It("[test_id:1591]should close console connection when new console connection is opened", func(done Done) {
				vmi := libvmi.NewAlpine()
				vmi = tests.RunVMIAndExpectLaunch(vmi, 30)

				By("opening 1st console connection")
				expecter, errChan := OpenConsole(vmi)
				defer expecter.Close()

				By("expecting error on 1st console connection")
				go func() {
					defer GinkgoRecover()
					select {
					case receivedErr := <-errChan:
						Expect(receivedErr.Error()).To(ContainSubstring("close"))
						close(done)
					case <-time.After(60 * time.Second):
						Fail("timed out waiting for closed 1st connection")
					}
				}()

				By("opening 2nd console connection")
				ExpectConsoleOutput(vmi, "login")
			}, 220)

			It("[test_id:1592]should wait until the virtual machine is in running state and return a stream interface", func() {
				vmi := libvmi.NewAlpine()
				By("Creating a new VirtualMachineInstance")
				vmi, err := virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())

				_, err = virtClient.VirtualMachineInstance(vmi.Namespace).SerialConsole(vmi.Name, &kubecli.SerialConsoleOptions{ConnectionTimeout: 30 * time.Second})
				Expect(err).ToNot(HaveOccurred())
			})

			It("[test_id:1593]should fail waiting for the virtual machine instance to be running", func() {
				vmi := libvmi.NewAlpine()
				vmi.Spec.Affinity = &k8sv1.Affinity{
					NodeAffinity: &k8sv1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &k8sv1.NodeSelector{
							NodeSelectorTerms: []k8sv1.NodeSelectorTerm{
								{
									MatchExpressions: []k8sv1.NodeSelectorRequirement{
										{Key: "kubernetes.io/hostname", Operator: k8sv1.NodeSelectorOpIn, Values: []string{"notexist"}},
									},
								},
							},
						},
					},
				}

				By("Creating a new VirtualMachineInstance")
				vmi, err := virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())

				_, err = virtClient.VirtualMachineInstance(vmi.Namespace).SerialConsole(vmi.Name, &kubecli.SerialConsoleOptions{ConnectionTimeout: 30 * time.Second})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Timeout trying to connect to the virtual machine instance"))
			})

			It("[test_id:1594]should fail waiting for the expecter", func() {
				vmi := libvmi.NewAlpine()
				vmi.Spec.Affinity = &k8sv1.Affinity{
					NodeAffinity: &k8sv1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &k8sv1.NodeSelector{
							NodeSelectorTerms: []k8sv1.NodeSelectorTerm{
								{
									MatchExpressions: []k8sv1.NodeSelectorRequirement{
										{Key: "kubernetes.io/hostname", Operator: k8sv1.NodeSelectorOpIn, Values: []string{"notexist"}},
									},
								},
							},
						},
					},
				}

				By("Creating a new VirtualMachineInstance")
				vmi, err := virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())

				By("Expecting the VirtualMachineInstance console")
				_, _, err = console.NewExpecter(virtClient, vmi, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Timeout trying to connect to the virtual machine instance"))
			})
		})

		Context("without a serial console", func() {
			var vmi *v1.VirtualMachineInstance

			BeforeEach(func() {
				vmi = tests.NewRandomVMIWithEphemeralDisk(cd.ContainerDiskFor(cd.ContainerDiskAlpine))
				f := false
				vmi.Spec.Domain.Devices.AutoattachSerialConsole = &f
			})

			It("[test_id:4116]should create the vmi without any issue", func() {
				tests.RunVMIAndExpectLaunch(vmi, 30)
			})

			It("[test_id:4117]should not have the  serial console in xml", func() {
				tests.RunVMIAndExpectLaunch(vmi, 30)

				runningVMISpec, err := tests.GetRunningVMIDomainSpec(vmi)
				Expect(err).ToNot(HaveOccurred(), "should get vmi spec without problem")

				Expect(len(runningVMISpec.Devices.Serials)).To(Equal(0), "should not have any serial consoles present")
				Expect(len(runningVMISpec.Devices.Consoles)).To(Equal(0), "should not have any virtio console for serial consoles")
			})

			It("[test_id:4118]should not connect to the serial console", func() {
				vmi = tests.RunVMIAndExpectLaunch(vmi, 30)

				_, err := virtClient.VirtualMachineInstance(vmi.ObjectMeta.Namespace).SerialConsole(vmi.ObjectMeta.Name, &kubecli.SerialConsoleOptions{})

				Expect(err.Error()).To(Equal("No serial consoles are present."), "serial console should not connect if there are no serial consoles present")
			})
		})
	})
})
