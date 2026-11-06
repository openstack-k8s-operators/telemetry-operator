/*
Copyright 2026.

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

package metricstorage

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func newOwnedTestEnv(t *testing.T, objs ...client.Object) (*telemetryv1.MetricStorage, client.Client, *helper.Helper) {
	t.Helper()
	g := gomega.NewWithT(t)

	scheme := runtime.NewScheme()
	g.Expect(telemetryv1.AddToScheme(scheme)).To(gomega.Succeed())
	g.Expect(persesv1alpha1.AddToScheme(scheme)).To(gomega.Succeed())

	instance := &telemetryv1.MetricStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric-storage",
			Namespace: "openstack",
			UID:       types.UID("metric-storage-uid"),
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	h, err := helper.NewHelper(instance, c, nil, scheme, logr.Discard())
	g.Expect(err).NotTo(gomega.HaveOccurred())
	return instance, c, h
}

func newDashboard(name string) *persesv1alpha1.PersesDashboard {
	return &persesv1alpha1.PersesDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "openstack",
		},
	}
}

// TestEnsureControlledDeleted checks only Perses objects controlled by the
// MetricStorage are deleted; same-named objects created by others are kept and
// a missing object is not an error.
func TestEnsureControlledDeleted(t *testing.T) {
	g := gomega.NewWithT(t)

	owned := newDashboard("owned")
	unowned := newDashboard("unowned")
	instance, c, h := newOwnedTestEnv(t, unowned)
	g.Expect(controllerutil.SetControllerReference(instance, owned, c.Scheme())).To(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), owned)).To(gomega.Succeed())

	_, err := EnsureControlledDeleted(context.TODO(), h, instance, newDashboard("owned"))
	g.Expect(err).NotTo(gomega.HaveOccurred())
	err = c.Get(context.TODO(), client.ObjectKeyFromObject(owned), &persesv1alpha1.PersesDashboard{})
	g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())

	_, err = EnsureControlledDeleted(context.TODO(), h, instance, newDashboard("unowned"))
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(c.Get(context.TODO(), client.ObjectKeyFromObject(unowned), &persesv1alpha1.PersesDashboard{})).To(gomega.Succeed())

	_, err = EnsureControlledDeleted(context.TODO(), h, instance, newDashboard("missing"))
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

// TestEnsureAdoptable checks new objects and objects controlled by the
// MetricStorage can be reconciled, while existing objects it does not control
// are rejected instead of being taken over.
func TestEnsureAdoptable(t *testing.T) {
	g := gomega.NewWithT(t)

	instance, c, _ := newOwnedTestEnv(t)

	g.Expect(EnsureAdoptable(instance, newDashboard("new"))).To(gomega.Succeed())

	existing := newDashboard("existing")
	existing.CreationTimestamp = metav1.Now()
	g.Expect(EnsureAdoptable(instance, existing)).To(gomega.MatchError(ErrNotControlledByMetricStorage))

	g.Expect(controllerutil.SetControllerReference(instance, existing, c.Scheme())).To(gomega.Succeed())
	g.Expect(EnsureAdoptable(instance, existing)).To(gomega.Succeed())
}
