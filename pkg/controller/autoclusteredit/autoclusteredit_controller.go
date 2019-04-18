package autoclusteredit

import (
	"context"

	rancherv3 "github.com/rancher/types/apis/management.cattle.io/v3"

	rancheroperatorv1alpha1 "github.com/barpilot/rancher-operator/pkg/apis/rancheroperator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/imdario/mergo"
)

var log = logf.Log.WithName("controller_autoclusteredit")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AutoClusterEdit Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAutoClusterEdit{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("autoclusteredit-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AutoClusterEdit
	err = c.Watch(&source.Kind{Type: &rancheroperatorv1alpha1.AutoClusterEdit{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AutoClusterEdit
	err = c.Watch(&source.Kind{Type: &rancherv3.Cluster{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			apps := []reconcile.Request{}

			autoClusterEdits := &rancheroperatorv1alpha1.AutoClusterEditList{}
			err := r.client.List(context.TODO(), &client.ListOptions{Namespace: ""}, autoClusterEdits)
			if err != nil {
				return apps
			}
			for _, autoClusterEdit := range autoClusterEdits.Items {
				apps = append(apps, reconcile.Request{NamespacedName: types.NamespacedName{
					Name:      autoClusterEdit.Name,
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

var _ reconcile.Reconciler = &ReconcileAutoClusterEdit{}

// ReconcileAutoClusterEdit reconciles a AutoClusterEdit object
type ReconcileAutoClusterEdit struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AutoClusterEdit object and makes changes based on the state read
// and what is in the AutoClusterEdit.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAutoClusterEdit) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AutoClusterEdit")

	// Fetch the AutoClusterEdit instance
	instance := &rancheroperatorv1alpha1.AutoClusterEdit{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set AutoClusterEdit instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	clusters := &managementrancherv3.ClusterList{}

	opt := &client.ListOptions{}
	opt.InNamespace("")
	opt.SetFieldSelector(instance.Spec.ClusterSelector)

	if err := r.client.List(ctx, opt, clusters); err != nil {
		reqLogger.Info("Failed to list clusters")
		return reconcile.Result{}, err
	}

	for _, cluster := range clusters.Items {
		if err := mergo.Merge(&cluster, instance.Spec.ClusterTemplate); err != nil {
			reqLogger.Error(err, "failed to merge cluster with clusterTemplate")
			continue
		}
		if err = r.client.Update(context.TODO(), cluster); err != nil {
			reqLogger.Error(err, "failed to update cluster")
			continue
		}

	}
	return reconcile.Result{}, nil
}
