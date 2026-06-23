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
			name:    "empty base and empty patches",
			base:    []corev1.Container{},
			patches: []corev1.Container{},
			want:    []corev1.Container{},
		},
		{
			name:    "nil base and nil patches",
			base:    nil,
			patches: nil,
			want:    nil,
		},
		{
			name:    "nil base with patch containers",
			base:    nil,
			patches: []corev1.Container{{Name: "new-container", Image: "new-image"}},
			want:    []corev1.Container{{Name: "new-container", Image: "new-image"}},
		},
		{
			name:    "base containers with nil patches",
			base:    []corev1.Container{{Name: "base-container", Image: "base-image"}},
			patches: nil,
			want:    []corev1.Container{{Name: "base-container", Image: "base-image"}},
		},
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
			name: "multiple new patch containers appended in patch order",
			base: []corev1.Container{
				{Name: "base-a", Image: "image-a"},
			},
			patches: []corev1.Container{
				{Name: "new-first", Image: "image-first"},
				{Name: "new-second", Image: "image-second"},
				{Name: "new-third", Image: "image-third"},
			},
			want: []corev1.Container{
				{Name: "base-a", Image: "image-a"},
				{Name: "new-first", Image: "image-first"},
				{Name: "new-second", Image: "image-second"},
				{Name: "new-third", Image: "image-third"},
			},
		},
		{
			name: "same-name merge patch fields overwrite base fields",
			base: []corev1.Container{
				{
					Name:    "shared-container",
					Image:   "base-image:v1",
					Command: []string{"/bin/sh"},
					Args:    []string{"base-arg"},
				},
			},
			patches: []corev1.Container{
				{
					Name:  "shared-container",
					Image: "patched-image:v2",
					Args:  []string{"patched-arg"},
				},
			},
			want: []corev1.Container{
				{
					Name:    "shared-container",
					Image:   "patched-image:v2",
					Command: []string{"/bin/sh"},
					Args:    []string{"patched-arg"},
				},
			},
		},
		{
			name: "base order preserved with interleaved patches",
			base: []corev1.Container{
				{Name: "alpha", Image: "image-alpha"},
				{Name: "beta", Image: "image-beta"},
				{Name: "gamma", Image: "image-gamma"},
			},
			patches: []corev1.Container{
				{Name: "beta", Image: "image-beta-patched"},
				{Name: "new-delta", Image: "image-delta"},
			},
			want: []corev1.Container{
				{Name: "alpha", Image: "image-alpha"},
				{Name: "beta", Image: "image-beta-patched"},
				{Name: "gamma", Image: "image-gamma"},
				{Name: "new-delta", Image: "image-delta"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeInitContainers(tt.base, tt.patches)
			assert.NoError(t, err, "should not return error")
			assert.Equal(t, tt.want, got, "should return merged containers")
		})
	}
}
