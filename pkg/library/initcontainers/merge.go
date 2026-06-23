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

package initcontainers

import (
	corev1 "k8s.io/api/core/v1"
)

// MergeInitContainers performs a strategic merge of init containers using 'name'
// as the merge key. Containers sharing the same name in base and patches are merged
// (non-zero patch fields overwrite base fields). New containers in patches that are
// not present in base are appended in patch order. Base container order is preserved.
func MergeInitContainers(base, patches []corev1.Container) ([]corev1.Container, error) {
	if len(patches) == 0 {
		return base, nil
	}

	// Build a lookup map for patch containers by name
	patchByName := make(map[string]corev1.Container, len(patches))
	for _, p := range patches {
		patchByName[p.Name] = p
	}

	result := make([]corev1.Container, 0, len(base)+len(patches))
	baseNames := make(map[string]bool, len(base))

	// Iterate base containers in order; merge where a patch exists
	for _, b := range base {
		baseNames[b.Name] = true
		if patch, ok := patchByName[b.Name]; ok {
			result = append(result, mergeContainer(b, patch))
		} else {
			result = append(result, b)
		}
	}

	// Append new patch containers (not present in base) in patch order
	for _, p := range patches {
		if !baseNames[p.Name] {
			result = append(result, p)
		}
	}

	return result, nil
}

// mergeContainer merges a patch container into a base container.
// Non-zero/non-nil patch fields overwrite the corresponding base fields.
func mergeContainer(base, patch corev1.Container) corev1.Container {
	merged := base

	if patch.Image != "" {
		merged.Image = patch.Image
	}
	if patch.ImagePullPolicy != "" {
		merged.ImagePullPolicy = patch.ImagePullPolicy
	}
	if len(patch.Command) > 0 {
		merged.Command = patch.Command
	}
	if len(patch.Args) > 0 {
		merged.Args = patch.Args
	}
	if patch.WorkingDir != "" {
		merged.WorkingDir = patch.WorkingDir
	}
	if len(patch.Ports) > 0 {
		merged.Ports = patch.Ports
	}
	if len(patch.EnvFrom) > 0 {
		merged.EnvFrom = patch.EnvFrom
	}
	if len(patch.Env) > 0 {
		merged.Env = mergeEnvVars(base.Env, patch.Env)
	}
	if patch.Resources.Limits != nil || patch.Resources.Requests != nil {
		merged.Resources = patch.Resources
	}
	if len(patch.VolumeMounts) > 0 {
		merged.VolumeMounts = patch.VolumeMounts
	}
	if len(patch.VolumeDevices) > 0 {
		merged.VolumeDevices = patch.VolumeDevices
	}
	if patch.LivenessProbe != nil {
		merged.LivenessProbe = patch.LivenessProbe
	}
	if patch.ReadinessProbe != nil {
		merged.ReadinessProbe = patch.ReadinessProbe
	}
	if patch.StartupProbe != nil {
		merged.StartupProbe = patch.StartupProbe
	}
	if patch.Lifecycle != nil {
		merged.Lifecycle = patch.Lifecycle
	}
	if patch.TerminationMessagePath != "" {
		merged.TerminationMessagePath = patch.TerminationMessagePath
	}
	if patch.TerminationMessagePolicy != "" {
		merged.TerminationMessagePolicy = patch.TerminationMessagePolicy
	}
	if patch.SecurityContext != nil {
		merged.SecurityContext = patch.SecurityContext
	}
	if patch.Stdin {
		merged.Stdin = patch.Stdin
	}
	if patch.StdinOnce {
		merged.StdinOnce = patch.StdinOnce
	}
	if patch.TTY {
		merged.TTY = patch.TTY
	}

	return merged
}

// mergeEnvVars merges patch env vars into base env vars using name as merge key.
// Patch vars with the same name overwrite base vars; new patch vars are appended.
func mergeEnvVars(base, patch []corev1.EnvVar) []corev1.EnvVar {
	patchByName := make(map[string]corev1.EnvVar, len(patch))
	for _, e := range patch {
		patchByName[e.Name] = e
	}

	result := make([]corev1.EnvVar, 0, len(base)+len(patch))
	baseNames := make(map[string]bool, len(base))

	for _, e := range base {
		baseNames[e.Name] = true
		if p, ok := patchByName[e.Name]; ok {
			result = append(result, p)
		} else {
			result = append(result, e)
		}
	}

	for _, e := range patch {
		if !baseNames[e.Name] {
			result = append(result, e)
		}
	}

	return result
}
