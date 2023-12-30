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

package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	ashappsv1 "github.com/AshrafulHaqueToni/kubebuilder/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AshReconciler reconciles a Ash object
type AshReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ash.dev.ash.dev,resources=ashes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ash.dev.ash.dev,resources=ashes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ash.dev.ash.dev,resources=ashes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ash object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *AshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logs := log.FromContext(ctx)
	logs.WithValues("ReqName", req.Name, "ReqNamespace", req.Namespace)

	// TODO(user): your logic here

	/*
		### 1: Load the Ash by name

		We'll fetch the Ash using our client.  All client methods take a
		context (to allow for cancellation) as their first argument, and the object
		in question as their last.  Get is a bit special, in that it takes a
		[`NamespacedName`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#ObjectKey)
		as the middle argument (most don't have a middle argument, as we'll see
		below).

		Many client methods also take variadic options at the end.
	*/

	var ash ashappsv1.Ash

	if err := r.Get(ctx, req.NamespacedName, &ash); err != nil {
		klog.Info(err, "unable to fetch ash")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, nil
	}
	klog.Info("Name:", ash.Name)

	// deploymentObject carry the all data of deployment in specific namespace and name
	var deploymentObject appsv1.Deployment

	// Naming of deployment
	deploymentName := ash.Name + "-" + ash.Spec.DeploymentName
	if ash.Spec.DeploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		//utilruntime.HandleError(fmt.Errorf("%s : deployment name must be specified", key))

		deploymentName = ash.Name + "-" + "dep"
		//return nil
	}

	//Creating NamespacedName for deploymentObject
	objectkey := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      deploymentName,
	}

	if err := r.Get(ctx, objectkey, &deploymentObject); err != nil {
		if errors.IsNotFound(err) {
			klog.Info("Couldn't find existing Deployment for", ash.Name, " creating one ...")
			err := r.Client.Create(ctx, newDeployment(&ash, deploymentName))
			if err != nil {
				klog.Info("error while creating deployment %s", err)
				return ctrl.Result{}, err
			} else {
				klog.Info("%s Deployment Created ", ash.Name)
			}
		} else {
			klog.Info("error fetching deployment %s", err)
			return ctrl.Result{}, err
		}
	} else {
		if ash.Spec.Replicas != nil && *ash.Spec.Replicas != *deploymentObject.Spec.Replicas {
			klog.Info(*ash.Spec.Replicas, *deploymentObject.Spec.Replicas)
			klog.Info("Deployment replica don't match ... updating")
			// As the replica count didn't match the, we need to update it
			deploymentObject.Spec.Replicas = ash.Spec.Replicas
			if err := r.Update(ctx, &deploymentObject); err != nil {
				klog.Info("error updating deployment %s", err)
			}
			klog.Info("deployment updated")
		}
	}

	// service reconcile
	var serviceObject corev1.Service
	// service Name
	serviceName := ash.Name + "-" + ash.Spec.Service.ServiceName + "-service"
	if ash.Spec.Service.ServiceName == "" {

		serviceName = ash.Name + "-" + "-service"
	}

	objectkey = client.ObjectKey{
		Namespace: req.Namespace,
		Name:      serviceName,
	}

	// create or update service
	if err := r.Get(ctx, objectkey, &serviceObject); err != nil {
		if errors.IsNotFound(err) {
			klog.Info("Could not find existing Service for", ash.Name, "creating one ...")
			err := r.Create(ctx, newService(&ash, serviceName))
			if err != nil {
				klog.Info("error while creating service %s", err)
				return ctrl.Result{}, err
			} else {
				klog.Info("%s Service Created ", ash.Name)
			}
		} else {
			fmt.Printf("error fetching service %s", err)
			return ctrl.Result{}, err
		}
	} else {
		if ash.Spec.Replicas != nil && *ash.Spec.Replicas != ash.Status.AvailableReplicas {
			klog.Info("Is this problem ?")
			klog.Info("Service replica miss match .... updating ")
			ash.Status.AvailableReplicas = *ash.Spec.Replicas

			if err := r.Status().Update(ctx, &ash); err != nil {
				klog.Info("error while updating service %s", err)
				return ctrl.Result{}, err
			}
			klog.Info("service updated")

		}
	}
	klog.Info("reconcile done")

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 1 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
var (
	deployOwnerKey = ".metadata.controller"
	svcOwnerKey    = ".metadata.controller"
	ourApiGVStr    = ashappsv1.GroupVersion.String()
	ourKind        = "Ash"
)

// SetupWithManager sets up the controller with the Manager.
func (r *AshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Indexing our Owns resource. This will allow for quickly answer the question
	// If Owns Resource x is updated, which Ash are affected?
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, deployOwnerKey, func(object client.Object) []string {
		// grab the deployment object, extract the owner
		deployment := object.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)

		if owner == nil {
			return nil
		}
		// make sure it is Ash
		if owner.APIVersion != ourApiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}

	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, svcOwnerKey, func(object client.Object) []string {
		// grab the deployment object, extract the owner
		svc := object.(*corev1.Service)
		owner := metav1.GetControllerOf(svc)

		if owner == nil {
			return nil
		}
		// make sure it is Ash
		if owner.APIVersion != ourApiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}

	}); err != nil {
		return err
	}
	// Implementation with watches and custom eventHandler
	// If someone edit hte resources(here example given for deployment resource)  by kubectl

	handlerForDeployment := handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
		// List all the CR
		ashs := &ashappsv1.AshList{}
		if err := r.List(context.Background(), ashs); err != nil {
			return nil
		}
		// this function return reconcile request array
		var request []reconcile.Request

		for _, deploy := range ashs.Items {
			deploymentName := func() string {
				name := deploy.Name + "-" + deploy.Spec.DeploymentName + "-depl"
				if deploy.Spec.DeploymentName == "" {
					name = deploy.Name + "-missingname-depl"
				}
				return name
			}()
			// Find the deployment owned by the CR
			if deploymentName == obj.GetName() && deploy.Namespace == obj.GetNamespace() {
				curdeploy := &appsv1.Deployment{}
				if err := r.Get(context.Background(), types.NamespacedName{
					Namespace: obj.GetNamespace(),
					Name:      obj.GetName(),
				}, curdeploy); err != nil {
					// This case can happen if somehow deployment gets deleted by
					// Kubectl command. We need to append new reconcile request to array
					// to create desired number of deployment again.
					if errors.IsNotFound(err) {
						request = append(request, reconcile.Request{
							NamespacedName: types.NamespacedName{
								Namespace: deploy.Namespace,
								Name:      deploy.Name,
							},
						})
						continue
					} else {
						return nil
					}
				}
				// Only append to the reconcile request array if replica count miss match.
				if curdeploy.Spec.Replicas != deploy.Spec.Replicas {
					request = append(request, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: deploy.Namespace,
							Name:      deploy.Name,
						},
					})
				}
			}
		}
		return request

	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&ashappsv1.Ash{}).
		//Owns(&appsv1.Deployment{}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, handlerForDeployment).
		Owns(&corev1.Service{}).
		Complete(r)
}
