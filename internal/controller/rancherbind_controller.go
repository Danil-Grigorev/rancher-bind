/*
Copyright 2023 SUSE.

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

	"github.com/kube-bind/kube-bind/contrib/example-backend/controllers/serviceexport"
	"github.com/kube-bind/kube-bind/contrib/example-backend/controllers/serviceexportrequest"
	"github.com/kube-bind/kube-bind/contrib/example-backend/controllers/servicenamespace"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubebindv1alpha1 "github.com/kube-bind/kube-bind/pkg/apis/kubebind/v1alpha1"
)

type ThreadedRunable interface {
	// Start starts running the component.  The component will stop running
	// when the context is closed. Start blocks until the context is closed or
	// an error occurs.
	Start(context.Context, int)
}

type Threaded struct {
	threads  int
	runnable ThreadedRunable
}

func NewThreaded(runnable ThreadedRunable, threads int) *Threaded {
	return &Threaded{
		threads:  threads,
		runnable: runnable,
	}
}

func (t *Threaded) Start(ctx context.Context) error {
	t.runnable.Start(ctx, t.threads)
	return nil
}

type InformerRunnable interface {
	Start(stopCh <-chan struct{})
}

type Informer struct {
	informer InformerRunnable
}

func NewInformer(informer InformerRunnable) *Informer {
	return &Informer{
		informer: informer,
	}
}

func (i *Informer) Start(ctx context.Context) error {
	i.informer.Start(ctx.Done())
	return nil
}

// RancherBindReconciler reconciles a RancherBind object
type RancherBindReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservicebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservicebindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservicebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiserviceexportrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiserviceexportrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiserviceexportrequests/finalizers,verbs=update
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiserviceexports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiserviceexports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiserviceexports/finalizers,verbs=update
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservicenamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservicenamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservicenamespaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube-bind.io,resources=apiservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=kube-bind.io,resources=clusterbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube-bind.io,resources=clusterbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube-bind.io,resources=clusterbindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RancherBind object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *RancherBindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RancherBindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	config, err := NewConfig()
	if err != nil {
		return fmt.Errorf("unable to get config: %w", err)
	}

	serviceNamespace, err := servicenamespace.NewController(
		config.ClientConfig,
		kubebindv1alpha1.ClusterScope,
		config.BindInformers.KubeBind().V1alpha1().APIServiceNamespaces(),
		config.BindInformers.KubeBind().V1alpha1().ClusterBindings(),
		config.BindInformers.KubeBind().V1alpha1().APIServiceExports(),
		config.KubeInformers.Core().V1().Namespaces(),
		config.KubeInformers.Rbac().V1().Roles(),
		config.KubeInformers.Rbac().V1().RoleBindings(),
	)
	if err != nil {
		return fmt.Errorf("error setting up APIServiceNamespace Controller: %w", err)
	}
	serviceExport, err := serviceexport.NewController(
		config.ClientConfig,
		config.BindInformers.KubeBind().V1alpha1().APIServiceExports(),
		config.ApiextensionsInformers.Apiextensions().V1().CustomResourceDefinitions(),
	)
	if err != nil {
		return fmt.Errorf("error setting up APIServiceExport Controller: %w", err)
	}

	serviceExportRequest, err := serviceexportrequest.NewController(
		config.ClientConfig,
		kubebindv1alpha1.NamespacedScope,
		config.BindInformers.KubeBind().V1alpha1().APIServiceExportRequests(),
		config.BindInformers.KubeBind().V1alpha1().APIServiceExports(),
		config.ApiextensionsInformers.Apiextensions().V1().CustomResourceDefinitions(),
	)
	if err != nil {
		return fmt.Errorf("error setting up ServiceExportRequest Controller: %w", err)
	}

	// start informer factories
	mgr.Add(NewInformer(config.KubeInformers))
	mgr.Add(NewInformer(config.BindInformers))
	mgr.Add(NewInformer(config.ApiextensionsInformers))

	mgr.Add(NewThreaded(serviceNamespace, 1))
	mgr.Add(NewThreaded(serviceExport, 1))
	mgr.Add(NewThreaded(serviceExportRequest, 1))

	return nil
}
