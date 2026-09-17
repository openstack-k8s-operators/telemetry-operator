/*
Copyright 2023.

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

func TestAutoscalingSpecDefaultWithUnifiedImage(t *testing.T) {
	g := NewWithT(t)

	// Set the unified Aodh image env var
	_ = os.Setenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT", "aodh-unified-image")
	SetupDefaultsAutoscaling()
	defer func() {
		_ = os.Unsetenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT")
		SetupDefaultsAutoscaling()
	}()

	spec := &AutoscalingSpec{}

	// Call Default()
	spec.Default()

	// Verify all four Aodh images use the unified image
	g.Expect(spec.Aodh.APIImage).To(Equal("aodh-unified-image"),
		"Aodh API should use the unified image")
	g.Expect(spec.Aodh.EvaluatorImage).To(Equal("aodh-unified-image"),
		"Aodh Evaluator should use the unified image")
	g.Expect(spec.Aodh.NotifierImage).To(Equal("aodh-unified-image"),
		"Aodh Notifier should use the unified image")
	g.Expect(spec.Aodh.ListenerImage).To(Equal("aodh-unified-image"),
		"Aodh Listener should use the unified image")
}

func TestAutoscalingSpecDefaultWithNoEnvVar(t *testing.T) {
	g := NewWithT(t)

	// Ensure all Aodh image env vars are unset to test fallback to hardcoded defaults
	_ = os.Unsetenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT")
	SetupDefaultsAutoscaling()

	spec := &AutoscalingSpec{}

	// Call Default()
	spec.Default()

	// Verify all use the hardcoded default (AodhContainerImage constant)
	g.Expect(spec.Aodh.APIImage).To(Equal(AodhContainerImage),
		"Aodh API should use the hardcoded default when no env var is set")
	g.Expect(spec.Aodh.EvaluatorImage).To(Equal(AodhContainerImage),
		"Aodh Evaluator should use the hardcoded default when no env var is set")
	g.Expect(spec.Aodh.NotifierImage).To(Equal(AodhContainerImage),
		"Aodh Notifier should use the hardcoded default when no env var is set")
	g.Expect(spec.Aodh.ListenerImage).To(Equal(AodhContainerImage),
		"Aodh Listener should use the hardcoded default when no env var is set")
}

func TestAutoscalingSpecDefaultWithUnifiedImageAndContainerImagesInCR(t *testing.T) {
	g := NewWithT(t)

	// Set the unified Aodh image env var
	_ = os.Setenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT", "aodh-unified-image")
	SetupDefaultsAutoscaling()
	defer func() {
		_ = os.Unsetenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT")
		SetupDefaultsAutoscaling()
	}()

	// Create spec with explicit container images in CR
	spec := &AutoscalingSpec{
		Aodh: Aodh{
			APIImage:       "aodh-api-custom-image-cr",
			EvaluatorImage: "aodh-evaluator-custom-image-cr",
			NotifierImage:  "aodh-notifier-custom-image-cr",
			ListenerImage:  "aodh-listener-custom-image-cr",
		},
	}

	// Call Default()
	spec.Default()

	// Verify the CR values are preserved
	g.Expect(spec.Aodh.APIImage).To(Equal("aodh-api-custom-image-cr"),
		"Aodh API should use the image from CR spec")
	g.Expect(spec.Aodh.EvaluatorImage).To(Equal("aodh-evaluator-custom-image-cr"),
		"Aodh Evaluator should use the image from CR spec")
	g.Expect(spec.Aodh.NotifierImage).To(Equal("aodh-notifier-custom-image-cr"),
		"Aodh Notifier should use the image from CR spec")
	g.Expect(spec.Aodh.ListenerImage).To(Equal("aodh-listener-custom-image-cr"),
		"Aodh Listener should use the image from CR spec")
}

func TestAutoscalingSpecDefaultWithPartialCRImages(t *testing.T) {
	g := NewWithT(t)

	// Set the unified Aodh image env var
	_ = os.Setenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT", "aodh-unified-image")
	SetupDefaultsAutoscaling()
	defer func() {
		_ = os.Unsetenv("RELATED_IMAGE_AODH_IMAGE_URL_DEFAULT")
		SetupDefaultsAutoscaling()
	}()

	// Create spec with only API container image in CR (others are empty)
	spec := &AutoscalingSpec{
		Aodh: Aodh{
			APIImage: "aodh-api-custom-image-cr",
			// EvaluatorImage, NotifierImage, ListenerImage are empty
		},
	}

	// Call Default()
	spec.Default()

	// Verify API uses CR value, others use env var default
	g.Expect(spec.Aodh.APIImage).To(Equal("aodh-api-custom-image-cr"),
		"Aodh API should use the image from CR spec")
	g.Expect(spec.Aodh.EvaluatorImage).To(Equal("aodh-unified-image"),
		"Aodh Evaluator should use the unified image from env var when not set in CR")
	g.Expect(spec.Aodh.NotifierImage).To(Equal("aodh-unified-image"),
		"Aodh Notifier should use the unified image from env var when not set in CR")
	g.Expect(spec.Aodh.ListenerImage).To(Equal("aodh-unified-image"),
		"Aodh Listener should use the unified image from env var when not set in CR")
}
