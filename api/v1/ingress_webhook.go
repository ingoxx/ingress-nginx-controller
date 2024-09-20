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

package v1

import (
	"fmt"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	proxyPathAnnotation = "ingress.nginx.kubebuilder.io/proxy-host"
)

// log is for logging in this package.
var ingresslog = logf.Log.WithName("ingress-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Ingress) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-ingress-nginx-kubebuilder-io-v1-ingress,mutating=true,failurePolicy=fail,sideEffects=None,groups=ingress.nginx.kubebuilder.io,resources=ingresses,verbs=create;update,versions=v1,name=mingress.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Ingress{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Ingress) Default() {
	ingresslog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ingress-nginx-kubebuilder-io-v1-ingress,mutating=false,failurePolicy=fail,sideEffects=None,groups=ingress.nginx.kubebuilder.io,resources=ingresses,verbs=create;update,versions=v1,name=vingress.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Ingress{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Ingress) ValidateCreate() (admission.Warnings, error) {
	ingresslog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.ValidData()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Ingress) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	ingresslog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, r.ValidData()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Ingress) ValidateDelete() (admission.Warnings, error) {
	ingresslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *Ingress) ValidData() error {
	if err := r.ValidSpec(); err != nil {
		return err
	}

	if err := r.ValidPathAndHost(); err != nil {
		return err
	}

	return nil
}

func (r *Ingress) ValidSpec() error {
	if r.Spec.DefaultBackend == nil && len(r.Spec.Rules) == 0 {
		return fmt.Errorf("no available backend found in ingress: %s, namespace: %s", r.Name, r.Namespace)
	}

	return nil
}

func (r *Ingress) ValidPathAndHost() error {
	for _, v := range r.Spec.Rules {
		proxyHost, ok := r.Annotations[proxyPathAnnotation]
		if ok {
			if !r.ValidHost(v.Host) {
				return fmt.Errorf("proxy-host: %s is an invalid value in ingress: %s, namespace: %s", proxyHost, r.Name, r.Namespace)
			}
		}
		var path string
		proxyPath, ok := r.Annotations[proxyPathAnnotation]
		if ok {
			pp := v1.HTTPIngressPath{
				Path: proxyPath,
			}
			v.IngressRuleValue.HTTP.Paths = append(v.IngressRuleValue.HTTP.Paths, pp)
		}
		for _, p := range v.IngressRuleValue.HTTP.Paths {
			if p.Path == path {
				return fmt.Errorf("not allow duplicate path: %s in ingress: %s, namespace: %s", path, r.Name, r.Namespace)
			}
			path = p.Path
		}
	}
	return nil
}

func (r *Ingress) ValidHost(str string) bool {
	validHost := func(str string) bool {
		p := `([a0-z9]+\.)+([a-z]+)`
		matched := regexp.MustCompile(p)
		if !matched.MatchString(str) {
			return false
		}
		return true
	}

	return validHost(str)
}
