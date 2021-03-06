/*
Copyright 2021.

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

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	expenseappv1alpha1 "github.com/nmajorov/expenses-app-operator.git/api/v1alpha1"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=expense-app.majorov.biz,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=expense-app.majorov.biz,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=expense-app.majorov.biz,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Database object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	labels := map[string]string{
		"app":  "visitors",
		"tier": "postgresql",
	}

	size := int32(1)

	userSecret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: "database-auth"}, userSecret)
	if err != nil {
		logger.Info("database-auth secret not found.")
	}

	dep := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "postgresql",
			Namespace: req.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &v1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "quay.io/centos7/postgresql-10-centos7:latest",
						Name:  "postgresql",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 5432,
							Name:          "postgresql",
						}},
						/**
						POSTGRESQL_USER  POSTGRESQL_PASSWORD  POSTGRESQL_DATABASE
						Env:	[]corev1.EnvVar{
							{
								Name:	"MYSQL_ROOT_PASSWORD",
								Value: 	"password",
							},
							{
								Name:	"MYSQL_DATABASE",
								Value:	"visitors",
							},
							{
								Name:	"MYSQL_USER",
								ValueFrom: userSecret,
							},
							{
								Name:	"MYSQL_PASSWORD",
								ValueFrom: passwordSecret,
							},
							**/
					},
					}},
			},
		},
	}

	logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	err = r.Create(ctx, dep)

	if err != nil {
		logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&expenseappv1alpha1.Database{}).
		Complete(r)
}
