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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ListenerSpec defines the desired state of Listener
type ListenerSpec struct {
	// Name is the unique identifier for this listener in the xDS API
	Name string `json:"name"`

	// Address is the address the listener should bind to (e.g., 0.0.0.0 for all interfaces)
	// +optional
	Address string `json:"address,omitempty"`

	// Port is the port the listener should bind to
	Port int32 `json:"port"`

	// TLS configuration for the listener
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`

	// Routes configurations attached to this listener
	// +optional
	Routes []Route `json:"routes,omitempty"`
}

// TLSConfig defines TLS configuration for a listener
type TLSConfig struct {
	// Mode specifies the TLS mode (DISABLED, SIMPLE, MUTUAL)
	Mode string `json:"mode"`

	// CertificateRef references a Kubernetes secret containing TLS certificates
	// +optional
	CertificateRef *SecretReference `json:"certificateRef,omitempty"`
}

// SecretReference contains the name and namespace of a Kubernetes secret
type SecretReference struct {
	// Name is the name of the secret
	Name string `json:"name"`

	// Namespace is the namespace of the secret
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// Route defines a route configuration for a listener
type Route struct {
	// Name is a unique name for this route
	Name string `json:"name"`

	// Match defines criteria for matching requests to this route
	Match RouteMatch `json:"match"`

	// Route defines the action to take when a match occurs
	Route RouteAction `json:"route"`
}

// RouteMatch defines criteria for matching requests to a route
type RouteMatch struct {
	// Prefix path to match (e.g., /api/)
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// Exact path to match
	// +optional
	Path string `json:"path,omitempty"`

	// Regular expression to match path
	// +optional
	Regex string `json:"regex,omitempty"`
}

// RouteAction defines the action to take when a route is matched
type RouteAction struct {
	// Cluster is the target cluster for this route
	Cluster string `json:"cluster"`

	// Timeout for the route request
	// +optional
	Timeout string `json:"timeout,omitempty"`
}

// ListenerStatus defines the observed state of Listener
type ListenerStatus struct {
	// ObservedGeneration is the most recent generation observed for this Listener
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the Listener's state
	// +optional
	Conditions []ListenerCondition `json:"conditions,omitempty"`
}

// ListenerCondition represents a condition of the Listener resource
type ListenerCondition struct {
	// Type of the condition
	Type string `json:"type"`

	// Status of the condition, one of True, False, Unknown
	Status bool `json:"status"`

	// Reason for the condition's last transition
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message associated with the condition
	// +optional
	Message string `json:"message,omitempty"`

	// LastTransitionTime is the last time the condition transitioned from one status to another
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Listener is the Schema for the listeners API
type Listener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ListenerSpec   `json:"spec,omitempty"`
	Status ListenerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ListenerList contains a list of Listener
type ListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Listener `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Listener{}, &ListenerList{})
}
