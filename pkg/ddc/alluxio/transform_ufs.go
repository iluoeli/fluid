/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alluxio

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// transform dataset which has ufsPaths and ufsVolumes
func (e *AlluxioEngine) transformDatasetToVolume(runtime *datav1alpha1.AlluxioRuntime, dataset *datav1alpha1.Dataset, value *Alluxio) {

	mounts := dataset.Spec.Mounts
	for _, mount := range mounts {
		// if mount.MountPoint
		if strings.HasPrefix(mount.MountPoint, pathScheme) {
			if len(value.UFSPaths) == 0 {
				value.UFSPaths = []UFSPath{}
			}

			ufsPath := UFSPath{}
			ufsPath.Name = mount.Name
			ufsPath.ContainerPath = fmt.Sprintf("%s/%s", e.getLocalStorageDirectory(), mount.Name)
			ufsPath.HostPath = strings.TrimPrefix(mount.MountPoint, pathScheme)
			value.UFSPaths = append(value.UFSPaths, ufsPath)

		} else if strings.HasPrefix(mount.MountPoint, volumeScheme) {
			if len(value.UFSVolumes) == 0 {
				value.UFSVolumes = []UFSVolume{}
			}

			// If a pvc contains subdirectory, we firstly mount it to somewhere else,
			// and after master is setup, create a soft link to under storage.
			// NOTE: better finished creating soft link before master loads metadata,
			// or we will have to do that sync once again.
			if e.containsPersistVolumeClaimSubdir(mount.MountPoint) {
				value.UFSVolumes = append(value.UFSVolumes, UFSVolume{
					Name:          mount.Name,
					ContainerPath: fmt.Sprintf("%s/%s", e.getPersistVolumeClainDirectory(), mount.Name),
				})
			} else {
				value.UFSVolumes = append(value.UFSVolumes, UFSVolume{
					Name:          mount.Name,
					ContainerPath: fmt.Sprintf("%s/%s", e.getLocalStorageDirectory(), mount.Name),
				})
			}
		}
	}

	if len(value.UFSPaths) > 0 {
		// fmt.Println("UFSPaths length 1")
		if dataset.Spec.NodeAffinity != nil {
			value.Master.Affinity = Affinity{
				NodeAffinity: translateCacheToNodeAffinity(dataset.Spec.NodeAffinity),
			}
		}
	}

}
