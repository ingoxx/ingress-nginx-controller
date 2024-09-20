package store

import (
	"context"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Storer interface {
	ReconcilerInfo() *IngressReconciler
	Render(data interface{}) ([]byte, error)
	Generate(name string, b []byte) error
}

type IngressReconciler struct {
	Client           client.Client
	Scheme           *runtime.Scheme
	Ingress          *ingressv1.Ingress
	Context          context.Context
	IngressInfos     *IngressInfo
	DynamicClientSet *dynamic.DynamicClient
}

func (i *IngressReconciler) ReconcilerInfo() *IngressReconciler {
	return i
}

func (i *IngressReconciler) Render(data interface{}) ([]byte, error) {
	return nil, nil
}

func (i *IngressReconciler) Generate(name string, b []byte) error {
	return nil
}
