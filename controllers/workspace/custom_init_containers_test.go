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
	"fmt"
	"net/http"
	"time"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	controllerv1alpha1 "github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	workspacecontroller "github.com/devfile/devworkspace-operator/controllers/workspace"
	"github.com/devfile/devworkspace-operator/controllers/workspace/internal/testutil"
	"github.com/devfile/devworkspace-operator/pkg/common"
	"github.com/devfile/devworkspace-operator/pkg/config"
	"github.com/devfile/devworkspace-operator/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	customInitTestURL = "http://custom-init-test-url"
)

// createAndWaitForDevWorkspace creates a DevWorkspace from a file and waits for it to get an ID.
// Unlike createDevWorkspace in util_test.go, this correctly uses the passed name for the Eventually check.
func createAndWaitForDevWorkspace(name, fromFile string) {
	By("Loading DevWorkspace from test file")
	devworkspace := &dw.DevWorkspace{}
	Expect(loadObjectFromFile(name, devworkspace, fromFile)).Should(Succeed())

	By("Creating DevWorkspace on cluster")
	Expect(k8sClient.Create(ctx, devworkspace)).Should(Succeed())

	By("Waiting for DevWorkspace to get an ID")
	createdDW := &dw.DevWorkspace{}
	Eventually(func() bool {
		if err := k8sClient.Get(ctx, namespacedName(name, testNamespace), createdDW); err != nil {
			return false
		}
		return createdDW.Status.DevWorkspaceId != ""
	}, 10*time.Second, 250*time.Millisecond).Should(BeTrue(), "DevWorkspace should get an ID")
}

// getExistingCustomDevWorkspace returns the DevWorkspace with the given name, waiting for it to have an ID.
func getExistingCustomDevWorkspace(name string) *dw.DevWorkspace {
	By(fmt.Sprintf("Getting existing DevWorkspace %s", name))
	devworkspace := &dw.DevWorkspace{}
	dwNN := namespacedName(name, testNamespace)
	Eventually(func() (string, error) {
		if err := k8sClient.Get(ctx, dwNN, devworkspace); err != nil {
			return "", err
		}
		return devworkspace.Status.DevWorkspaceId, nil
	}, timeout, interval).Should(Not(BeEmpty()))
	return devworkspace
}

var _ = Describe("Custom Init Containers", func() {
	const fixtureFile = "test-custom-init-container-workspace.yaml"
	const wsName = "test-custom-init-dw"

	BeforeEach(func() {
		workspacecontroller.SetupHttpClientsForTesting(&http.Client{
			Transport: &testutil.TestRoundTripper{
				Data: map[string]testutil.TestResponse{
					fmt.Sprintf("%s/healthz", customInitTestURL): {
						StatusCode: http.StatusOK,
					},
				},
			},
		})
	})

	AfterEach(func() {
		deleteDevWorkspace(wsName)
		config.SetGlobalConfigForTesting(nil)
		workspacecontroller.SetupHttpClientsForTesting(getBasicTestHttpClient())
	})

	Context("Scenario 1: DWOC with custom init-persistent-home (args only)", func() {
		It("Pod init containers include the custom args, not the default stow logic", func() {
			customArgs := "echo 'custom home init'"

			By("Configuring DWOC with custom init-persistent-home container (args only)")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(true),
					},
					InitContainers: []corev1.Container{
						{
							Name: constants.HomeInitComponentName,
							Args: []string{customArgs},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createAndWaitForDevWorkspace(wsName, fixtureFile)
			devworkspace := getExistingCustomDevWorkspace(wsName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady(customInitTestURL, common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that init-persistent-home container has custom args")
			initContainers := deploy.Spec.Template.Spec.InitContainers
			Expect(len(initContainers)).To(BeNumerically(">", 0), "No initContainers found in deployment")

			var homeInitContainer *corev1.Container
			for i := range initContainers {
				if initContainers[i].Name == constants.HomeInitComponentName {
					homeInitContainer = &initContainers[i]
					break
				}
			}
			Expect(homeInitContainer).NotTo(BeNil(), "init-persistent-home container should be present")

			By("Verifying custom args are used instead of default stow logic")
			Expect(homeInitContainer.Args).To(Equal([]string{customArgs}),
				"init-persistent-home should use custom args from DWOC")

			By("Verifying command is set to [/bin/sh, -c] by EnsureHomeInitContainerFields")
			Expect(homeInitContainer.Command).To(Equal([]string{"/bin/sh", "-c"}),
				"command should be set to [/bin/sh, -c] by EnsureHomeInitContainerFields")
		})
	})

	Context("Scenario 2: DWOC with additional non-init-persistent-home init container", func() {
		It("Pod init containers include both the custom container AND init-persistent-home", func() {
			By("Configuring DWOC with init-persistent-home and an extra init container")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(true),
					},
					InitContainers: []corev1.Container{
						{
							Name: constants.HomeInitComponentName,
							Args: []string{"echo 'custom home init'"},
						},
						{
							Name:    "extra-init-container",
							Image:   "busybox:latest",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{"echo 'extra init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			createAndWaitForDevWorkspace(wsName, fixtureFile)
			devworkspace := getExistingCustomDevWorkspace(wsName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady(customInitTestURL, common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking both init containers are present")
			initContainers := deploy.Spec.Template.Spec.InitContainers
			Expect(len(initContainers)).To(BeNumerically(">=", 2), "Should have at least 2 init containers")

			containerNames := make([]string, 0, len(initContainers))
			for _, c := range initContainers {
				containerNames = append(containerNames, c.Name)
			}
			Expect(containerNames).To(ContainElement(constants.HomeInitComponentName),
				"init-persistent-home container should be present")
			Expect(containerNames).To(ContainElement("extra-init-container"),
				"extra-init-container should be present")
		})
	})

	Context("Scenario 3: DWOC init-persistent-home with invalid command", func() {
		It("Workspace reaches Failed status with message about invalid command", func() {
			By("Configuring DWOC with init-persistent-home having invalid command [/bin/bash, -c]")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(true),
					},
					InitContainers: []corev1.Container{
						{
							Name:    constants.HomeInitComponentName,
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"echo 'custom home init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace")
			devworkspace := &dw.DevWorkspace{}
			Expect(loadObjectFromFile(wsName, devworkspace, fixtureFile)).Should(Succeed())
			Expect(k8sClient.Create(ctx, devworkspace)).Should(Succeed())
			dwNamespacedName := namespacedName(wsName, testNamespace)

			By("Waiting for DevWorkspace to enter Failed phase")
			currDW := &dw.DevWorkspace{}
			Eventually(func() (dw.DevWorkspacePhase, error) {
				if err := k8sClient.Get(ctx, dwNamespacedName, currDW); err != nil {
					return "", err
				}
				GinkgoWriter.Printf("Waiting for Failed phase -- Phase: %s, Message: %s\n",
					currDW.Status.Phase, currDW.Status.Message)
				return currDW.Status.Phase, nil
			}, timeout, interval).Should(Equal(dw.DevWorkspaceStatusFailed),
				"Workspace should enter Failed phase due to invalid init-persistent-home command")

			By("Verifying the failure message contains the expected text")
			Expect(currDW.Status.Message).To(ContainSubstring("Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]"),
				"Failure message should describe the invalid command")
		})
	})

	Context("Scenario 4: No DWOC init containers (backward compatibility)", func() {
		It("init-persistent-home is still present from default logic", func() {
			By("Configuring DWOC with PersistUserHome enabled but no custom InitContainers")
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					PersistUserHome: &controllerv1alpha1.PersistentHomeConfig{
						Enabled: ptr.To(true),
					},
				},
			})

			By("Creating DevWorkspace")
			createAndWaitForDevWorkspace(wsName, fixtureFile)
			devworkspace := getExistingCustomDevWorkspace(wsName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady(customInitTestURL, common.DevWorkspaceRoutingName(workspaceID))

			By("Waiting for deployment to be created")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Eventually(func() error {
				return k8sClient.Get(ctx, deployNN, deploy)
			}, timeout, interval).Should(Succeed(), "Getting workspace deployment from cluster")

			By("Checking that init-persistent-home is present from default logic")
			initContainers := deploy.Spec.Template.Spec.InitContainers
			Expect(len(initContainers)).To(BeNumerically(">", 0), "Should have at least one init container")

			var homeInitContainer *corev1.Container
			for i := range initContainers {
				if initContainers[i].Name == constants.HomeInitComponentName {
					homeInitContainer = &initContainers[i]
					break
				}
			}
			Expect(homeInitContainer).NotTo(BeNil(),
				"init-persistent-home container should be present from default logic")

			By("Verifying the default stow script is used (not custom args)")
			// The default logic uses /bin/sh -c with the stow script
			Expect(homeInitContainer.Command).To(Equal([]string{"/bin/sh", "-c"}),
				"Default init-persistent-home should have /bin/sh -c command")
			Expect(homeInitContainer.Args).NotTo(BeEmpty(),
				"Default init-persistent-home should have args (the stow script)")
		})
	})

	Context("Scenario 5: DWOC with InitContainers but PersistUserHome not set", func() {
		It("Workspace reconciles without panicking and reaches Running status", func() {
			By("Configuring DWOC with InitContainers but no PersistUserHome field")
			// When PersistUserHome is not set in the custom config, the default config's
			// PersistUserHome (Enabled: false) is preserved after merging.
			// This tests that the nil-guard added in the controller works correctly.
			config.SetGlobalConfigForTesting(&controllerv1alpha1.OperatorConfiguration{
				Workspace: &controllerv1alpha1.WorkspaceConfig{
					// PersistUserHome intentionally omitted (nil in custom config)
					InitContainers: []corev1.Container{
						{
							Name:    "extra-init-container",
							Image:   "busybox:latest",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{"echo 'extra init'"},
						},
					},
				},
			})

			By("Creating DevWorkspace with ephemeral storage to avoid per-workspace PVC issues")
			// Use ephemeral storage since PersistUserHome is disabled and we don't need a PVC.
			// Ephemeral storage avoids the per-workspace PVC size calculation issue.
			ephemeralWorkspace := &dw.DevWorkspace{}
			Expect(loadObjectFromFile(wsName, ephemeralWorkspace, fixtureFile)).Should(Succeed())
			// Override storage type to ephemeral
			ephemeralWorkspace.Spec.Template.Attributes.PutString(
				constants.DevWorkspaceStorageTypeAttribute,
				constants.EphemeralStorageClassType,
			)
			Expect(k8sClient.Create(ctx, ephemeralWorkspace)).Should(Succeed())

			By("Waiting for DevWorkspace to get an ID")
			createdDW := &dw.DevWorkspace{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, namespacedName(wsName, testNamespace), createdDW); err != nil {
					return false
				}
				return createdDW.Status.DevWorkspaceId != ""
			}, 10*time.Second, 250*time.Millisecond).Should(BeTrue(), "DevWorkspace should get an ID")

			devworkspace := getExistingCustomDevWorkspace(wsName)
			workspaceID := devworkspace.Status.DevWorkspaceId

			By("Manually making Routing ready to continue")
			markRoutingReady(customInitTestURL, common.DevWorkspaceRoutingName(workspaceID))

			By("Setting the deployment to have 1 ready replica")
			markDeploymentReady(common.DeploymentName(workspaceID))

			By("Verifying workspace reaches Running status without panicking")
			currDW := &dw.DevWorkspace{}
			Eventually(func() (dw.DevWorkspacePhase, error) {
				if err := k8sClient.Get(ctx, namespacedName(wsName, testNamespace), currDW); err != nil {
					return "", err
				}
				GinkgoWriter.Printf("Waiting for Running phase -- Phase: %s, Message: %s\n",
					currDW.Status.Phase, currDW.Status.Message)
				return currDW.Status.Phase, nil
			}, timeout, interval).Should(Equal(dw.DevWorkspaceStatusRunning),
				"Workspace should reach Running status without panicking")

			By("Verifying the extra init container is injected")
			deploy := &appsv1.Deployment{}
			deployNN := namespacedName(common.DeploymentName(workspaceID), testNamespace)
			Expect(k8sClient.Get(ctx, deployNN, deploy)).Should(Succeed())
			containerNames := make([]string, 0)
			for _, c := range deploy.Spec.Template.Spec.InitContainers {
				containerNames = append(containerNames, c.Name)
			}
			Expect(containerNames).To(ContainElement("extra-init-container"),
				"extra-init-container should be injected even when PersistUserHome is not configured")
		})
	})
})
