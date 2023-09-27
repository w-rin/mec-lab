package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EDPComponentSpec defines the desired state of EDPComponent
// +k8s:openapi-gen=true
type EDPComponentSpec struct {
	Type string `json:"type"`
	Url  string `json:"url"`
	Icon string `json:"icon"`
}

// EDPComponentStatus defines the observed state of EDPComponent
// +k8s:openapi-gen=true
type EDPComponentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EDPComponent is the Schema for the edpcomponents API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=edpcomponents,scope=Namespaced
type EDPComponent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EDPComponentSpec   `json:"spec,omitempty"`
	Status EDPComponentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EDPComponentList contains a list of EDPComponent
type EDPComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EDPComponent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EDPComponent{}, &EDPComponentList{})
}
