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

package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/devfile/devworkspace-operator/controllers/controller/devworkspacerouting/conversion"
	"github.com/devfile/devworkspace-operator/pkg/dwerrors"
	"github.com/devfile/devworkspace-operator/pkg/provision/sync"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
	maputils "github.com/devfile/devworkspace-operator/internal/map"
	"github.com/devfile/devworkspace-operator/pkg/common"
	"github.com/devfile/devworkspace-operator/pkg/constants"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func checkRoutingConflicts(
	ctx context.Context,
	c client.Client,
	workspace *common.DevWorkspaceWithConfig,
	reqLogger logr.Logger) error {

	// Collect all endpoint names from the current workspace into a set for efficient lookup.
	workspaceEndpoints := map[string]bool{}
	for _, component := range workspace.Spec.Template.Components {
		if component.Container != nil {
			for _, endpoint := range component.Container.Endpoints {
				if endpoint.Exposure == "internal" || endpoint.Exposure == "public" {
					endpointName := common.EndpointName(endpoint.Name)
					workspaceEndpoints[endpointName] = true
				}
			}
		}
	}

	// If there are no endpoints to check, we can exit early.
	if len(workspaceEndpoints) == 0 {
		return nil
	}

	// Check for conflicts with other DevWorkspaces in the same namespace.
	devWorkspaceList := &dw.DevWorkspaceList{}
	if err := c.List(ctx, devWorkspaceList, &client.ListOptions{Namespace: workspace.Namespace}); err != nil {
		return err
	}

	for _, otherWorkspace := range devWorkspaceList.Items {
		if otherWorkspace.UID == workspace.UID {
			continue // Skip the current workspace
		}
		if otherWorkspace.Status.Phase == dw.DevWorkspaceStatusRunning || otherWorkspace.Status.Phase == dw.DevWorkspaceStatusStarting {
			for _, component := range otherWorkspace.Spec.Template.Components {
				if component.Container != nil {
					for _, endpoint := range component.Container.Endpoints {
						endpointName := common.EndpointName(endpoint.Name)
						if _, ok := workspaceEndpoints[endpointName]; ok {
							return &dwerrors.FailError{
								Message: fmt.Sprintf("Endpoint name '%s' conflicts with an active workspace '%s' in the same namespace. Please choose a different endpoint name.", endpointName, otherWorkspace.Name),
							}
						}
					}
				}
			}
		}
	}

	// Check for conflicts with existing services in the namespace that are not owned by this workspace.
	serviceList := &corev1.ServiceList{}
	if err := c.List(ctx, serviceList, &client.ListOptions{Namespace: workspace.Namespace}); err != nil {
		return err
	}

	for _, service := range serviceList.Items {
		if _, ok := workspaceEndpoints[service.Name]; ok {
			if ownerId, ok := service.Labels[constants.DevWorkspaceIDLabel]; !ok || ownerId != workspace.Status.DevWorkspaceId {
				return fmt.Errorf("service '%s' already exists in this namespace and is not owned by this workspace, this may indicate an endpoint name conflict, please choose a different endpoint name", service.Name)
			}
		}
	}

	return nil
}

func SyncRoutingToCluster(
	workspace *common.DevWorkspaceWithConfig,
	clusterAPI sync.ClusterAPI) (*v1alpha1.PodAdditions, map[string]v1alpha1.ExposedEndpointList, string, error) {

	// Call the new conflict check function
	if err := checkRoutingConflicts(clusterAPI.Ctx, clusterAPI.Client, workspace, clusterAPI.Logger); err != nil {
		return nil, nil, "", err
	}

	specRouting, err := getSpecRouting(workspace, clusterAPI.Scheme)
	if err != nil {
		return nil, nil, "", err
	}

	clusterObj, err := sync.SyncObjectWithCluster(specRouting, clusterAPI)
	if err != nil {
		return nil, nil, "", dwerrors.WrapSyncError(err)
	}

	clusterRouting := clusterObj.(*v1alpha1.DevWorkspaceRouting)
	statusMsg := clusterRouting.Status.Message
	if clusterRouting.Status.Phase == v1alpha1.RoutingFailed {
		return nil, nil, statusMsg, &dwerrors.FailError{Message: statusMsg}
	}
	if clusterRouting.Status.Phase != v1alpha1.RoutingReady {
		return nil, nil, statusMsg, &dwerrors.RetryError{Message: statusMsg, RequeueAfter: 5 * time.Second}
	}

	// Configure securityContext for pod additions, for example che-gateway container
	// https://github.com/eclipse-che/che/issues/22747
	if clusterRouting.Status.PodAdditions != nil &&
		workspace.Config.Workspace != nil &&
		workspace.Config.Workspace.ContainerSecurityContext != nil {

		for i, container := range clusterRouting.Status.PodAdditions.Containers {
			if container.SecurityContext == nil {
				clusterRouting.Status.PodAdditions.Containers[i].SecurityContext = workspace.Config.Workspace.ContainerSecurityContext
			}
		}
	}

	return clusterRouting.Status.PodAdditions, clusterRouting.Status.ExposedEndpoints, statusMsg, nil
}

func getSpecRouting(
	workspace *common.DevWorkspaceWithConfig,
	scheme *runtime.Scheme) (*v1alpha1.DevWorkspaceRouting, error) {

	endpoints := map[string]v1alpha1.EndpointList{}
	for _, component := range workspace.Spec.Template.Components {
		if component.Container == nil {
			continue
		}
		componentEndpoints := component.Container.Endpoints
		if len(componentEndpoints) > 0 {
			endpoints[component.Name] = append(endpoints[component.Name], conversion.ConvertAllDevfileEndpoints(componentEndpoints)...)
		}
	}

	var annotations map[string]string
	if val, ok := workspace.Annotations[constants.DevWorkspaceRestrictedAccessAnnotation]; ok {
		annotations = maputils.Append(annotations, constants.DevWorkspaceRestrictedAccessAnnotation, val)
	}
	annotations = maputils.Append(annotations, constants.DevWorkspaceStartedStatusAnnotation, "true")

	// copy the annotations for the specific routingClass from the workspace object to the routing
	expectedAnnotationPrefix := workspace.Spec.RoutingClass + constants.RoutingAnnotationInfix
	for k, v := range workspace.GetAnnotations() {
		if strings.HasPrefix(k, expectedAnnotationPrefix) {
			annotations = maputils.Append(annotations, k, v)
		}
	}

	routingClass := workspace.Spec.RoutingClass
	if routingClass == "" {
		routingClass = workspace.Config.Routing.DefaultRoutingClass
	}

	routing := &v1alpha1.DevWorkspaceRouting{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.DevWorkspaceRoutingName(workspace.Status.DevWorkspaceId),
			Namespace: workspace.Namespace,
			Labels: map[string]string{
				constants.DevWorkspaceIDLabel: workspace.Status.DevWorkspaceId,
			},
			Annotations: annotations,
		},
		Spec: v1alpha1.DevWorkspaceRoutingSpec{
			DevWorkspaceId: workspace.Status.DevWorkspaceId,
			RoutingClass:   v1alpha1.DevWorkspaceRoutingClass(routingClass),
			Endpoints:      endpoints,
			PodSelector: map[string]string{
				constants.DevWorkspaceIDLabel: workspace.Status.DevWorkspaceId,
			},
		},
	}
	err := controllerutil.SetControllerReference(workspace.DevWorkspace, routing, scheme)
	if err != nil {
		return nil, err
	}

	return routing, nil
}
