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

func TestCustomInitEnsureHomeInitContainerFields(t *testing.T) {
	makeWorkspaceWithImage := func(image string) *common.DevWorkspaceWithConfig {
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

	tests := []struct {
		name          string
		container     *corev1.Container
		workspace     *common.DevWorkspaceWithConfig
		wantErr       bool
		wantErrSubstr string
		wantCommand   []string
		wantMountPath string
		wantMountName string
		wantImage     string
	}{
		{
			name: "nil command is set to [/bin/sh, -c]",
			container: &corev1.Container{
				Name:    "init-persistent-home",
				Command: nil,
				Image:   "workspace:latest",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       false,
			wantCommand:   []string{"/bin/sh", "-c"},
			wantMountPath: constants.HomeUserDirectory,
			wantMountName: constants.HomeVolumeName,
			wantImage:     "workspace:latest",
		},
		{
			name: "empty command is set to [/bin/sh, -c]",
			container: &corev1.Container{
				Name:    "init-persistent-home",
				Command: []string{},
				Image:   "workspace:latest",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       false,
			wantCommand:   []string{"/bin/sh", "-c"},
			wantMountPath: constants.HomeUserDirectory,
			wantMountName: constants.HomeVolumeName,
			wantImage:     "workspace:latest",
		},
		{
			name: "correct command [/bin/sh, -c] is accepted without error",
			container: &corev1.Container{
				Name:    "init-persistent-home",
				Command: []string{"/bin/sh", "-c"},
				Image:   "workspace:latest",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       false,
			wantCommand:   []string{"/bin/sh", "-c"},
			wantMountPath: constants.HomeUserDirectory,
			wantMountName: constants.HomeVolumeName,
			wantImage:     "workspace:latest",
		},
		{
			name: "wrong command [bash] returns error",
			container: &corev1.Container{
				Name:    "init-persistent-home",
				Command: []string{"bash"},
				Image:   "workspace:latest",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       true,
			wantErrSubstr: "command must be exactly [/bin/sh, -c]",
		},
		{
			name: "wrong command [/bin/sh] (missing -c) returns error",
			container: &corev1.Container{
				Name:    "init-persistent-home",
				Command: []string{"/bin/sh"},
				Image:   "workspace:latest",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       true,
			wantErrSubstr: "command must be exactly [/bin/sh, -c]",
		},
		{
			name: "volumeMounts are always overridden with persistent-home at /home/user",
			container: &corev1.Container{
				Name:  "init-persistent-home",
				Image: "workspace:latest",
				VolumeMounts: []corev1.VolumeMount{
					{Name: "some-other-volume", MountPath: "/data"},
					{Name: "another-volume", MountPath: "/tmp/other"},
				},
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       false,
			wantCommand:   []string{"/bin/sh", "-c"},
			wantMountPath: constants.HomeUserDirectory,
			wantMountName: constants.HomeVolumeName,
			wantImage:     "workspace:latest",
		},
		{
			name: "empty image is inferred from workspace",
			container: &corev1.Container{
				Name:  "init-persistent-home",
				Image: "",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       false,
			wantCommand:   []string{"/bin/sh", "-c"},
			wantMountPath: constants.HomeUserDirectory,
			wantMountName: constants.HomeVolumeName,
			wantImage:     "workspace-inferred:latest",
		},
		{
			name: "non-empty image is preserved and not overridden",
			container: &corev1.Container{
				Name:  "init-persistent-home",
				Image: "custom:tag",
			},
			workspace:     makeWorkspaceWithImage("workspace-inferred:latest"),
			wantErr:       false,
			wantCommand:   []string{"/bin/sh", "-c"},
			wantMountPath: constants.HomeUserDirectory,
			wantMountName: constants.HomeVolumeName,
			wantImage:     "custom:tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureHomeInitContainerFields(tt.container, tt.workspace)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrSubstr != "" {
					assert.Contains(t, err.Error(), tt.wantErrSubstr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCommand, tt.container.Command, "Command should match expected value")
			assert.Equal(t, tt.wantImage, tt.container.Image, "Image should match expected value")
			assert.Len(t, tt.container.VolumeMounts, 1, "VolumeMounts should be overridden to exactly one entry")
			if len(tt.container.VolumeMounts) == 1 {
				assert.Equal(t, tt.wantMountName, tt.container.VolumeMounts[0].Name, "VolumeMount name should be persistent-home")
				assert.Equal(t, tt.wantMountPath, tt.container.VolumeMounts[0].MountPath, "VolumeMount path should be /home/user/")
			}
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
