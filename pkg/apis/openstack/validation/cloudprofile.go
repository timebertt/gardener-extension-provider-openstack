// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"fmt"
	"net"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	api "github.com/gardener/gardener-extension-provider-openstack/pkg/apis/openstack"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateCloudProfileConfig validates a CloudProfileConfig object.
func ValidateCloudProfileConfig(cloudProfile *api.CloudProfileConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	floatingPoolPath := field.NewPath("constraints", "floatingPools")
	if len(cloudProfile.Constraints.FloatingPools) == 0 {
		allErrs = append(allErrs, field.Required(floatingPoolPath, "must provide at least one floating pool"))
	}

	combinationFound := sets.NewString()
	for i, pool := range cloudProfile.Constraints.FloatingPools {
		idxPath := floatingPoolPath.Index(i)
		if len(pool.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), "must provide a name"))
		}

		if pool.Region != nil || pool.Domain != nil {
			region := "*"
			domain := "*"
			if pool.Region != nil {
				if len(*pool.Region) == 0 {
					allErrs = append(allErrs, field.Required(idxPath.Child("region"), "must provide a region if key is present"))
				}
				region = *pool.Region
			}
			if pool.Domain != nil {
				if len(*pool.Domain) == 0 {
					allErrs = append(allErrs, field.Required(idxPath.Child("domain"), "must provide a domain if key is present"))
				}
				domain = *pool.Domain
			}
			key := fmt.Sprintf("%s,%s,%s", pool.Name, domain, region)
			if combinationFound.Has(key) {
				// duplicate for given name/domain/region combination
				allErrs = append(allErrs, field.Duplicate(idxPath.Child("name"), pool.Name))
			}
			combinationFound.Insert(key)
		}
	}

	loadBalancerProviderPath := field.NewPath("constraints", "loadBalancerProviders")
	if len(cloudProfile.Constraints.LoadBalancerProviders) == 0 {
		allErrs = append(allErrs, field.Required(loadBalancerProviderPath, "must provide at least one load balancer provider"))
	}

	regionsFound := sets.NewString()
	for i, pool := range cloudProfile.Constraints.LoadBalancerProviders {
		idxPath := loadBalancerProviderPath.Index(i)

		if len(pool.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), "must provide a name"))
		}

		if pool.Region != nil {
			if len(*pool.Region) == 0 {
				allErrs = append(allErrs, field.Required(idxPath.Child("region"), "must provide a region if key is present"))
			}
			if regionsFound.Has(*pool.Region) {
				allErrs = append(allErrs, field.Duplicate(idxPath.Child("region"), *pool.Region))
			}
			regionsFound.Insert(*pool.Region)
		}
	}

	machineImagesPath := field.NewPath("machineImages")
	if len(cloudProfile.MachineImages) == 0 {
		allErrs = append(allErrs, field.Required(machineImagesPath, "must provide at least one machine image"))
	}
	for i, machineImage := range cloudProfile.MachineImages {
		idxPath := machineImagesPath.Index(i)

		if len(machineImage.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), "must provide a name"))
		}

		if len(machineImage.Versions) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("versions"), fmt.Sprintf("must provide at least one version for machine image %q", machineImage.Name)))
		}
		for j, version := range machineImage.Versions {
			jdxPath := idxPath.Child("versions").Index(j)

			if len(version.Version) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("version"), "must provide a version"))
			}
			if len(version.Image) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("image"), "must provide an image"))
			}
		}
	}

	if len(cloudProfile.KeyStoneURL) == 0 && len(cloudProfile.KeyStoneURLs) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("keyStoneURL"), "must provide the URL to KeyStone"))
	}

	regionsFound = sets.NewString()
	for i, val := range cloudProfile.KeyStoneURLs {
		idxPath := field.NewPath("keyStoneURLs").Index(i)

		if len(val.Region) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("region"), "must provide a region"))
		}

		if len(val.URL) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("url"), "must provide an url"))
		}

		if regionsFound.Has(val.Region) {
			allErrs = append(allErrs, field.Duplicate(idxPath.Child("region"), val.Region))
		}
		regionsFound.Insert(val.Region)
	}

	for i, ip := range cloudProfile.DNSServers {
		if net.ParseIP(ip) == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("dnsServers").Index(i), ip, "must provide a valid IP"))
		}
	}

	if cloudProfile.DHCPDomain != nil && len(*cloudProfile.DHCPDomain) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("dhcpDomain"), "must provide a dhcp domain when the key is specified"))
	}

	if cloudProfile.RequestTimeout != nil {
		if _, err := time.ParseDuration(*cloudProfile.RequestTimeout); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("requestTimeout"), *cloudProfile.RequestTimeout, fmt.Sprintf("invalid duration: %v", err)))
		}
	}

	return allErrs
}
