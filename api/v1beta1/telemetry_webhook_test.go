package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
)

func TestTelemetrySpecCoreDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &TelemetrySpecCore{}
	spec.Ceilometer.RabbitMqClusterName = "test-rabbitmq"
	spec.CloudKitty.RabbitMqClusterName = "test-cloudkitty-rabbitmq"

	// Call Default() which should migrate rabbitMqClusterName to the appropriate bus config
	spec.Default()

	// Verify Ceilometer got notificationsBus defaulted (Ceilometer only uses Notifications, not RPC)
	g.Expect(spec.Ceilometer.NotificationsBus).ToNot(BeNil(),
		"Ceilometer notificationsBus should be initialized")
	g.Expect(spec.Ceilometer.NotificationsBus.Cluster).To(Equal("test-rabbitmq"),
		"Ceilometer notificationsBus.cluster should be defaulted from rabbitMqClusterName")

	// Verify CloudKitty got messagingBus defaulted (CloudKitty uses RPC, not Notifications)
	g.Expect(spec.CloudKitty.MessagingBus.Cluster).To(Equal("test-cloudkitty-rabbitmq"),
		"CloudKitty messagingBus.cluster should be defaulted from rabbitMqClusterName")
}

func TestCeilometerSpecCoreDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &CeilometerSpecCore{}
	spec.RabbitMqClusterName = "test-rabbitmq"

	// Call Default()
	spec.Default()

	// Verify NotificationsBus is defaulted (Ceilometer only uses Notifications)
	g.Expect(spec.NotificationsBus).ToNot(BeNil(),
		"Ceilometer notificationsBus should be initialized")
	g.Expect(spec.NotificationsBus.Cluster).To(Equal("test-rabbitmq"),
		"Ceilometer notificationsBus.cluster should be defaulted from rabbitMqClusterName")
}

func TestCeilometerSpecCoreDefaultWithExistingNotificationsBus(t *testing.T) {
	g := NewWithT(t)

	spec := &CeilometerSpecCore{}
	spec.RabbitMqClusterName = "test-rabbitmq"
	// User explicitly sets notificationsBus
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
	spec.RabbitMqClusterName = "test-rabbitmq"

	// Call Default()
	spec.Default()

	// Verify NotificationsBus is defaulted (Aodh only uses Notifications)
	g.Expect(spec.NotificationsBus).ToNot(BeNil(),
		"Aodh notificationsBus should be initialized")
	g.Expect(spec.NotificationsBus.Cluster).To(Equal("test-rabbitmq"),
		"Aodh notificationsBus.cluster should be defaulted from rabbitMqClusterName")
}

func TestCloudKittySpecBaseDefault(t *testing.T) {
	g := NewWithT(t)

	spec := &CloudKittySpecBase{}
	spec.RabbitMqClusterName = "test-rabbitmq"

	// Call Default()
	spec.Default()

	// Verify MessagingBus is defaulted (CloudKitty uses RPC messaging)
	g.Expect(spec.MessagingBus.Cluster).To(Equal("test-rabbitmq"),
		"CloudKitty messagingBus.cluster should be defaulted from rabbitMqClusterName")
}
