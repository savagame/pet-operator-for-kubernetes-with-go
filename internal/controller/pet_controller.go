/*
Copyright 2025.

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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	petstorev1alpha1 "github.com/savagame/pet-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PetFinalizer = "pet.petstore.example.com"

// PetReconciler reconciles a Pet object
type PetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=petstore.example.com,resources=pets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=petstore.example.com,resources=pets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=petstore.example.com,resources=pets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *PetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	pet := &petstorev1alpha1.Pet{}

	if err := r.Get(ctx, req.NamespacedName, pet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch pet.Status.Phase {
	case "":
		return r.reconcilePending(ctx, pet)
	case petstorev1alpha1.PetPending:
		return r.reconcileCreate(ctx, pet)
	case petstorev1alpha1.PetReady:
		return ctrl.Result{}, nil
	case petstorev1alpha1.PetDeleting:
		return r.reconcileDelete(ctx, pet)
	case petstorev1alpha1.PetError:
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return ctrl.Result{}, nil
}

func (r *PetReconciler) reconcilePending(ctx context.Context, pet *petstorev1alpha1.Pet) (ctrl.Result, error) {
	pet.Status.Phase = petstorev1alpha1.PetPending
	return r.updateStatus(ctx, pet)
}

func (r *PetReconciler) reconcileDelete(ctx context.Context, pet *petstorev1alpha1.Pet) (ctrl.Result, error) {
	pet.Status.Phase = petstorev1alpha1.PetDeleting
	return r.updateStatus(ctx, pet)
}

func (r *PetReconciler) reconcileCreate(ctx context.Context, pet *petstorev1alpha1.Pet) (ctrl.Result, error) {
	if err := notifyExternalAPI(pet); err != nil {
		pet.Status.Phase = petstorev1alpha1.PetError
		return r.updateStatus(ctx, pet)
	}

	pet.Status.Phase = petstorev1alpha1.PetReady
	return r.updateStatus(ctx, pet)
}

func (r *PetReconciler) updateStatus(ctx context.Context, pet *petstorev1alpha1.Pet) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, pet); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PetReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.TypedOptions{
			MaxConcurrentReconciles: 5,
		}).
		For(&petstorev1alpha1.Pet{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Named("pet").
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)

}

func notifyExternalAPI(pet *petstorev1alpha1.Pet) error {
	body := map[string]string{
		"name": pet.Spec.Name,
		"type": string(pet.Spec.Type),
	}

	b, _ := json.Marshal(body)
	resp, err := http.Post(
		"https://httpbin.org/post",
		"application/json",
		bytes.NewBuffer(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func setCondition(pet *petstorev1alpha1.Pet, cond metav1.Condition) {
	meta.SetStatusCondition(&pet.Status.Conditions, cond)
}

func (r *PetReconciler) ensureConfigMap(ctx context.Context, pet *Pet) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pet.Name + "-config",
			Namespace: pet.Namespace,
		},
		Data: map[string]string{
			"petName": pet.Spec.Name,
		},
	}

	if err := controllerutil.SetControllerReference(pet, cm, r.Scheme); err != nil {
		return err
	}

	return r.Client.Patch(ctx, cm, client.Apply, client.ForceOwnership)
}
