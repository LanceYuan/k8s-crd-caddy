/*
Copyright 2022 Lance Yuan.

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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devopsv1 "k8s-crd-caddy/api/v1"
)

const (
	controllerName string = "caddy-controller"
)

var logger logr.Logger

// StaticReconciler reconciles a Static object
type StaticReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devops.codepy.net,resources=statics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.codepy.net,resources=statics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.codepy.net,resources=statics/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/rollback,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/scale,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Static object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *StaticReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger = log.FromContext(ctx)

	instance := &devopsv1.Static{}
	// TODO(user): your logic here
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("get App not found!!!!!")
			return ctrl.Result{}, nil
		}
		logger.Info("get App error!!!!!")
		return reconcile.Result{}, err
	}
	deployment := &appsv1.Deployment{}
	svc := &corev1.Service{}
	caddyObjectKey := client.ObjectKey{
		Name:      controllerName,
		Namespace: req.Namespace,
	}
	if err := r.Client.Get(ctx, caddyObjectKey, deployment); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		deployment = NewDeployment(instance)
		if err := r.Client.Create(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := r.Client.Get(ctx, caddyObjectKey, svc); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		svc = NewService(instance)
		if err := r.Client.Create(ctx, svc); err != nil {
			return ctrl.Result{}, err
		}
	}
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("start delete ingress....")
		ingress := &networkingv1.Ingress{}
		if err := r.Client.Get(ctx, req.NamespacedName, ingress); err != nil {
			if errors.IsNotFound(err) {
				instance.ObjectMeta.Finalizers = []string{}
				if err := r.Update(ctx, instance); err != nil {
					return ctrl.Result{}, err
				}
				return reconcile.Result{}, nil
			}
			return ctrl.Result{}, err
		} else {
			if err := r.Client.Delete(ctx, ingress); err != nil {
				return ctrl.Result{}, err
			} else {
				instance.ObjectMeta.Finalizers = []string{}
				if err := r.Update(ctx, instance); err != nil {
					return ctrl.Result{}, err
				}
				if err := DeleteCaddyRoute(ingress.Name, ingress.Namespace); err != nil {
					logger.Info("delete caddy route error !!!!!")
				}
			}
		}
		return reconcile.Result{}, nil
	}
	ingress := &networkingv1.Ingress{}
	if err := r.Client.Get(ctx, req.NamespacedName, ingress); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		logger.Info("add caddy route !!!!!")
		if err := AddCaddyRoute(instance); err != nil {
			logger.Info("add caddy route error !!!!!")
			return ctrl.Result{}, err
		}
		ingress = NewIngress(instance)
		logger.Info("create ingress !!!!!")
		if err := r.Client.Create(ctx, ingress); err != nil {
			if err := DeleteCaddyRoute(ingress.Name, ingress.Namespace); err != nil {
				logger.Info("delete caddy route error !!!!!")
			}
			return ctrl.Result{}, err
		} else {
			logger.Info("update App finalizer !!!!!")
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, instance.Name)
			if err := r.Client.Update(ctx, instance); err != nil {
				logger.Info("update finalizer err !!!!!")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		logger.Info("get Ingress exist !!!!!")
		newIngress := NewIngress(instance)
		if !reflect.DeepEqual(newIngress.Spec, ingress.Spec) {
			if err := DeleteCaddyRoute(ingress.Name, ingress.Namespace); err != nil {
				logger.Info("delete caddy route error !!!!!")
			}
			if err := AddCaddyRoute(instance); err != nil {
				logger.Info("add caddy route error !!!!!")
				return ctrl.Result{}, err
			}
			logger.Info("update Ingress !!!!!")
			if err := r.Client.Update(ctx, newIngress); err != nil {
				if err := DeleteCaddyRoute(newIngress.Name, newIngress.Namespace); err != nil {
					logger.Info("delete caddy route error !!!!!")
				}
				return ctrl.Result{}, err
			}
		} else {
			if err := DeleteCaddyRoute(ingress.Name, ingress.Namespace); err != nil {
				logger.Info("delete caddy route error !!!!!")
				return ctrl.Result{}, err
			}
			if err := AddCaddyRoute(instance); err != nil {
				logger.Info("add caddy route error !!!!!")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
}

func (r *StaticReconciler) DeleteIngress(event event.DeleteEvent, limiter workqueue.RateLimitingInterface) {
	name := event.Object.GetName()
	namespace := event.Object.GetNamespace()
	instance := &devopsv1.Static{}
	reqNamespaceName := types.NamespacedName{Name: name, Namespace: namespace}
	if err := r.Get(context.TODO(), reqNamespaceName, instance); err != nil {
		logger.Info(err.Error())
	} else {
		if err := r.Delete(context.TODO(), instance); err != nil {
			logger.Info(err.Error())
		}
		if err := DeleteCaddyRoute(instance.Name, instance.Namespace); err != nil {
			logger.Info(err.Error())
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *StaticReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Static{}).
		Watches(&source.Kind{
			Type: &networkingv1.Ingress{}},
			handler.Funcs{DeleteFunc: r.DeleteIngress}).
		Complete(r)
}
