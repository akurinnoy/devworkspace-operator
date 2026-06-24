// Copyright (c) 2019-2025 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers_test

import (
	"time"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	controllerv1alpha1 "github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	"github.com/devfile/devworkspace-operator/pkg/common"
	"github.com/devfile/devworkspace-operator/pkg/config"
	"github.com/devfile/devworkspace-operator/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// createInitContainerTestDW creates a DevWorkspace for init container tests and waits
// for it to have a DevWorkspace ID assigned. Unlike createDevWorkspace, this function
// correctly uses the provided name (not the hardcoded devWorkspaceName constant).
func createInitContainerTestDW(name, fromFile string) {
	By("Loading DevWorkspace from test file")
	devworkspace := &dw.DevWorkspace{}
	Expect(loadObjectFromFile(name, devworkspace, fromFile)).Should(Succeed())

	By("Creating DevWorkspace on cluster")
	Expect(k8sClient.Create(ctx, devworkspace)).Should(Succeed())

	By("Waiting for DevWorkspace ID to be assigned")
	createdDW := &dw.DevWorkspace{}
	Eventually(func() bool {
		if err := k8sClient.Get(ctx, namespacedName(name, testNamespace), createdDW); err != nil {
			return false
		}
		return createdDW.Status.DevWorkspaceId != ""
	}, 10*time.Second, 250*time.Millisecond).Should(BeTrue(),
		"DevWorkspace %s should have a DevWorkspaceId assigned", name)
}

// getInitContainerTestDW gets an existing DevWorkspace for init container tests.
func getInitContainerTestDW(name string) *dw.DevWorkspace {
	By("Getting existing DevWorkspace")
	dw := &dw.DevWorkspace{}
	dwNN := namespacedName(name, testNamespace)
	Eventually(func() (string, error) {
		if err := k8sClient.Get(ctx, dwNN, dw); err != nil {
			return "", err
		}
		return dw.Status.DevWorkspaceId, nil
	}, timeout, interval).Should(Not(BeEmpty()),
		"DevWorkspace %s should have a DevWorkspaceId", name)
	return dw
}

var _ = Describe("Init Container Injection", func() {
	const initContainerTestDWName = "test-init-container-dw"

	AfterEach(func() {
		deleteDevWorkspace(initContainerTestDWName)
		config.SetGlobalConfigForTesting(nil)
	})

	Context("DWOC InitContainers injection", func() {

		It("Scenario 1: No DWOC init containers — devfile containers unchanged", func() {
			By("Setting global config with no InitContainers")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{},
			})

			By("Creating DevWorkspace")
			createInitContainerTestDW(initContainerTestDWName, "test-devworkspace.yaml")
			devworkspace := getInitContainerTestDW(initContainerTestDWName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady("test-url", common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for Deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that no DWOC-injected init containers are present")
			for _, initContainer := range deploy.Spec.Template.Spec.InitContainers {
				Expect(initContainer.Name).ShouldNot(Equal(constants.HomeInitComponentName),
					"init-persistent-home should not be present when no DWOC init containers are configured")
			}
		})

		It("Scenario 2: DWOC has custom init-persistent-home with valid args — merged into pod additions", func() {
			By("Setting global config with init-persistent-home having valid args")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(true),
					},
					InitContainers: []corev1.Container{
						{
							Name: constants.HomeInitComponentName,
							Args: []string{"echo 'custom persistent home init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createInitContainerTestDW(initContainerTestDWName, "test-devworkspace.yaml")
			devworkspace := getInitContainerTestDW(initContainerTestDWName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady("test-url", common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for Deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that init-persistent-home is present in init containers")
			var homeInitContainer *corev1.Container
			for i := range deploy.Spec.Template.Spec.InitContainers {
				if deploy.Spec.Template.Spec.InitContainers[i].Name == constants.HomeInitComponentName {
					homeInitContainer = &deploy.Spec.Template.Spec.InitContainers[i]
					break
				}
			}
			Expect(homeInitContainer).ShouldNot(BeNil(),
				"init-persistent-home container should be present in deployment init containers")
			Expect(homeInitContainer.Command).Should(Equal([]string{"/bin/sh", "-c"}),
				"init-persistent-home container should have default command set by EnsureHomeInitContainerFields")
		})

		It("Scenario 3: DWOC has custom init-persistent-home with wrong command — workspace fails", func() {
			By("Setting global config with init-persistent-home having invalid command")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(true),
					},
					InitContainers: []corev1.Container{
						{
							Name:    constants.HomeInitComponentName,
							Command: []string{"/wrong/command"},
							Args:    []string{"echo 'custom persistent home init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createInitContainerTestDW(initContainerTestDWName, "test-devworkspace.yaml")
			dwNamespacedName := namespacedName(initContainerTestDWName, testNamespace)

			By("Checking that DevWorkspace enters Failed phase")
			currDW := &dw.DevWorkspace{}
			Eventually(func() (dw.DevWorkspacePhase, error) {
				if err := k8sClient.Get(ctx, dwNamespacedName, currDW); err != nil {
					return "", err
				}
				GinkgoWriter.Printf("Waiting for DevWorkspace to fail -- Phase: %s, Message: %s\n",
					currDW.Status.Phase, currDW.Status.Message)
				return currDW.Status.Phase, nil
			}, timeout, interval).Should(Equal(dw.DevWorkspaceStatusFailed),
				"DevWorkspace should fail due to invalid init-persistent-home command")

			By("Checking that the failure message references invalid command")
			Expect(currDW.Status.Message).Should(ContainSubstring(
				"Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]"),
				"Failure message should describe the invalid command error")
		})

		It("Scenario 4: DWOC has non-home init container — appended after base", func() {
			By("Setting global config with a custom non-home init container")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					InitContainers: []corev1.Container{
						{
							Name:  "custom-init-container",
							Image: "busybox:latest",
							Args:  []string{"echo 'custom non-home init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createInitContainerTestDW(initContainerTestDWName, "test-devworkspace.yaml")
			devworkspace := getInitContainerTestDW(initContainerTestDWName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady("test-url", common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for Deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that custom init container is present in deployment init containers")
			initContainerNames := make([]string, 0, len(deploy.Spec.Template.Spec.InitContainers))
			for _, c := range deploy.Spec.Template.Spec.InitContainers {
				initContainerNames = append(initContainerNames, c.Name)
			}

			Expect(initContainerNames).Should(ContainElement("custom-init-container"),
				"custom-init-container should be present in deployment init containers")

			By("Checking that custom init container appears after base init containers from devfile")
			customInitIdx := -1
			for i, name := range initContainerNames {
				if name == "custom-init-container" {
					customInitIdx = i
					break
				}
			}
			Expect(customInitIdx).Should(BeNumerically(">", -1),
				"custom-init-container should be found in init containers list")

			// Verify any project-clone container (base init) comes before custom-init-container
			for i, name := range initContainerNames {
				if name == "project-clone" {
					Expect(i).Should(BeNumerically("<", customInitIdx),
						"project-clone init container should come before custom-init-container")
				}
			}
		})

		It("Scenario 5: persistUserHome.enabled=false — init-persistent-home from DWOC skipped", func() {
			By("Setting global config with persistUserHome disabled but init-persistent-home in InitContainers")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(false),
					},
					InitContainers: []corev1.Container{
						{
							Name: constants.HomeInitComponentName,
							Args: []string{"echo 'custom persistent home init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createInitContainerTestDW(initContainerTestDWName, "test-devworkspace.yaml")
			devworkspace := getInitContainerTestDW(initContainerTestDWName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady("test-url", common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for Deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that init-persistent-home is NOT present in init containers")
			for _, initContainer := range deploy.Spec.Template.Spec.InitContainers {
				Expect(initContainer.Name).ShouldNot(Equal(constants.HomeInitComponentName),
					"init-persistent-home should be skipped when persistUserHome.enabled=false")
			}
		})

		It("Scenario 6: disableInitContainer=true — init-persistent-home from DWOC skipped", func() {
			By("Setting global config with disableInitContainer=true and init-persistent-home in InitContainers")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled:              ptr.To(true),
						DisableInitContainer: ptr.To(true),
					},
					InitContainers: []corev1.Container{
						{
							Name: constants.HomeInitComponentName,
							Args: []string{"echo 'custom persistent home init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createInitContainerTestDW(initContainerTestDWName, "test-devworkspace.yaml")
			devworkspace := getInitContainerTestDW(initContainerTestDWName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady("test-url", common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for Deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that init-persistent-home is NOT present in init containers")
			for _, initContainer := range deploy.Spec.Template.Spec.InitContainers {
				Expect(initContainer.Name).ShouldNot(Equal(constants.HomeInitComponentName),
					"init-persistent-home should be skipped when disableInitContainer=true")
			}
		})
	})
})
