/*
Copyright 2022.

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

package v1beta1

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCloudKittySpecDefaultWithUnifiedImage(t *testing.T) {
	g := NewWithT(t)

	// Set the unified CloudKitty image env var
	_ = os.Setenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT", "cloudkitty-unified-image")
	SetupDefaultsCloudKitty()
	defer func() {
		_ = os.Unsetenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT")
		SetupDefaultsCloudKitty()
	}()

	spec := &CloudKittySpec{}

	// Call Default()
	spec.Default()

	// Verify both API and Processor use the unified image
	g.Expect(spec.CloudKittyAPI.ContainerImage).To(Equal("cloudkitty-unified-image"),
		"CloudKittyAPI should use the unified image")
	g.Expect(spec.CloudKittyProc.ContainerImage).To(Equal("cloudkitty-unified-image"),
		"CloudKittyProc should use the unified image")
}

func TestCloudKittySpecDefaultWithNoEnvVar(t *testing.T) {
	g := NewWithT(t)

	// Ensure all CloudKitty image env vars are unset to test fallback to hardcoded defaults
	_ = os.Unsetenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT")
	SetupDefaultsCloudKitty()

	spec := &CloudKittySpec{}

	// Call Default()
	spec.Default()

	// Verify both use the hardcoded default (CloudKittyContainerImage constant)
	g.Expect(spec.CloudKittyAPI.ContainerImage).To(Equal(CloudKittyContainerImage),
		"CloudKittyAPI should use the hardcoded default when no env var is set")
	g.Expect(spec.CloudKittyProc.ContainerImage).To(Equal(CloudKittyContainerImage),
		"CloudKittyProc should use the hardcoded default when no env var is set")
}

func TestCloudKittySpecDefaultWithUnifiedImageAndContainerImagesInCR(t *testing.T) {
	g := NewWithT(t)

	// Set the unified CloudKitty image env var
	_ = os.Setenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT", "cloudkitty-unified-image")
	SetupDefaultsCloudKitty()
	defer func() {
		_ = os.Unsetenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT")
		SetupDefaultsCloudKitty()
	}()

	// Create spec with explicit container images in CR
	spec := &CloudKittySpec{
		CloudKittyAPI: CloudKittyAPITemplate{
			ContainerImage: "cloudkitty-api-custom-image-cr",
		},
		CloudKittyProc: CloudKittyProcTemplate{
			ContainerImage: "cloudkitty-proc-custom-image-cr",
		},
	}

	// Call Default()
	spec.Default()

	// Verify the CR values are preserved
	g.Expect(spec.CloudKittyAPI.ContainerImage).To(Equal("cloudkitty-api-custom-image-cr"),
		"CloudKittyAPI should use the image from CR spec")
	g.Expect(spec.CloudKittyProc.ContainerImage).To(Equal("cloudkitty-proc-custom-image-cr"),
		"CloudKittyProc should use the image from CR spec")
}

func TestCloudKittySpecDefaultWithPartialCRImages(t *testing.T) {
	g := NewWithT(t)

	// Set the unified CloudKitty image env var
	_ = os.Setenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT", "cloudkitty-unified-image")
	SetupDefaultsCloudKitty()
	defer func() {
		_ = os.Unsetenv("RELATED_IMAGE_CLOUDKITTY_IMAGE_URL_DEFAULT")
		SetupDefaultsCloudKitty()
	}()

	// Create spec with only API container image in CR (Proc is empty)
	spec := &CloudKittySpec{
		CloudKittyAPI: CloudKittyAPITemplate{
			ContainerImage: "cloudkitty-api-custom-image-cr",
		},
	}

	// Call Default()
	spec.Default()

	// Verify API uses CR value, Proc uses env var default
	g.Expect(spec.CloudKittyAPI.ContainerImage).To(Equal("cloudkitty-api-custom-image-cr"),
		"CloudKittyAPI should use the image from CR spec")
	g.Expect(spec.CloudKittyProc.ContainerImage).To(Equal("cloudkitty-unified-image"),
		"CloudKittyProc should use the unified image from env var when not set in CR")
}
