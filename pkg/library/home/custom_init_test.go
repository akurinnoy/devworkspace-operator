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

func TestEnsureHomeInitContainerFields(t *testing.T) {
	tests := []struct {
		name        string
		input       corev1.Container
		wantErr     bool
		errMsg      string
		wantCommand []string
		wantMounts  []corev1.VolumeMount
	}{
		{
			name: "Empty command gets default command set",
			input: corev1.Container{
				Name:  constants.HomeInitComponentName,
				Image: "workspace:latest",
			},
			wantErr:     false,
			wantCommand: []string{"/bin/sh", "-c"},
			wantMounts: []corev1.VolumeMount{
				{Name: constants.HomeVolumeName, MountPath: constants.HomeUserDirectory},
			},
		},
		{
			name: "Command /bin/sh -c is accepted unchanged",
			input: corev1.Container{
				Name:    constants.HomeInitComponentName,
				Image:   "workspace:latest",
				Command: []string{"/bin/sh", "-c"},
				Args:    []string{"echo hello"},
			},
			wantErr:     false,
			wantCommand: []string{"/bin/sh", "-c"},
			wantMounts: []corev1.VolumeMount{
				{Name: constants.HomeVolumeName, MountPath: constants.HomeUserDirectory},
			},
		},
		{
			name: "Command /bin/bash -c returns error with exact message",
			input: corev1.Container{
				Name:    constants.HomeInitComponentName,
				Image:   "workspace:latest",
				Command: []string{"/bin/bash", "-c"},
			},
			wantErr: true,
			errMsg:  "Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]",
		},
		{
			name: "Command /bin/sh without -c returns error with exact message",
			input: corev1.Container{
				Name:    constants.HomeInitComponentName,
				Image:   "workspace:latest",
				Command: []string{"/bin/sh"},
			},
			wantErr: true,
			errMsg:  "Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]",
		},
		{
			name: "VolumeMounts set to persistent-home on /home/user/ for empty command container",
			input: corev1.Container{
				Name:  constants.HomeInitComponentName,
				Image: "workspace:latest",
			},
			wantErr: false,
			wantMounts: []corev1.VolumeMount{
				{Name: "persistent-home", MountPath: "/home/user/"},
			},
		},
		{
			name: "VolumeMounts set to persistent-home on /home/user/ for valid command container",
			input: corev1.Container{
				Name:    constants.HomeInitComponentName,
				Image:   "workspace:latest",
				Command: []string{"/bin/sh", "-c"},
			},
			wantErr: false,
			wantMounts: []corev1.VolumeMount{
				{Name: "persistent-home", MountPath: "/home/user/"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.input.DeepCopy()
			err := EnsureHomeInitContainerFields(c)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
				if tt.wantCommand != nil {
					assert.Equal(t, tt.wantCommand, c.Command)
				}
				if tt.wantMounts != nil {
					assert.Equal(t, tt.wantMounts, c.VolumeMounts)
				}
			}
		})
	}
}

func makeWorkspaceWithDWOCContainers(initContainers []corev1.Container) *common.DevWorkspaceWithConfig {
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
											Image: "workspace-image:latest",
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
				InitContainers: initContainers,
			},
		},
	}
}

func TestDWOCMultipleInitContainersFullFlow(t *testing.T) {
	tests := []struct {
		name               string
		dwocInitContainers []corev1.Container
		expectNames        []string // expected container names after merge, in order
	}{
		{
			name: "Two non-init-persistent-home DWOC init containers appear after init-persistent-home",
			dwocInitContainers: []corev1.Container{
				{
					Name:  "tool-init-a",
					Image: "tool-a:latest",
				},
				{
					Name:  "tool-init-b",
					Image: "tool-b:latest",
				},
			},
			// init-persistent-home is added by AddPersistentHomeVolume;
			// DWOC containers are new-named patches so they are appended after
			expectNames: []string{constants.HomeInitComponentName, "tool-init-a", "tool-init-b"},
		},
		{
			name: "Order of DWOC init containers is preserved in merged output",
			dwocInitContainers: []corev1.Container{
				{
					Name:  "step-1",
					Image: "step-1:latest",
				},
				{
					Name:  "step-2",
					Image: "step-2:latest",
				},
				{
					Name:  "step-3",
					Image: "step-3:latest",
				},
			},
			expectNames: []string{constants.HomeInitComponentName, "step-1", "step-2", "step-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := makeWorkspaceWithDWOCContainers(tt.dwocInitContainers)

			// Step 1: AddPersistentHomeVolume produces the devfile-level template spec
			// with init-persistent-home component added
			spec, err := AddPersistentHomeVolume(workspace)
			assert.NoError(t, err)
			assert.NotNil(t, spec)

			// Verify init-persistent-home component is present
			var devfileInitContainers []corev1.Container
			for _, component := range spec.Components {
				if component.Name == constants.HomeInitComponentName && component.Container != nil {
					devfileInitContainers = append(devfileInitContainers, corev1.Container{
						Name:    component.Name,
						Image:   component.Container.Image,
						Command: component.Container.Command,
						Args:    component.Container.Args,
					})
				}
			}
			assert.Len(t, devfileInitContainers, 1, "expected exactly one init-persistent-home component in spec")

			// Step 2: Simulate controller merge: base = devfile init containers,
			// patches = DWOC init containers (new-named containers are appended in order)
			merged := append([]corev1.Container{}, devfileInitContainers...)
			baseNames := make(map[string]bool)
			for _, c := range devfileInitContainers {
				baseNames[c.Name] = true
			}
			for _, patch := range tt.dwocInitContainers {
				if !baseNames[patch.Name] {
					merged = append(merged, patch)
				}
			}

			// Verify count and order
			assert.Len(t, merged, len(tt.expectNames))
			for i, name := range tt.expectNames {
				if i < len(merged) {
					assert.Equal(t, name, merged[i].Name, "container at index %d should be %q", i, name)
				}
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
