package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestTelemetrySpecCoreDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &TelemetrySpecCore{}
	spec.Ceilometer.RabbitMqClusterName = "test-rabbitmq"
	spec.CloudKitty.RabbitMqClusterName = "test-cloudkitty-rabbitmq"

	// Call Default() which should migrate rabbitMqClusterName to messagingBus.cluster
	spec.Default()

	// Verify Ceilometer got defaulted
	g.Expect(spec.Ceilometer.MessagingBus.Cluster).To(Equal("test-rabbitmq"),
		"Ceilometer messagingBus.cluster should be defaulted from rabbitMqClusterName")

	// Verify CloudKitty got defaulted
	g.Expect(spec.CloudKitty.MessagingBus.Cluster).To(Equal("test-cloudkitty-rabbitmq"),
		"CloudKitty messagingBus.cluster should be defaulted from rabbitMqClusterName")
}
