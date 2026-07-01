//
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
//

package home

import (
	"testing"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	"github.com/devfile/devworkspace-operator/pkg/common"
	"github.com/devfile/devworkspace-operator/pkg/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestCustomInitPersistentHome(t *testing.T) {
	tests := []struct {
		name                    string
		workspace               *common.DevWorkspaceWithConfig
		expectDefaultInitAdded  bool
		expectCustomInitSkipped bool
	}{
		{
			name: "Adds default init when custom init-persistent-home is provided",
			workspace: &common.DevWorkspaceWithConfig{
				DevWorkspace: &dw.DevWorkspace{
					Spec: dw.DevWorkspaceSpec{
						Template: dw.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
								Components: []dw.Component{
									{
										Name: "test-container",
										ComponentUnion: dw.ComponentUnion{
											Container: &dw.ContainerComponent{
												Container: dw.Container{
													Image: "test-image:latest",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Config: &v1alpha1.OperatorConfiguration{
					Workspace: &v1alpha1.WorkspaceConfig{
						PersistUserHome: &v1alpha1.PersistentHomeConfig{
							Enabled: ptr.To(true),
						},
						InitContainers: []corev1.Container{
							{
								Name: constants.HomeInitComponentName,
								Args: []string{"echo 'custom init'"},
							},
						},
					},
				},
			},
			expectDefaultInitAdded: true,
		},
		{
			name: "Adds default init when no custom init-persistent-home is provided",
			workspace: &common.DevWorkspaceWithConfig{
				DevWorkspace: &dw.DevWorkspace{
					Spec: dw.DevWorkspaceSpec{
						Template: dw.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
								Components: []dw.Component{
									{
										Name: "test-container",
										ComponentUnion: dw.ComponentUnion{
											Container: &dw.ContainerComponent{
												Container: dw.Container{
													Image: "test-image:latest",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Config: &v1alpha1.OperatorConfiguration{
					Workspace: &v1alpha1.WorkspaceConfig{
						PersistUserHome: &v1alpha1.PersistentHomeConfig{
							Enabled: ptr.To(true),
						},
						InitContainers: []corev1.Container{
							{
								Name:  "custom-container",
								Image: "custom:latest",
								Args:  []string{"echo 'other init'"},
							},
						},
					},
				},
			},
			expectDefaultInitAdded: true,
		},
		{
			name: "Adds default init when custom init containers list is empty",
			workspace: &common.DevWorkspaceWithConfig{
				DevWorkspace: &dw.DevWorkspace{
					Spec: dw.DevWorkspaceSpec{
						Template: dw.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
								Components: []dw.Component{
									{
										Name: "test-container",
										ComponentUnion: dw.ComponentUnion{
											Container: &dw.ContainerComponent{
												Container: dw.Container{
													Image: "test-image:latest",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Config: &v1alpha1.OperatorConfiguration{
					Workspace: &v1alpha1.WorkspaceConfig{
						PersistUserHome: &v1alpha1.PersistentHomeConfig{
							Enabled:              ptr.To(true),
							DisableInitContainer: ptr.To(false),
						},
						InitContainers: []corev1.Container{},
					},
				},
			},
			expectDefaultInitAdded: true,
		},
		{
			name: "Skips default init when DisableInitContainer is true even with custom init",
			workspace: &common.DevWorkspaceWithConfig{
				DevWorkspace: &dw.DevWorkspace{
					Spec: dw.DevWorkspaceSpec{
						Template: dw.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
								Components: []dw.Component{
									{
										Name: "test-container",
										ComponentUnion: dw.ComponentUnion{
											Container: &dw.ContainerComponent{
												Container: dw.Container{
													Image: "test-image:latest",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Config: &v1alpha1.OperatorConfiguration{
					Workspace: &v1alpha1.WorkspaceConfig{
						PersistUserHome: &v1alpha1.PersistentHomeConfig{
							Enabled:              ptr.To(true),
							DisableInitContainer: ptr.To(true),
						},
					},
				},
			},
			expectDefaultInitAdded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddPersistentHomeVolume(tt.workspace)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check if default init component was added
			hasDefaultInit := false
			for _, component := range result.Components {
				if component.Name == constants.HomeInitComponentName {
					hasDefaultInit = true
					break
				}
			}

			if tt.expectDefaultInitAdded {
				assert.True(t, hasDefaultInit, "Expected default init component to be added")
			} else {
				assert.False(t, hasDefaultInit, "Expected default init component NOT to be added")
			}

			// Verify persistent-home volume is always added
			hasPersistentHomeVolume := false
			for _, component := range result.Components {
				if component.Name == constants.HomeVolumeName {
					hasPersistentHomeVolume = true
					break
				}
			}
			assert.True(t, hasPersistentHomeVolume, "persistent-home volume should always be added")
		})
	}
}

func TestInferWorkspaceImage(t *testing.T) {
	tests := []struct {
		name          string
		template      *dw.DevWorkspaceTemplateSpec
		expectedImage string
	}{
		{
			name: "Returns first non-imported container image",
			template: &dw.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
					Components: []dw.Component{
						{
							Name: "main-container",
							ComponentUnion: dw.ComponentUnion{
								Container: &dw.ContainerComponent{
									Container: dw.Container{
										Image: "my-workspace:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedImage: "my-workspace:latest",
		},
		{
			name: "Skips imported containers and returns first non-imported",
			template: &dw.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
					Components: []dw.Component{
						{
							Name: "plugin-container",
							Attributes: attributes.Attributes{}.
								PutString(constants.PluginSourceAttribute, "plugin-registry"),
							ComponentUnion: dw.ComponentUnion{
								Container: &dw.ContainerComponent{
									Container: dw.Container{
										Image: "plugin-image:latest",
									},
								},
							},
						},
						{
							Name: "main-container",
							ComponentUnion: dw.ComponentUnion{
								Container: &dw.ContainerComponent{
									Container: dw.Container{
										Image: "my-workspace:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedImage: "my-workspace:latest",
		},
		{
			name: "Returns empty string when no suitable container found",
			template: &dw.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
					Components: []dw.Component{
						{
							Name: "volume",
							ComponentUnion: dw.ComponentUnion{
								Volume: &dw.VolumeComponent{},
							},
						},
					},
				},
			},
			expectedImage: "",
		},
		{
			name: "Returns empty string when all containers are imported",
			template: &dw.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
					Components: []dw.Component{
						{
							Name: "plugin-container",
							Attributes: attributes.Attributes{}.
								PutString(constants.PluginSourceAttribute, "plugin-registry"),
							ComponentUnion: dw.ComponentUnion{
								Container: &dw.ContainerComponent{
									Container: dw.Container{
										Image: "plugin-image:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedImage: "",
		},
		{
			name: "Treats parent-sourced containers as non-imported",
			template: &dw.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
					Components: []dw.Component{
						{
							Name: "parent-container",
							Attributes: attributes.Attributes{}.
								PutString(constants.PluginSourceAttribute, "parent"),
							ComponentUnion: dw.ComponentUnion{
								Container: &dw.ContainerComponent{
									Container: dw.Container{
										Image: "parent-image:latest",
									},
								},
							},
						},
					},
				},
			},
			expectedImage: "parent-image:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferWorkspaceImage(tt.template)
			assert.Equal(t, tt.expectedImage, result)
		})
	}
}

// TestEnsureHomeInitContainerFields tests the EnsureHomeInitContainerFields function.
// Note: No YAML fixture is created in testdata/persistent-home/ for this function because
// EnsureHomeInitContainerFields is NOT on the AddPersistentHomeVolume call path — it is called
// later in the controller after MergeInitContainers. The YAML-driven tests in persistentHome_test.go
// exercise NeedsPersistentHomeDirectory + AddPersistentHomeVolume and would never invoke this
// validation logic. Creating an error-type fixture would cause TestPersistentHomeVolume to assert
// an error from AddPersistentHomeVolume that is never returned, breaking existing tests.
func TestEnsureHomeInitContainerFields(t *testing.T) {
	tests := []struct {
		name          string
		inputCommand  []string
		expectErr     bool
		errContains   string
		expectCommand []string
	}{
		{
			name:          "Empty command sets default [/bin/sh, -c], no error",
			inputCommand:  []string{},
			expectErr:     false,
			expectCommand: []string{"/bin/sh", "-c"},
		},
		{
			name:          "Valid command [/bin/sh, -c] explicitly provided, no change",
			inputCommand:  []string{"/bin/sh", "-c"},
			expectErr:     false,
			expectCommand: []string{"/bin/sh", "-c"},
		},
		{
			name:         "Invalid command [/bin/bash, -c] returns error",
			inputCommand: []string{"/bin/bash", "-c"},
			expectErr:    true,
			errContains:  "Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]",
		},
		{
			name:         "Invalid command with single element [/bin/sh] returns error",
			inputCommand: []string{"/bin/sh"},
			expectErr:    true,
			errContains:  "Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &corev1.Container{
				Name:    constants.HomeInitComponentName,
				Command: tt.inputCommand,
			}

			err := EnsureHomeInitContainerFields(container)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectCommand, container.Command)
				// VolumeMounts are always set to persistent-home -> /home/user/ for non-error cases
				assert.Len(t, container.VolumeMounts, 1)
				assert.Equal(t, constants.HomeVolumeName, container.VolumeMounts[0].Name)
				assert.Equal(t, constants.HomeUserDirectory, container.VolumeMounts[0].MountPath)
			}
		})
	}
}
