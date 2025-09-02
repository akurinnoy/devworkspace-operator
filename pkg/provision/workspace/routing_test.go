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

package workspace

import (
	"context"
	"testing"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/devworkspace-operator/pkg/common"
	"github.com/devfile/devworkspace-operator/pkg/constants"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCheckRoutingConflicts(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = dw.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	log := logr.Discard()

	tests := []struct {
		name        string
		workspace   *common.DevWorkspaceWithConfig
		existing    []runtime.Object
		expectErr   bool
		expectedMsg string
	}{
		{
			name:      "No conflicts",
			workspace: testWorkspace("test-ws", "test-ns", "test-uid", []string{"endpoint1"}),
			existing:  []runtime.Object{},
			expectErr: false,
		},
		{
			name:      "Conflict with another running workspace",
			workspace: testWorkspace("test-ws", "test-ns", "test-uid", []string{"endpoint1"}),
			existing: []runtime.Object{
				testWorkspace("other-ws", "test-ns", "other-uid", []string{"endpoint1"}).DevWorkspace,
			},
			expectErr:   true,
			expectedMsg: "Endpoint name 'endpoint1' conflicts with an active workspace 'other-ws' in the same namespace. Please choose a different endpoint name.",
		},
		{
			name:      "Conflict with an existing service",
			workspace: testWorkspace("test-ws", "test-ns", "test-uid", []string{"endpoint1"}),
			existing: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "endpoint1",
						Namespace: "test-ns",
					},
				},
			},
			expectErr:   true,
			expectedMsg: "service 'endpoint1' already exists in this namespace and is not owned by this workspace, this may indicate an endpoint name conflict, please choose a different endpoint name",
		},
		{
			name:      "No conflict with service owned by the same workspace",
			workspace: testWorkspace("test-ws", "test-ns", "test-uid", []string{"endpoint1"}),
			existing: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "endpoint1",
						Namespace: "test-ns",
						Labels: map[string]string{
							constants.DevWorkspaceIDLabel: "test-ws-id",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "No conflict with workspace in another namespace",
			workspace: testWorkspace("test-ws", "test-ns", "test-uid", []string{"endpoint1"}),
			existing: []runtime.Object{
				testWorkspace("other-ws", "other-ns", "other-uid", []string{"endpoint1"}).DevWorkspace,
			},
			expectErr: false,
		},
		{
			name:      "No conflict with stopped workspace",
			workspace: testWorkspace("test-ws", "test-ns", "test-uid", []string{"endpoint1"}),
			existing: []runtime.Object{
				func() *dw.DevWorkspace {
					ws := testWorkspace("other-ws", "test-ns", "other-uid", []string{"endpoint1"}).DevWorkspace
					ws.Status.Phase = dw.DevWorkspaceStatusStopped
					return ws
				}(),
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.existing...).Build()
			err := checkRoutingConflicts(context.Background(), fakeClient, tt.workspace, log)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func testWorkspace(name, namespace, uid string, endpoints []string) *common.DevWorkspaceWithConfig {
	dwEndpoints := []dw.Endpoint{}
	for _, e := range endpoints {
		dwEndpoints = append(dwEndpoints, dw.Endpoint{
			Name:       e,
			TargetPort: 8080,
			Exposure:   dw.PublicEndpointExposure,
		})
	}
	return &common.DevWorkspaceWithConfig{
		DevWorkspace: &dw.DevWorkspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       types.UID(uid),
			},
			Spec: dw.DevWorkspaceSpec{
				Template: dw.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: dw.DevWorkspaceTemplateSpecContent{
						Components: []dw.Component{
							{
								Name: "test-component",
								ComponentUnion: dw.ComponentUnion{
									Container: &dw.ContainerComponent{
										Endpoints: dwEndpoints,
									},
								},
							},
						},
					},
				},
			},
			Status: dw.DevWorkspaceStatus{
				Phase:          dw.DevWorkspaceStatusRunning,
				DevWorkspaceId: name + "-id",
			},
		},
	}
}
