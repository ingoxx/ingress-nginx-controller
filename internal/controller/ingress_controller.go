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
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/controller/store"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/nginx"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/pkg/resources"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ingress.nginx.kubebuilder.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingress.nginx.kubebuilder.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingress.nginx.kubebuilder.io,resources=ingresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logger := log.FromContext(ctx)

	// TODO(user): your logic here
	var ic = new(ingressv1.Ingress)
	if err := r.Get(ctx, req.NamespacedName, ic); err != nil {
		klog.ErrorS(err, fmt.Sprintf("unable to fetch Ingress in namesapce %s.", req.Namespace))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(ic.Spec.Rules) == 0 && ic.Spec.DefaultBackend == nil {
		klog.Info(fmt.Sprintf("due to the empty values of spec field, the ingress controller will not take any action, ingress name %s", req.Name))
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(11)}, nil
	}

	var ns client.ObjectKey
	if ic.Spec.DefaultBackend != nil {
		ns = types.NamespacedName{Name: ic.Spec.DefaultBackend.Service.Name, Namespace: ic.Namespace}
		if err := r.checkService(ctx, ns); err != nil {
			klog.Fatal(err)
		}
	}

	if len(ic.Spec.Rules) > 0 {
		for _, v := range ic.Spec.Rules {
			for _, h := range v.HTTP.Paths {
				ns = types.NamespacedName{Name: h.Backend.Service.Name, Namespace: ic.Namespace}
				if err := r.checkService(ctx, ns); err != nil {
					klog.Fatal(err)
				}
			}
		}
	}

	rs := r.GetReconcilerInfo()
	rs.Context = ctx
	rs.Ingress = ic
	rs.IngressInfos = store.NewIngressInfo(rs)

	if err := resources.ReconcileResource(rs); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(15)}, nil
	}

	parsed, err := annotations.NewAnnotationExtractor(rs.IngressInfos).Extract(ic)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("fail to parse annotations, ingress: %s, namespace: %s", req.Namespace, req.Namespace))
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(15)}, nil
	}

	var ings = annotations.IngressAnnotations{
		Ingress:           ic,
		ParsedAnnotations: parsed,
	}

	if err := NewNginxConfigure(rs).GenerateNginxConfigure(ings); err != nil {
		klog.ErrorS(err, fmt.Sprintf("fail to generate nginx configure, ingress: %s, namespace: %s", req.Namespace, req.Namespace))
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(15)}, nil
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) GetReconcilerInfo() *store.IngressReconciler {
	si := &store.IngressReconciler{
		Client: r.Client,
		Scheme: r.Scheme,
	}

	return si
}

func (r *IngressReconciler) checkService(ctx context.Context, key client.ObjectKey) error {
	svc := new(v1.Service)
	if err := r.Get(ctx, key, svc); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("no service with name %v found in namespace %v: %v", key.Name, key.Namespace, err)
		}

		return fmt.Errorf("unexpected error searching service with name %v in namespace %v: %v", key.Name, key.Namespace, err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	go nginx.Start()

	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1.Ingress{}).
		Complete(r)
}
