package resolver

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resolver interface {
	GetSecret(client.ObjectKey) (*corev1.Secret, error)
	GetDefaultService() (*corev1.Service, error)
	GetService(string) (*corev1.Service, error)
	GetHostName() []string
	GetSvcPort(netv1.IngressBackend) int32
	GetTlsData(client.ObjectKey) (map[string][]byte, error)
}
