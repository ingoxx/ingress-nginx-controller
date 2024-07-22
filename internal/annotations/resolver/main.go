package resolver

import (
	"context"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resolver interface {
	GetSecret() (map[string]byte, error)
	GetDefaultService() (*corev1.Service, error)
	GetService(string) (*corev1.Service, error)
	GetHostName() []string
}

type IngressInfo struct {
	ingress *ingressv1.Ingress
	r       client.Client
	ctx     context.Context
}

func NewIngressInfo(ingress *ingressv1.Ingress, r client.Client, ctx context.Context) *IngressInfo {
	return &IngressInfo{
		ingress: ingress,
		r:       r,
		ctx:     ctx,
	}
}

func (t *IngressInfo) GetSecret() (map[string]byte, error) {
	return nil, nil
}

func (t *IngressInfo) getDnsTlsFile() {

}

func (t *IngressInfo) GetHostName() []string {
	hosts := make([]string, 0)
	if len(t.ingress.Spec.Rules) > 0 {
		for _, v := range t.ingress.Spec.Rules {
			hosts = append(hosts, v.Host)
		}
	} else {
		service, err := t.GetDefaultService()
		if err != nil {
			return hosts
		}

		dns := service.Name + "." + service.Namespace
		clusterDns := dns + "cluster.svc"
		localDns := dns + ".svc"

		hosts = []string{clusterDns, localDns}
	}

	return hosts
}

func (t *IngressInfo) GetDefaultService() (*corev1.Service, error) {
	svc := new(corev1.Service)
	svcName := t.ingress.Spec.DefaultBackend.Service.Name
	if err := t.r.Get(t.ctx, types.NamespacedName{Name: svcName, Namespace: t.ingress.Namespace}, svc); err != nil {
		return svc, err
	}

	return svc, nil
}

func (t *IngressInfo) GetService(name string) (*corev1.Service, error) {
	svc := new(corev1.Service)
	if err := t.r.Get(t.ctx, types.NamespacedName{Name: name, Namespace: t.ingress.Namespace}, svc); err != nil {
		if errors.IsNotFound(err) {
			return svc, fmt.Errorf("service: %s not fount in namespace: %s", name, t.ingress.Namespace)
		}

		return svc, fmt.Errorf("unexpected error searching service with name %v in namespace %v: %v", name, t.ingress.Namespace, err)
	}

	return svc, nil
}
