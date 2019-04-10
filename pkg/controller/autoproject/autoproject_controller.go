package autoproject

import (
	"context"

	rancheroperatorv1alpha1 "github.com/barpilot/rancher-operator/pkg/apis/rancheroperator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	managementrancherv3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

var log = logf.Log.WithName("controller_autoproject")

// Add creates a new AutoProject Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) *ReconcileAutoProject {
	return &ReconcileAutoProject{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileAutoProject) error {
	// Create a new controller
	c, err := controller.New("autoproject-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AutoProject
	err = c.Watch(&source.Kind{Type: &rancheroperatorv1alpha1.AutoProject{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner AutoProject
	err = c.Watch(&source.Kind{Type: &managementrancherv3.Cluster{}}, &handler.EnqueueRequestsFromMapFunc{
		// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			apps := []reconcile.Request{}

			autoProjects := &rancheroperatorv1alpha1.AutoProjectList{}
			err := r.client.List(context.TODO(), &client.ListOptions{Namespace: ""}, autoProjects)
			if err != nil {
				return apps
			}
			for _, app := range autoProjects.Items {
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

	err = c.Watch(&source.Kind{Type: &managementrancherv3.Project{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rancheroperatorv1alpha1.AutoProject{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAutoProject{}

// ReconcileAutoProject reconciles a AutoProject object
type ReconcileAutoProject struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AutoProject object and makes changes based on the state read
// and what is in the AutoProject.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAutoProject) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)

	ctx := context.TODO()

	// Our resource is cluster scoped
	request.NamespacedName.Namespace = ""
	reqLogger.Info("Reconciling AutoProject")

	// Fetch the AutoProject instance
	instance := &rancheroperatorv1alpha1.AutoProject{}

	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Failed to get instance")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Info("Failed to get instance")
		return reconcile.Result{}, err
	}

	clusters := &managementrancherv3.ClusterList{}

	opt := &client.ListOptions{}
	opt.InNamespace("")

	if err := r.client.List(ctx, opt, clusters); err != nil {
		reqLogger.Info("Failed to list clusters")
		return reconcile.Result{}, err
	}

	if len(clusters.Items) == 0 {
		reqLogger.Info("empty cluster list")
	}

	for _, cluster := range clusters.Items {
		project := newProjectForCR(instance, cluster.Name)

		if err := controllerutil.SetControllerReference(instance, project, r.scheme); err != nil {
			reqLogger.Info("Failed to ser owner")
			return reconcile.Result{}, err
		}

		projects := &managementrancherv3.ProjectList{}

		opt := &client.ListOptions{}
		opt.InNamespace(cluster.Name)
		opt.MatchingLabels(map[string]string{"autoproject/displayname": instance.Spec.ProjectSpec.DisplayName})

		if err := r.client.List(ctx, opt, projects); err != nil {
			reqLogger.Info("Failed to list projects")
			return reconcile.Result{}, err
		}

		if len(projects.Items) == 0 {
			reqLogger.Info("Creating a new Project", "Project.Namespace", project.Namespace, "Project.Name", project.Name)
			if err := r.client.Create(ctx, project); err != nil {
				return reconcile.Result{}, err
			}
		} else if len(projects.Items) == 1 {
			reqLogger.Info("Skip reconcile: Project already exists", "Project.Namespace", projects.Items[0].Namespace, "Project.Name", projects.Items[0].Name)
		} else {
			reqLogger.Info("! Skip reconcile: multiple Projects already exists!")
		}
	}
	return reconcile.Result{}, nil
}

func newProjectForCR(cr *rancheroperatorv1alpha1.AutoProject, clusterName string) *managementrancherv3.Project {

	projectSpec := cr.Spec.ProjectSpec.DeepCopy()
	projectSpec.ClusterName = clusterName

	return &managementrancherv3.Project{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    clusterName,
			GenerateName: "p-",
			Labels: map[string]string{
				"autoproject/displayname": projectSpec.DisplayName,
			},
		},
		Spec: *projectSpec,
	}
}
