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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Ingress) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	ingresslog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Ingress) ValidateDelete() (admission.Warnings, error) {
	ingresslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
