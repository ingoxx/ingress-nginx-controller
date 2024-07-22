package resources

import (
	"context"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resources struct {
	dynamicClientSet *dynamic.DynamicClient
	client           client.Client
	ingress          *ingressv1.Ingress
	ctx              context.Context
}

func ReconcileResource(client client.Client, ing *ingressv1.Ingress, ctx context.Context, rr resolver.Resolver) error {
	r := NewResource(client, ing, ctx)
	if err := r.reconcileCert(rr); err != nil {
		klog.ErrorS(err, "fail to reconcile certificate resource")
		return err
	}

	if err := r.reconcileIssuer(); err != nil {
		klog.ErrorS(err, "fail to reconcile issuer resource")
		return err
	}

	return nil
}

func NewResource(client client.Client, ing *ingressv1.Ingress, ctx context.Context) *Resources {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("fail to create InClusterConfig: %v", err)
		}
		config = inClusterConfig
	}

	clientSet, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Fatalf("fail to create clientSet: %v", err)
	}

	return &Resources{
		dynamicClientSet: clientSet,
		client:           client,
		ingress:          ing,
		ctx:              ctx,
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

			return nil
		}
		return err
	}

	spec, found, err := unstructured.NestedMap(certificate.Object, "spec")
	if !found || err != nil {
		return err
	}

	domains, found, err := unstructured.NestedStringSlice(spec, "dnsNames")
	if !found || err != nil {
		return err
	}

	if len(t.ingress.Spec.Rules) != len(domains) && t.ingress.Spec.DefaultBackend == nil {
		hosts := rr.GetHostName()
		if err := unstructured.SetNestedStringSlice(spec, hosts, "dnsNames"); err != nil {
			return err
		}

		if _, err := t.dynamicClientSet.Resource(certGVK).Namespace(t.ingress.Namespace).Update(context.Background(), certificate, metav1.UpdateOptions{}); err != nil {
			return err
		}
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
