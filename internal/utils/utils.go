/*
Copyright 2024.

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

// Package utils provides utility functions for telemetry operator components
package utils //nolint:revive // utils is a legitimate package name for utility functions

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
)

// ConditionalWatchingReconciler is a reconciler that can conditionally watch resources
type ConditionalWatchingReconciler struct {
	client.Client
	Kclient    kubernetes.Interface
	Scheme     *runtime.Scheme
	Controller controller.Controller
	Watching   []string
	RESTMapper meta.RESTMapper
	Cache      cache.Cache
}

// MergeCustomConfigMounts returns base with each override applied: an override
// whose MountPath matches an existing base mount replaces it in place, so a
// custom-config file overrides the default file mounted at that path (matching
// the pre-kolla last-write-wins copy behaviour where custom-config/* was copied
// over the rendered defaults). Overrides that don't match a base path are
// appended. This keeps every MountPath unique so the pod spec stays valid.
func MergeCustomConfigMounts(base, overrides []corev1.VolumeMount) []corev1.VolumeMount {
	idx := make(map[string]int, len(base))
	for i, m := range base {
		idx[m.MountPath] = i
	}
	out := append([]corev1.VolumeMount(nil), base...)
	for _, o := range overrides {
		if i, ok := idx[o.MountPath]; ok {
			out[i] = o
		} else {
			idx[o.MountPath] = len(out)
			out = append(out, o)
		}
	}
	return out
}

// EnsureDeleted - Delete the object which in turn will clean the sub resources
func EnsureDeleted(ctx context.Context, helper *helper.Helper, obj client.Object) (ctrl.Result, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := helper.GetClient().Get(ctx, key, obj); err != nil {
		if k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Delete the object
	if obj.GetDeletionTimestamp().IsZero() {
		if err := helper.GetClient().Delete(ctx, obj); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil

}

// EnsureWatches ensures that a watch is set up for a given resource
func EnsureWatches(
	_ context.Context,
	r *ConditionalWatchingReconciler,
	name string,
	kind client.Object,
	handler handler.EventHandler,
	helper *helper.Helper,
) error {
	Log := helper.GetLogger()
	for _, item := range r.Watching {
		if item == name {
			// We are already watching the resource
			return nil
		}
	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Kind:    "CustomResourceDefinition",
		Version: "v1",
	})

	err := r.Get(context.Background(), client.ObjectKey{
		Name: name,
	}, u)
	if err != nil {
		return err
	}

	Log.Info(fmt.Sprintf("Starting to watch %s", name))
	err = r.Controller.Watch(source.Kind(r.Cache, kind, handler))
	if err == nil {
		r.Watching = append(r.Watching, name)
	}
	return err
}
