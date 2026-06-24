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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/devfile/devworkspace-operator/pkg/constants"
)

// EnsureHomeInitContainerFields validates and sets required fields on an init-persistent-home container.
// If container.Command is nil, it is set to ["/bin/sh", "-c"]. If container.Command is non-nil and
// does not equal ["/bin/sh", "-c"] exactly, an error is returned. The function also ensures that
// container.VolumeMounts includes a mount for the persistent-home volume.
func EnsureHomeInitContainerFields(container *corev1.Container) error {
	if container.Command == nil {
		container.Command = []string{"/bin/sh", "-c"}
	} else if len(container.Command) != 2 || container.Command[0] != "/bin/sh" || container.Command[1] != "-c" {
		return fmt.Errorf("Invalid init-persistent-home container: command must be exactly [/bin/sh, -c]")
	}

	// Ensure the persistent-home volume mount is present
	for _, vm := range container.VolumeMounts {
		if vm.Name == constants.HomeVolumeName {
			// Mount already present; return without adding a duplicate
			return nil
		}
	}

	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      constants.HomeVolumeName,
		MountPath: constants.HomeUserDirectory,
	})

	return nil
}
