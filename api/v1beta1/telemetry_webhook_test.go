package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
)

func TestTelemetrySpecCoreDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &TelemetrySpecCore{}
	// Set NotificationsBus and MessagingBus without Cluster to test defaulting
	spec.Ceilometer.NotificationsBus = &rabbitmqv1.RabbitMqConfig{}
	spec.CloudKitty.MessagingBus = rabbitmqv1.RabbitMqConfig{}

	// Call Default()
	spec.Default()

	// Verify Ceilometer NotificationsBus.Cluster was defaulted
	g.Expect(spec.Ceilometer.NotificationsBus.Cluster).To(Equal("rabbitmq"),
		"Ceilometer notificationsBus.cluster should be defaulted to rabbitmq")

	// Verify CloudKitty MessagingBus.Cluster was defaulted
	g.Expect(spec.CloudKitty.MessagingBus.Cluster).To(Equal("rabbitmq"),
		"CloudKitty messagingBus.cluster should be defaulted to rabbitmq")
}

func TestCeilometerSpecCoreDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &CeilometerSpecCore{}
	// Set NotificationsBus without Cluster to test defaulting
	spec.NotificationsBus = &rabbitmqv1.RabbitMqConfig{}

	// Call Default()
	spec.Default()

	// Verify NotificationsBus.Cluster is defaulted
	g.Expect(spec.NotificationsBus.Cluster).To(Equal("rabbitmq"),
		"Ceilometer notificationsBus.cluster should be defaulted to rabbitmq")
}

func TestCeilometerSpecCoreDefaultWithExistingNotificationsBus(t *testing.T) {
	g := NewWithT(t)

	spec := &CeilometerSpecCore{}
	// User explicitly sets notificationsBus with all fields
	spec.NotificationsBus = &rabbitmqv1.RabbitMqConfig{
		Cluster: "custom-notifications-rabbitmq",
		User:    "custom-user",
		Vhost:   "custom-vhost",
	}

	// Call Default()
	spec.Default()

	// Verify NotificationsBus preserves explicit values
	g.Expect(spec.NotificationsBus.Cluster).To(Equal("custom-notifications-rabbitmq"),
		"Ceilometer notificationsBus.cluster should preserve user-specified value")
	g.Expect(spec.NotificationsBus.User).To(Equal("custom-user"),
		"Ceilometer notificationsBus.user should preserve user-specified value")
	g.Expect(spec.NotificationsBus.Vhost).To(Equal("custom-vhost"),
		"Ceilometer notificationsBus.vhost should preserve user-specified value")
}

func TestAodhCoreDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &AodhCore{}
	// Set NotificationsBus without Cluster to test defaulting
	spec.NotificationsBus = &rabbitmqv1.RabbitMqConfig{}

	// Call Default()
	spec.Default()

	// Verify NotificationsBus.Cluster is defaulted
	g.Expect(spec.NotificationsBus.Cluster).To(Equal("rabbitmq"),
		"Aodh notificationsBus.cluster should be defaulted to rabbitmq")
}

func TestCloudKittySpecBaseDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &CloudKittySpecBase{}
	// MessagingBus.Cluster is empty, should be defaulted

	// Call Default()
	spec.Default()

	// Verify MessagingBus.Cluster is defaulted
	g.Expect(spec.MessagingBus.Cluster).To(Equal("rabbitmq"),
		"CloudKitty messagingBus.cluster should be defaulted to rabbitmq")
}
