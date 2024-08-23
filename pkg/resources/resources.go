package resources

import (
	"context"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/controller/store"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resources struct {
	dynamicClientSet *dynamic.DynamicClient
	client           client.Client
	ingress          *ingressv1.Ingress
	sc               *runtime.Scheme
	ctx              context.Context
}

// ReconcileResource If the spec.tls field is not empty, the certificate and issuer resources here will not be created.
func ReconcileResource(store store.Storer) error {
	ctlInfo := store.GetReconcilerInfo()
	r := NewResource(ctlInfo)

	if len(r.ingress.Spec.Rules) == 0 {
		return nil
	}

	if len(r.ingress.Spec.TLS) == 0 {
		if err := r.reconcileIssuer(); err != nil {
			klog.ErrorS(err, "fail to reconcile issuer resource")
			return err
		}

		if err := r.reconcileCert(ctlInfo.IngressInfos); err != nil {
			klog.ErrorS(err, "fail to reconcile certificate resource")
			return err
		}
	}

	return nil
}

func NewResource(ctlInfo *store.IngressReconciler) *Resources {
	return &Resources{
		dynamicClientSet: ctlInfo.DynamicClientSet,
		client:           ctlInfo.Client,
		ingress:          ctlInfo.Ingress,
		ctx:              ctlInfo.Context,
		sc:               ctlInfo.Scheme,
	}
}

func (t *Resources) reconcileCert(rr resolver.Resolver) error {
	certGVK := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
		},
	}

	if err := t.client.Get(t.ctx, types.NamespacedName{Name: t.ingress.Name + "-cert", Namespace: t.ingress.Namespace}, certificate); err != nil {
		if kerrs.IsNotFound(err) {
			hosts := rr.GetHostName()
			createCert := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cert-manager.io/v1",
					"kind":       "Certificate",
					"metadata": map[string]interface{}{
						"name":      t.ingress.Name + "-cert",
						"namespace": t.ingress.Namespace,
					},
					"spec": map[string]interface{}{
						"dnsNames": hosts,
						"issuerRef": map[string]interface{}{
							"kind": "Issuer",
							"name": t.ingress.Name + "-issuer",
						},
						"secretName": t.ingress.Name + "-secret",
					},
				},
			}
			_, err = t.dynamicClientSet.Resource(certGVK).Namespace(t.ingress.Namespace).Create(context.Background(), createCert, metav1.CreateOptions{})
			if err != nil {
				return err
			}

			if _, err := rr.GetSecret(types.NamespacedName{Name: t.ingress.Name + "-secret", Namespace: t.ingress.Namespace}); err != nil {
				return err
			}

			return nil
		}
		return err
	}

	domains, found, err := unstructured.NestedStringSlice(certificate.Object, "spec", "dnsNames")
	if !found || err != nil {
		return err
	}

	if len(t.ingress.Spec.Rules) != len(domains) {
		hosts := rr.GetHostName()
		if err := unstructured.SetNestedStringSlice(certificate.Object, hosts, "spec", "dnsNames"); err != nil {
			return err
		}

		if _, err := t.dynamicClientSet.Resource(certGVK).Namespace(t.ingress.Namespace).Update(context.TODO(), certificate, metav1.UpdateOptions{}); err != nil {
			return err
		}

		if _, err := rr.GetSecret(types.NamespacedName{Name: t.ingress.Name + "-secret", Namespace: t.ingress.Namespace}); err != nil {
			return err
		}

		klog.Infof("update certificate: %s successfully", t.ingress.Name+"-cert")
	}

	return nil
}

func (t *Resources) reconcileIssuer() error {
	issuerGVK := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "issuers"}
	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Issuer",
		},
	}

	if err := t.client.Get(t.ctx, types.NamespacedName{Name: t.ingress.Name + "-issuer", Namespace: t.ingress.Namespace}, issuer); err != nil {
		if kerrs.IsNotFound(err) {
			createIssuer := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cert-manager.io/v1",
					"kind":       "Issuer",
					"metadata": map[string]interface{}{
						"name":      t.ingress.Name + "-issuer",
						"namespace": t.ingress.Namespace,
					},
					"spec": map[string]interface{}{
						"selfSigned": map[string]interface{}{},
					},
				},
			}
			_, err = t.dynamicClientSet.Resource(issuerGVK).Namespace(t.ingress.Namespace).Create(context.Background(), createIssuer, metav1.CreateOptions{})
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}
