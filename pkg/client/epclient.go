package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	externaldns "sigs.k8s.io/external-dns/endpoint"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "externaldns.k8s.io", Version: "v1alpha1"}

	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = schemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &externaldns.DNSEndpoint{}, &externaldns.DNSEndpointList{})
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
