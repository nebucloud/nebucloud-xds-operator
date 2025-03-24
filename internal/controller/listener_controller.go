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
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	xdsv1alpha1 "github.com/nebucloud/nebucloud-xds-operator/api/v1alpha1"
	"github.com/nebucloud/nebucloud-xds-operator/pkg/xds"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListenerReconciler reconciles a Listener object
type ListenerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	XDSClient xds.Client
}

// +kubebuilder:rbac:groups=xds.xds.nebucloud.io,resources=listeners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=xds.xds.nebucloud.io,resources=listeners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=xds.xds.nebucloud.io,resources=listeners/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Listener object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ListenerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("listener", req.NamespacedName)
	logger.Info("Reconciling Listener")

	// Fetch the Listener resource
	listener := &xdsv1alpha1.Listener{}
	err := r.Get(ctx, req.NamespacedName, listener)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Resource deleted - clean up resources
			logger.Info("Listener resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Listener")
		return ctrl.Result{}, err
	}

	// Initialize status if needed
	if listener.Status.Conditions == nil {
		listener.Status.Conditions = []xdsv1alpha1.ListenerCondition{}
	}

	// Set finalizer if not set
	finalizerName := "listeners.xds.nebucloud.io/finalizer"
	if !controllerutil.ContainsFinalizer(listener, finalizerName) {
		controllerutil.AddFinalizer(listener, finalizerName)
		if err := r.Update(ctx, listener); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil // Requeue to continue after update
	}

	// Handle deletion
	if !listener.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, listener, finalizerName)
	}

	// Set defaults if needed
	if listener.Spec.Address == "" {
		listener.Spec.Address = "0.0.0.0" // Default to all interfaces
		if err := r.Update(ctx, listener); err != nil {
			logger.Error(err, "Failed to set default address")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil // Requeue to continue after update
	}

	// Validate the resource
	if err := r.validateListener(listener); err != nil {
		logger.Error(err, "Listener validation failed")
		r.setStatusCondition(ctx, listener, xdsv1alpha1.ListenerCondition{
			Type:    "Ready",
			Status:  false,
			Reason:  "ValidationFailed",
			Message: err.Error(),
		})
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Convert to xDS resource
	xdsResource, err := r.convertToXDS(listener)
	if err != nil {
		logger.Error(err, "Failed to convert Listener to xDS resource")
		r.setStatusCondition(ctx, listener, xdsv1alpha1.ListenerCondition{
			Type:    "Ready",
			Status:  false,
			Reason:  "ConversionFailed",
			Message: fmt.Sprintf("Failed to convert to xDS: %v", err),
		})
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Update xDS server with the resource
	if err := r.XDSClient.UpdateResource(ctx, xdsResource); err != nil {
		logger.Error(err, "Failed to update xDS server")
		r.setStatusCondition(ctx, listener, xdsv1alpha1.ListenerCondition{
			Type:    "Ready",
			Status:  false,
			Reason:  "SyncFailed",
			Message: fmt.Sprintf("Failed to sync with xDS server: %v", err),
		})
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Update status to indicate success
	r.setStatusCondition(ctx, listener, xdsv1alpha1.ListenerCondition{
		Type:    "Ready",
		Status:  true,
		Reason:  "Synced",
		Message: "Successfully synced with xDS server",
	})

	// Reconcile again after some time to ensure resources are still in sync
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// handleDeletion handles the deletion of a Listener resource
func (r *ListenerReconciler) handleDeletion(ctx context.Context, listener *xdsv1alpha1.Listener, finalizerName string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Delete the resource from xDS server
	if err := r.XDSClient.DeleteResource(ctx, "listener", listener.Spec.Name); err != nil {
		logger.Error(err, "Failed to delete resource from xDS server")
		return ctrl.Result{}, err
	}

	// Remove finalizer to allow deletion
	controllerutil.RemoveFinalizer(listener, finalizerName)
	if err := r.Update(ctx, listener); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully cleaned up resources for deleted Listener")
	return ctrl.Result{}, nil
}

// validateListener validates the Listener resource
func (r *ListenerReconciler) validateListener(listener *xdsv1alpha1.Listener) error {
	// Basic validation - could be extended with CUE validation later
	if listener.Spec.Name == "" {
		return fmt.Errorf("listener name is required")
	}
	if listener.Spec.Port < 1 || listener.Spec.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	// Validate each route
	for _, route := range listener.Spec.Routes {
		if route.Name == "" {
			return fmt.Errorf("route name is required")
		}
		if route.Match.Prefix == "" && route.Match.Path == "" && route.Match.Regex == "" {
			return fmt.Errorf("route match requires at least one of prefix, path, or regex")
		}
		if route.Route.Cluster == "" {
			return fmt.Errorf("route target cluster is required")
		}
	}

	return nil
}

// convertToXDS converts a Kubernetes Listener to an xDS resource
func (r *ListenerReconciler) convertToXDS(listener *xdsv1alpha1.Listener) (*xds.Resource, error) {
	// This is a simplified example - actual conversion would depend on your xDS client implementation
	resource := &xds.Resource{
		Type: "listener",
		Name: listener.Spec.Name,
		Data: map[string]interface{}{
			"address": listener.Spec.Address,
			"port":    listener.Spec.Port,
		},
	}

	// Add routes if present
	if len(listener.Spec.Routes) > 0 {
		routes := make([]map[string]interface{}, 0, len(listener.Spec.Routes))
		for _, route := range listener.Spec.Routes {
			routeMap := map[string]interface{}{
				"name": route.Name,
				"match": map[string]interface{}{
					"prefix": route.Match.Prefix,
					"path":   route.Match.Path,
					"regex":  route.Match.Regex,
				},
				"route": map[string]interface{}{
					"cluster": route.Route.Cluster,
				},
			}

			// Add timeout if specified
			if route.Route.Timeout != "" {
				routeMap["route"].(map[string]interface{})["timeout"] = route.Route.Timeout
			}

			routes = append(routes, routeMap)
		}
		resource.Data["routes"] = routes
	}

	// Add TLS config if present
	if listener.Spec.TLS != nil {
		resource.Data["tls"] = map[string]interface{}{
			"mode": listener.Spec.TLS.Mode,
		}

		if listener.Spec.TLS.CertificateRef != nil {
			resource.Data["tls"].(map[string]interface{})["certificateRef"] = map[string]string{
				"name":      listener.Spec.TLS.CertificateRef.Name,
				"namespace": listener.Spec.TLS.CertificateRef.Namespace,
			}
		}
	}

	return resource, nil
}

// setStatusCondition updates the status conditions for the Listener
func (r *ListenerReconciler) setStatusCondition(ctx context.Context, listener *xdsv1alpha1.Listener, condition xdsv1alpha1.ListenerCondition) {
	logger := log.FromContext(ctx)

	now := metav1.Now()
	condition.LastTransitionTime = &now

	// Find and update existing condition or append new one
	found := false
	for i, cond := range listener.Status.Conditions {
		if cond.Type == condition.Type {
			if cond.Status != condition.Status {
				condition.LastTransitionTime = &now
			} else {
				condition.LastTransitionTime = cond.LastTransitionTime
			}
			listener.Status.Conditions[i] = condition
			found = true
			break
		}
	}

	if !found {
		listener.Status.Conditions = append(listener.Status.Conditions, condition)
	}

	// Update status subresource
	listener.Status.ObservedGeneration = listener.Generation
	if err := r.Status().Update(ctx, listener); err != nil {
		logger.Error(err, "Failed to update Listener status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ListenerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&xdsv1alpha1.Listener{}).
		Complete(r)
}
