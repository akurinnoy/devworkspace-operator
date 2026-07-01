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

func TestEnsureHomeInitContainerFields(t *testing.T) {
	baseWorkspace := &common.DevWorkspaceWithConfig{
		DevWorkspace: &dw.DevWorkspace{
			Spec: dw.DevWorkspaceSpec{
				Template: dw.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
						Components: []dw.Component{
							{
								Name: "main-container",
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
			},
		},
	}

	tests := []struct {
		name            string
		inputCommand    []string
		expectError     bool
		expectedCommand []string
	}{
		{
			name:            "no command — succeeds, command becomes [/bin/sh, -c]",
			inputCommand:    nil,
			expectError:     false,
			expectedCommand: []string{"/bin/sh", "-c"},
		},
		{
			name:            "command [/bin/sh, -c] explicitly — succeeds",
			inputCommand:    []string{"/bin/sh", "-c"},
			expectError:     false,
			expectedCommand: []string{"/bin/sh", "-c"},
		},
		{
			name:         "command [/bin/bash, -c] — validation error",
			inputCommand: []string{"/bin/bash", "-c"},
			expectError:  true,
		},
		{
			name:         "command [/bin/sh] missing -c — validation error",
			inputCommand: []string{"/bin/sh"},
			expectError:  true,
		},
		{
			name:         "command [/bin/sh, -c, extra] too many — validation error",
			inputCommand: []string{"/bin/sh", "-c", "extra"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &corev1.Container{
				Name:    constants.HomeInitComponentName,
				Image:   "test-image:latest",
				Command: tt.inputCommand,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "existing-volume",
						MountPath: "/existing",
					},
				},
			}

			err := EnsureHomeInitContainerFields(container, baseWorkspace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "Invalid init-persistent-home container")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCommand, container.Command)
				// VolumeMounts must always be overwritten
				assert.Len(t, container.VolumeMounts, 1)
				assert.Equal(t, constants.HomeVolumeName, container.VolumeMounts[0].Name)
				assert.Equal(t, constants.HomeUserDirectory, container.VolumeMounts[0].MountPath)
			}
		})
	}

	t.Run("VolumeMounts are always overwritten even with valid command", func(t *testing.T) {
		container := &corev1.Container{
			Name:    constants.HomeInitComponentName,
			Image:   "test-image:latest",
			Command: []string{"/bin/sh", "-c"},
			VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1", MountPath: "/mnt/vol1"},
				{Name: "vol2", MountPath: "/mnt/vol2"},
			},
		}

		err := EnsureHomeInitContainerFields(container, baseWorkspace)
		assert.NoError(t, err)
		assert.Len(t, container.VolumeMounts, 1, "VolumeMounts should be overwritten to only contain persistent-home")
		assert.Equal(t, constants.HomeVolumeName, container.VolumeMounts[0].Name)
		assert.Equal(t, constants.HomeUserDirectory, container.VolumeMounts[0].MountPath)
	})
}

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

func TestEnsureHomeInitContainerFieldsImageInheritance(t *testing.T) {
	workspaceImage := "workspace-image:latest"

	makeWorkspace := func(image string) *common.DevWorkspaceWithConfig {
		return &common.DevWorkspaceWithConfig{
			DevWorkspace: &dw.DevWorkspace{
				Spec: dw.DevWorkspaceSpec{
					Template: dw.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
							Components: []dw.Component{
								{
									Name: "main-container",
									ComponentUnion: dw.ComponentUnion{
										Container: &dw.ContainerComponent{
											Container: dw.Container{
												Image: image,
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
				},
			},
		}
	}

	t.Run("init-persistent-home container with no image inherits workspace image", func(t *testing.T) {
		workspace := makeWorkspace(workspaceImage)
		container := &corev1.Container{
			Name:    constants.HomeInitComponentName,
			Image:   "",
			Command: []string{"/bin/sh", "-c"},
		}

		err := EnsureHomeInitContainerFields(container, workspace)
		assert.NoError(t, err)
		// Image should be inferred from the workspace's primary container
		inferredImage := InferWorkspaceImage(&workspace.Spec.Template)
		assert.Equal(t, inferredImage, container.Image, "empty image should be replaced with inferred workspace image")
		assert.Equal(t, workspaceImage, container.Image)
	})

	t.Run("init-persistent-home container with image already set preserves it", func(t *testing.T) {
		workspace := makeWorkspace(workspaceImage)
		customImage := "custom-init-image:v1"
		container := &corev1.Container{
			Name:    constants.HomeInitComponentName,
			Image:   customImage,
			Command: []string{"/bin/sh", "-c"},
		}

		err := EnsureHomeInitContainerFields(container, workspace)
		assert.NoError(t, err)
		// Pre-set image must not be overwritten
		assert.Equal(t, customImage, container.Image, "non-empty image should be preserved as-is")
	})

	t.Run("non-init-persistent-home custom container image is not touched", func(t *testing.T) {
		workspace := makeWorkspace(workspaceImage)
		originalImage := "other-init:latest"
		// EnsureHomeInitContainerFields does not filter by container name.
		// When the container already has an image, it must be preserved regardless of name.
		container := &corev1.Container{
			Name:    "other-custom-init",
			Image:   originalImage,
			Command: []string{"/bin/sh", "-c"},
		}

		err := EnsureHomeInitContainerFields(container, workspace)
		assert.NoError(t, err)
		assert.Equal(t, originalImage, container.Image, "image of a non-init-persistent-home container should not be changed")
	})
}
