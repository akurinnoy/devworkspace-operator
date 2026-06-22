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

package solvers

import (
	"context"
	"errors"
	"testing"

	controllerv1alpha1 "github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	"github.com/devfile/devworkspace-operator/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// discoverableEndpoint returns an Endpoint with the discoverable attribute set to true.
func discoverableEndpoint(name string, port int) controllerv1alpha1.Endpoint {
	attrs := controllerv1alpha1.Attributes{}
	attrs.PutBoolean(string(controllerv1alpha1.DiscoverableAttribute), true)
	return controllerv1alpha1.Endpoint{
		Name:       name,
		Exposure:   controllerv1alpha1.InternalEndpointExposure,
		TargetPort: port,
		Attributes: attrs,
	}
}

// makeScheme returns a runtime.Scheme that includes the corev1 types (required by the fake client
// for listing Services).
func makeScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))
	return s
}

// DeferCleanup registers a cleanup function to run when the test and all its sub-tests complete.
// This mirrors the Ginkgo DeferCleanup pattern for plain testing.T.
func DeferCleanup(t *testing.T, fn func()) {
	t.Helper()
	t.Cleanup(fn)
}

// TestGetDiscoverableServicesForEndpoints_NoConflict_SingleWorkspace verifies that a single
// DevWorkspaceRouting with a discoverable endpoint reconciles without error and the returned
// service list contains the expected service.
func TestGetDiscoverableServicesForEndpoints_NoConflict_SingleWorkspace(t *testing.T) {
	ctx := context.Background()
	scheme := makeScheme(t)
	clnt := fake.NewClientBuilder().WithScheme(scheme).Build()
	// Each test uses an isolated fake client; DeferCleanup is a no-op here since there is
	// no shared state, but is registered for symmetry with cleanup expectations.
	DeferCleanup(t, func() { /* isolated fake client — nothing to tear down */ })

	meta := DevWorkspaceMetadata{
		DevWorkspaceId: "workspace-1",
		Namespace:      "test-ns",
		PodSelector:    map[string]string{"app": "workspace-1"},
	}
	endpoints := map[string]controllerv1alpha1.EndpointList{
		"machine1": {discoverableEndpoint("postgresql", 5432)},
	}

	services, err := GetDiscoverableServicesForEndpoints(ctx, clnt, endpoints, meta)

	assert.NoError(t, err, "single workspace with discoverable endpoint should not produce a conflict error")
	assert.Len(t, services, 1, "expected one discoverable service to be returned")
	assert.Equal(t, "postgresql", services[0].Name)
}

// TestGetDiscoverableServicesForEndpoints_NoConflict_DifferentNames verifies that two
// DevWorkspaceRoutings in the same namespace with discoverable endpoints of different names
// do not conflict with each other.
func TestGetDiscoverableServicesForEndpoints_NoConflict_DifferentNames(t *testing.T) {
	ctx := context.Background()
	scheme := makeScheme(t)
	DeferCleanup(t, func() { /* isolated fake client — nothing to tear down */ })

	// Simulate workspace-1 having already created its discoverable service for "postgresql".
	existingService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: "test-ns",
			Labels: map[string]string{
				constants.DevWorkspaceIDLabel: "workspace-1",
			},
			Annotations: map[string]string{
				constants.DevWorkspaceDiscoverableServiceAnnotation: "true",
			},
		},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}
	clnt := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingService).Build()

	// workspace-2 requests a different endpoint name "redis".
	meta2 := DevWorkspaceMetadata{
		DevWorkspaceId: "workspace-2",
		Namespace:      "test-ns",
		PodSelector:    map[string]string{"app": "workspace-2"},
	}
	endpoints2 := map[string]controllerv1alpha1.EndpointList{
		"machine1": {discoverableEndpoint("redis", 6379)},
	}

	services, err := GetDiscoverableServicesForEndpoints(ctx, clnt, endpoints2, meta2)

	assert.NoError(t, err, "two workspaces with different discoverable endpoint names should not conflict")
	assert.Len(t, services, 1)
	assert.Equal(t, "redis", services[0].Name)
}

// TestGetDiscoverableServicesForEndpoints_ConflictDetected verifies that when two
// DevWorkspaceRoutings in the same namespace both request a discoverable endpoint with the same
// name, the first reconciles successfully and the second receives a RoutingInvalid error whose
// message contains the endpoint name and the conflicting workspace reference.
func TestGetDiscoverableServicesForEndpoints_ConflictDetected(t *testing.T) {
	ctx := context.Background()
	scheme := makeScheme(t)
	DeferCleanup(t, func() { /* isolated fake client — nothing to tear down */ })

	// Simulate workspace-1 having already created its "postgresql" discoverable service.
	existingService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: "test-ns",
			Labels: map[string]string{
				constants.DevWorkspaceIDLabel: "workspace-1",
			},
			Annotations: map[string]string{
				constants.DevWorkspaceDiscoverableServiceAnnotation: "true",
			},
		},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}
	clnt := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingService).Build()

	// workspace-1 reconciliation: should succeed (no conflict — the existing service belongs to it).
	meta1 := DevWorkspaceMetadata{
		DevWorkspaceId: "workspace-1",
		Namespace:      "test-ns",
		PodSelector:    map[string]string{"app": "workspace-1"},
	}
	endpoints1 := map[string]controllerv1alpha1.EndpointList{
		"machine1": {discoverableEndpoint("postgresql", 5432)},
	}
	services1, err1 := GetDiscoverableServicesForEndpoints(ctx, clnt, endpoints1, meta1)
	assert.NoError(t, err1, "workspace-1 should reconcile its own discoverable service without error")
	assert.Len(t, services1, 1, "workspace-1 should produce one discoverable service")

	// workspace-2 reconciliation: should fail with RoutingInvalid.
	meta2 := DevWorkspaceMetadata{
		DevWorkspaceId: "workspace-2",
		Namespace:      "test-ns",
		PodSelector:    map[string]string{"app": "workspace-2"},
	}
	endpoints2 := map[string]controllerv1alpha1.EndpointList{
		"machine1": {discoverableEndpoint("postgresql", 5432)},
	}
	_, err2 := GetDiscoverableServicesForEndpoints(ctx, clnt, endpoints2, meta2)

	require.Error(t, err2, "workspace-2 should receive an error when a same-name discoverable endpoint already exists")

	var invalidErr *RoutingInvalid
	require.True(t, errors.As(err2, &invalidErr), "error should be of type *RoutingInvalid, got: %T", err2)

	assert.Contains(t, invalidErr.Reason, "postgresql",
		"error reason should mention the conflicting endpoint name")
	assert.Contains(t, invalidErr.Reason, "workspace-1",
		"error reason should reference the workspace that owns the conflicting service")
}

// TestGetDiscoverableServicesForEndpoints_SameWorkspaceNoFalsePositive verifies that a single
// DevWorkspaceRouting with a discoverable endpoint does not trigger a conflict with itself when
// re-reconciled (i.e. when its own service is already present in the namespace).
func TestGetDiscoverableServicesForEndpoints_SameWorkspaceNoFalsePositive(t *testing.T) {
	ctx := context.Background()
	scheme := makeScheme(t)
	DeferCleanup(t, func() { /* isolated fake client — nothing to tear down */ })

	// Simulate the same workspace having already created its "postgresql" discoverable service
	// from a previous reconciliation pass (idempotent re-reconciliation of itself).
	existingService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: "test-ns",
			Labels: map[string]string{
				constants.DevWorkspaceIDLabel: "workspace-1",
			},
			Annotations: map[string]string{
				constants.DevWorkspaceDiscoverableServiceAnnotation: "true",
			},
		},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}
	clnt := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingService).Build()

	// Re-reconcile the same workspace — must not produce a conflict error.
	meta := DevWorkspaceMetadata{
		DevWorkspaceId: "workspace-1",
		Namespace:      "test-ns",
		PodSelector:    map[string]string{"app": "workspace-1"},
	}
	endpoints := map[string]controllerv1alpha1.EndpointList{
		"machine1": {discoverableEndpoint("postgresql", 5432)},
	}

	// The same workspace re-reconciling itself should not trigger a conflict with itself.
	services, err := GetDiscoverableServicesForEndpoints(ctx, clnt, endpoints, meta)

	assert.NoError(t, err, "re-reconciling the same workspace's discoverable endpoint should not trigger a conflict")
	assert.Len(t, services, 1, "re-reconciliation should still return the expected service")
}
