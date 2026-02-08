package controller

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestTransportURLCreateOrUpdateWithNilConfig tests that transportURLCreateOrUpdate
// handles nil rabbitmqConfig gracefully without panicking
func TestTransportURLCreateOrUpdateWithNilConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	// Create a fake client
	scheme := runtime.NewScheme()
	_ = telemetryv1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	reconciler := &AutoscalingReconciler{
		Client: client,
		Scheme: scheme,
	}

	instance := &telemetryv1.Autoscaling{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-autoscaling",
			Namespace: "test-namespace",
		},
		Spec: telemetryv1.AutoscalingSpec{
			Aodh: telemetryv1.Aodh{
				AodhCore: telemetryv1.AodhCore{
					// NotificationsBus is nil
					NotificationsBus: nil,
				},
			},
		},
	}

	serviceLabels := map[string]string{
		"app": "test",
	}

	// Call transportURLCreateOrUpdate with nil rabbitmqConfig
	// This should return an error, not panic
	transportURL, op, err := reconciler.transportURLCreateOrUpdate(
		context.Background(),
		instance,
		serviceLabels,
		nil, // nil rabbitmqConfig
	)

	// Should return an error
	g.Expect(err).To(gomega.HaveOccurred(), "Should return error when rabbitmqConfig is nil")
	g.Expect(err.Error()).To(gomega.ContainSubstring("rabbitmqConfig is nil"),
		"Error message should indicate rabbitmqConfig is nil")

	// Should return nil transportURL and OperationResultNone
	g.Expect(transportURL).To(gomega.BeNil(), "TransportURL should be nil when config is nil")
	g.Expect(string(op)).To(gomega.Equal("unchanged"), "Operation should be OperationResultNone when config is nil")
}
