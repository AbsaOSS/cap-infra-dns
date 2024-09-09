package controller

import (
	"sigs.k8s.io/external-dns/endpoint"
)

func CopyWithoutLastAppliedConfiguration(annotations map[string]string) map[string]string {
	const key = "kubectl.kubernetes.io/last-applied-configuration"
	newAnnotations := make(map[string]string)
	for k, v := range annotations {
		if k != key {
			newAnnotations[k] = v
		}
	}
	return newAnnotations
}

// EqualMaps compares two maps. It ignores nil and empty map difference
func EqualMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if b[k] != a[k] {
			return false
		}
	}
	return true
}

// EqualDNSEndpoints compares labels, annotations and spec
func EqualDNSEndpoints(newEp, oldEp *endpoint.DNSEndpoint) bool {
	if newEp == nil && oldEp == nil {
		return true
	}
	if newEp == nil || oldEp == nil {
		return false
	}

	if !EqualMaps(newEp.Annotations, oldEp.Annotations) {
		return false
	}

	if !EqualMaps(newEp.Labels, oldEp.Labels) {
		return false
	}

	if len(newEp.Spec.Endpoints) != 1 {
		return false
	}

	if len(oldEp.Spec.Endpoints) != 1 {
		return false
	}

	newE := newEp.Spec.Endpoints[0]
	oldE := oldEp.Spec.Endpoints[0]

	return newE.DNSName == oldE.DNSName &&
		newE.RecordType == oldE.RecordType &&
		newE.RecordTTL == oldE.RecordTTL &&
		newE.Targets.Same(oldE.Targets)
}
