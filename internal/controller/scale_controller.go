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

package controller

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tanlay/deploy-scale/api/v1beta1"
	apiv1beta1 "github.com/tanlay/deploy-scale/api/v1beta1"
)

var logger = log.Log.WithName("scale_controller")

// ScaleReconciler reconciles a Scale object
type ScaleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.tanlay.com,resources=scales,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.tanlay.com,resources=scales/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.tanlay.com,resources=scales/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scale object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ScaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconcile called")
	// create a scale instance
	scale := v1beta1.Scale{}
	err := r.Get(ctx, req.NamespacedName, &scale)
	if err != nil {
		log.Error(err, "unable fetch Scale")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	startTime := scale.Spec.Start
	endTime := scale.Spec.End
	replicas := scale.Spec.Replicas
	currentHour := time.Now().Local().Hour()
	log.Info(fmt.Sprintf("currentHour: %d", currentHour))

	if currentHour >= startTime && currentHour <= endTime {
		for _, deployment := range scale.Spec.Deployments {
			deploy := v1.Deployment{}
			err := r.Get(ctx, client.ObjectKey{
				Namespace: deployment.Namespace,
				Name:      deployment.Name,
			}, &deploy)
			if err != nil {
				log.Error(err, "unable fetch Deploy")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			if deploy.Spec.Replicas != &replicas {
				deploy.Spec.Replicas = &replicas
				err = r.Update(ctx, &deploy)
				if err != nil {
					log.Error(err, "unable update Deploy")
					return ctrl.Result{}, err
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1beta1.Scale{}).
		Complete(r)
}
