package store

import (
	"context"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	utils "github.com/Lxb921006/ingress-nginx-kubebuilder/pkg/utils/cert"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IngressInfo struct {
	ingress *ingressv1.Ingress
	r       client.Client
	ctx     context.Context
}

func NewIngressInfo(store Storer) *IngressInfo {
	c := store.GetReconcilerInfo()
	return &IngressInfo{
		ingress: c.Ingress,
		r:       c.Client,
		ctx:     c.Context,
	}
}

func (t *IngressInfo) getDnsTlsFile() {

}

func (t *IngressInfo) GetHostName() []string {
	hosts := make([]string, 0)
	for _, v := range t.ingress.Spec.Rules {
		hosts = append(hosts, v.Host)
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

func (t *IngressInfo) GetSvcPort(backend netv1.IngressBackend) int32 {
	var port int32
	svc, err := t.GetService(backend.Service.Name)
	if err != nil {
		return port
	}

	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Port == backend.Service.Port.Number || svcPort.Name == backend.Service.Port.Name {
			port = backend.Service.Port.Number
			break
		}
	}

	return port
}

func (t *IngressInfo) GetSecret(key client.ObjectKey) (*corev1.Secret, error) {
	sc := new(corev1.Secret)

	if err := t.r.Get(t.ctx, key, sc); err != nil {
		if errors.IsNotFound(err) {
			return sc, fmt.Errorf("secret: %s not fount in namespace: %s", t.ingress.Name+"-secret", t.ingress.Namespace)
		}

		return sc, fmt.Errorf("unexpected error searching service with name %v in namespace %v: %v", t.ingress.Name+"-secret", t.ingress.Namespace, err)
	}

	return sc, nil
}

func (t *IngressInfo) GetTlsData(key client.ObjectKey) (map[string][]byte, error) {
	var data map[string][]byte

	secret, err := t.GetSecret(key)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("fail to get secret: %s, in namespace: %s", t.ingress.Name+"-secret", t.ingress.Namespace))
		return data, err
	}

	data, err = utils.DecodeBase64(secret.Data)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("decoding tls failed, secret: %s, in namespace: %s", t.ingress.Name+"-secret", t.ingress.Namespace))
		return data, err
	}

	return data, nil
}
