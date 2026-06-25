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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestMergeInitContainers(t *testing.T) {
	tests := []struct {
		name    string
		base    []corev1.Container
		patches []corev1.Container
		want    []corev1.Container
	}{
		{
			name: "empty base",
			base: []corev1.Container{},
			patches: []corev1.Container{
				{Name: "new-container", Image: "new-image"},
			},
			want: []corev1.Container{
				{Name: "new-container", Image: "new-image"},
			},
		},
		{
			name: "empty patches",
			base: []corev1.Container{
				{Name: "base-container", Image: "base-image"},
			},
			patches: []corev1.Container{},
			want: []corev1.Container{
				{Name: "base-container", Image: "base-image"},
			},
		},
		{
			name: "multiple containers",
			base: []corev1.Container{
				{Name: "first", Image: "first-image"},
				{Name: "second", Image: "second-image"},
				{Name: "third", Image: "third-image"},
			},
			patches: []corev1.Container{
				{Name: "new-container", Image: "new-image"},
				{Name: "second", Image: "updated-second-image"},
			},
			want: []corev1.Container{
				{Name: "first", Image: "first-image"},
				{Name: "second", Image: "updated-second-image"},
				{Name: "third", Image: "third-image"},
				{Name: "new-container", Image: "new-image"},
			},
		},
		{
			name: "partial field merge",
			base: []corev1.Container{
				{
					Name:    "base-container",
					Image:   "base-image",
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"echo 'base'"},
					Env:     []corev1.EnvVar{{Name: "BASE_VAR", Value: "base-value"}},
				},
			},
			patches: []corev1.Container{
				{
					Name: "base-container",
					Args: []string{"echo 'patched'"}, // only this field changed
				},
			},
			want: []corev1.Container{
				{
					Name:    "base-container",
					Image:   "base-image",
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"echo 'patched'"},
					Env:     []corev1.EnvVar{{Name: "BASE_VAR", Value: "base-value"}},
				},
			},
		},
		{
			name: "preserve user-configured init-persistent-home content",
			base: []corev1.Container{
				{
					Name:    "init-persistent-home",
					Image:   "workspace-image:latest",
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"default stow script"},
				},
			},
			patches: []corev1.Container{
				{
					Name:  "init-persistent-home",
					Image: "custom-image:latest",
					Args:  []string{"echo 'custom init'"},
					Env: []corev1.EnvVar{
						{
							Name:  "CUSTOM_VAR",
							Value: "custom-value",
						},
					},
				},
			},
			want: []corev1.Container{
				{
					Name:    "init-persistent-home",
					Image:   "custom-image:latest",
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"echo 'custom init'"},
					Env: []corev1.EnvVar{
						{
							Name:  "CUSTOM_VAR",
							Value: "custom-value",
						},
					},
				},
			},
		},
		// Ordering: multiple new patch containers are appended in their declaration order
		{
			name: "new patch containers appended in declaration order",
			base: []corev1.Container{
				{Name: "base-a", Image: "image-a"},
			},
			patches: []corev1.Container{
				{Name: "patch-z", Image: "image-z"},
				{Name: "patch-m", Image: "image-m"},
				{Name: "patch-a", Image: "image-patch-a"},
			},
			want: []corev1.Container{
				{Name: "base-a", Image: "image-a"},
				{Name: "patch-z", Image: "image-z"},
				{Name: "patch-m", Image: "image-m"},
				{Name: "patch-a", Image: "image-patch-a"},
			},
		},
		// Ordering: base containers maintain their original order when multiple are patched
		{
			name: "base order preserved when multiple base containers are patched",
			base: []corev1.Container{
				{Name: "first", Image: "first-original"},
				{Name: "second", Image: "second-original"},
				{Name: "third", Image: "third-original"},
			},
			patches: []corev1.Container{
				{Name: "third", Image: "third-patched"},
				{Name: "first", Image: "first-patched"},
			},
			want: []corev1.Container{
				{Name: "first", Image: "first-patched"},
				{Name: "second", Image: "second-original"},
				{Name: "third", Image: "third-patched"},
			},
		},
		// Deduplication: same-name patch entry overwrites base-specific fields while keeping unset base fields
		{
			name: "deduplication: same-name patch overwrites specified base fields",
			base: []corev1.Container{
				{
					Name:    "shared",
					Image:   "base-image",
					Command: []string{"/bin/base"},
					Args:    []string{"base-arg"},
				},
			},
			patches: []corev1.Container{
				{
					Name:    "shared",
					Image:   "patch-image",
					Command: []string{"/bin/patch"},
				},
			},
			want: []corev1.Container{
				{
					Name:    "shared",
					Image:   "patch-image",
					Command: []string{"/bin/patch"},
					Args:    []string{"base-arg"},
				},
			},
		},
		// Deduplication: only one container survives when patch duplicates a base name
		{
			name: "deduplication: only one result container per name",
			base: []corev1.Container{
				{Name: "alpha", Image: "alpha-base"},
				{Name: "beta", Image: "beta-base"},
			},
			patches: []corev1.Container{
				{Name: "alpha", Image: "alpha-patch"},
			},
			want: []corev1.Container{
				{Name: "alpha", Image: "alpha-patch"},
				{Name: "beta", Image: "beta-base"},
			},
		},
		// Edge case: both base and patches are empty
		{
			name:    "both base and patches are empty",
			base:    []corev1.Container{},
			patches: []corev1.Container{},
			want:    []corev1.Container{},
		},
		// Edge case: nil base slice treated same as empty
		{
			name: "nil base slice",
			base: nil,
			patches: []corev1.Container{
				{Name: "patch-only", Image: "patch-image"},
			},
			want: []corev1.Container{
				{Name: "patch-only", Image: "patch-image"},
			},
		},
		// Edge case: nil patches slice treated same as empty
		{
			name: "nil patches slice",
			base: []corev1.Container{
				{Name: "base-only", Image: "base-image"},
			},
			patches: nil,
			want: []corev1.Container{
				{Name: "base-only", Image: "base-image"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeInitContainers(tt.base, tt.patches)
			assert.NoError(t, err, "should not return error")
			assert.Equal(t, tt.want, got, "should return merged containers")
		})
	}
}
