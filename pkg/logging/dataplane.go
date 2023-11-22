package logging

import (
	"context"

	dataplanev1 "github.com/openstack-k8s-operators/dataplane-operator/api/v1beta1"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"golang.org/x/exp/slices"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PatchDataPlaneNodeSet patches the appropiate OpenStackDataPlaneNodeSet to add an OpenStackDataPlaneService
func PatchDataPlaneNodeSet(instance *telemetryv1.Logging,
	dataplaneServiceName string,
	helper *helper.Helper,
) (*dataplanev1.OpenStackDataPlaneNodeSet, controllerutil.OperationResult, error) {

	logger := helper.GetLogger()

	dataplaneNodeSet := &dataplanev1.OpenStackDataPlaneNodeSet{}
	err := helper.GetClient().Get(context.TODO(), types.NamespacedName{
		Name:      instance.Spec.NodeSetName,
		Namespace: instance.Namespace,
	}, dataplaneNodeSet)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			logger.Error(err, "openstackdataplanenodeset not found")
			return nil, controllerutil.OperationResultNone, err
		}
		logger.Error(err, "error retrieving openstackdataplanenodeset")
		return nil, controllerutil.OperationResultNone, err
	}
	if !slices.Contains(dataplaneNodeSet.Spec.Services, dataplaneServiceName) {
		op, err := controllerutil.CreateOrPatch(context.TODO(), helper.GetClient(), dataplaneNodeSet, func() error {
			dataplaneNodeSet.Spec.Services = append(dataplaneNodeSet.Spec.Services, dataplaneServiceName)
			return nil
		})
		if err != nil {
			logger.Error(err, "error patching openstackdataplanenodeset")
			return nil, controllerutil.OperationResultNone, err
		}
		return dataplaneNodeSet, op, nil
	}

	return nil, controllerutil.OperationResultNone, nil
}

// DataPlaneDeployment creates an OpenStackDataPlaneDeployment CR to trigger ansible execution
func DataPlaneDeployment(instance *telemetryv1.Logging,
	helper *helper.Helper) (*dataplanev1.OpenStackDataPlaneDeployment, controllerutil.OperationResult, error) {

	logger := helper.GetLogger()

	dataplaneDeployment := &dataplanev1.OpenStackDataPlaneDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.NodeSetName + "-" + ServiceName,
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), helper.GetClient(), dataplaneDeployment, func() error {
		dataplaneDeployment.Spec.NodeSets = []string{instance.Spec.NodeSetName}

		err := controllerutil.SetControllerReference(instance, dataplaneDeployment, helper.GetScheme())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "error creating openstackdataplanedeployment")
		return nil, controllerutil.OperationResultNone, err
	}

	return dataplaneDeployment, op, nil
}
