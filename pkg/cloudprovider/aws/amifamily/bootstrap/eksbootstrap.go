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

package bootstrap

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

type EKS struct {
	Options
	ContainerRuntime string
}

func (e EKS) Script() (string, error) {
	var caBundleArg string
	if e.CABundle != nil {
		caBundleArg = fmt.Sprintf("--b64-cluster-ca '%s'", *e.CABundle)
	}
	var userData bytes.Buffer
	userData.WriteString("#!/bin/bash -xe\n")
	userData.WriteString("exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1\n")
	// Due to the way bootstrap.sh is written, parameters should not be passed to it with an equal sign
	userData.WriteString(fmt.Sprintf("/etc/eks/bootstrap.sh '%s' --apiserver-endpoint '%s' %s", e.ClusterName, e.ClusterEndpoint, caBundleArg))

	kubeletExtraArgs := strings.Join([]string{e.nodeLabelArg(), e.nodeTaintArg()}, " ")

	if !e.AWSENILimitedPodDensity {
		userData.WriteString(" \\\n--use-max-pods false")
		kubeletExtraArgs += " --max-pods=110"
	}
	if e.ContainerRuntime != "" {
		userData.WriteString(fmt.Sprintf(" \\\n--container-runtime %s", e.ContainerRuntime))
	}
	if kubeletExtraArgs = strings.Trim(kubeletExtraArgs, " "); len(kubeletExtraArgs) > 0 {
		userData.WriteString(fmt.Sprintf(" \\\n--kubelet-extra-args '%s'", kubeletExtraArgs))
	}
	if e.KubeletConfig != nil && len(e.KubeletConfig.ClusterDNS) > 0 {
		userData.WriteString(fmt.Sprintf(" \\\n--dns-cluster-ip '%s'", e.KubeletConfig.ClusterDNS[0]))
	}
	return base64.StdEncoding.EncodeToString(userData.Bytes()), nil
}
