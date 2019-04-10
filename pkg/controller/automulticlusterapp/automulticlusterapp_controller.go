package automulticlusterapp

import (
	"context"
	"fmt"

	rancheroperatorv1alpha1 "github.com/barpilot/rancher-operator/pkg/apis/rancheroperator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	managementrancherv3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

var log = logf.Log.WithName("controller_automulticlusterapp")

const (
	RancherGlobalNamespace = "cattle-global-data"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AutoMultiClusterApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) *ReconcileAutoMultiClusterApp {
	return &ReconcileAutoMultiClusterApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileAutoMultiClusterApp) error {
	// Create a new controller
	c, err := controller.New("automulticlusterapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AutoMultiClusterApp
	err = c.Watch(&source.Kind{Type: &rancheroperatorv1alpha1.AutoMultiClusterApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &managementrancherv3.Project{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			apps := []reconcile.Request{}

			// if _, ok := a.Meta.GetLabels()["autoproject/displayname"]; !ok {
			// 	log.Info("Project without good label")
			// 	return apps
			// }

			autoMultiClusterApps := &rancheroperatorv1alpha1.AutoMultiClusterAppList{}
			err := r.client.List(context.TODO(), &client.ListOptions{Namespace: ""}, autoMultiClusterApps)
			if err != nil {
				return apps
			}
			for _, app := range autoMultiClusterApps.Items {
				apps = append(apps, reconcile.Request{NamespacedName: types.NamespacedName{
					Name:      app.Name,
					Namespace: "",
				}})
			}
			return apps
		}),
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &managementrancherv3.MultiClusterApp{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			apps := []reconcile.Request{}

			autoMultiClusterApps := &rancheroperatorv1alpha1.AutoMultiClusterAppList{}
			err := r.client.List(context.TODO(), &client.ListOptions{Namespace: ""}, autoMultiClusterApps)
			if err != nil {
				return apps
			}
			for _, app := range autoMultiClusterApps.Items {
				apps = append(apps, reconcile.Request{NamespacedName: types.NamespacedName{
					Name:      app.Name,
					Namespace: "",
				}})
			}
			return apps
		}),
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAutoMultiClusterApp{}

// ReconcileAutoMultiClusterApp reconciles a AutoMultiClusterApp object
type ReconcileAutoMultiClusterApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AutoMultiClusterApp object and makes changes based on the state read
// and what is in the AutoMultiClusterApp.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAutoMultiClusterApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)

	ctx := context.TODO()

	// Our resource is cluster scoped
	request.NamespacedName.Namespace = ""

	reqLogger.Info("Reconciling AutoMultiClusterApp")

	// Fetch the AutoMultiClusterApp instance
	instance := &rancheroperatorv1alpha1.AutoMultiClusterApp{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	projects := &managementrancherv3.ProjectList{}

	opt := &client.ListOptions{}
	opt.InNamespace("")
	opt.SetLabelSelector(instance.Spec.ProjectSelector)

	if err := r.client.List(ctx, opt, projects); err != nil {
		reqLogger.Info("Failed to list projects")
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &managementrancherv3.MultiClusterApp{}
	err = r.client.Get(ctx, types.NamespacedName{Name: instance.Spec.MultiClusterApp, Namespace: RancherGlobalNamespace}, found)
	if err == nil {
		reqLogger.Info("Updating multiClusterApp", "App", found.Name)

		for _, project := range projects.Items {
			targetName := fmt.Sprintf("%s:%s", project.Spec.ClusterName, project.Name)

			alreadyThere := false
			for _, target := range found.Spec.Targets {
				if target.ProjectName == targetName {
					alreadyThere = true
					break
				}
			}

			if !alreadyThere {
				reqLogger.Info("Add target", "targetName", targetName)
				found.Spec.Targets = append(found.Spec.Targets, managementrancherv3.Target{ProjectName: targetName})
			}
		}

		if err := r.client.Update(ctx, found); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if errors.IsNotFound(err) {
		reqLogger.Info("multiClusterApp doesn't exists", "App", instance.Spec.MultiClusterApp)
	}
	return reconcile.Result{}, err
}
