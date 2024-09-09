package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/external-dns/endpoint"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.Cluster{}).
		Watches(
			&clusterv1.Machine{},
			handler.EnqueueRequestsFromMapFunc(r.MachineToEndpointMapFunc),
		).
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.ClusterToEndpointMapFunc),
		).
		Complete(r)
}

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconcile cluster")

	ep, err := r.GenerateDNSEndpoint(ctx, req)
	if err != nil {
		logger.Error(err, "Failed to generate endpoint")
	}
	err = r.createOrUpdateEndpoint(ctx, ep)
	if err != nil {
		return ctrl.Result{}, err
	}

	return res, err
}

func (r *ClusterReconciler) createOrUpdateEndpoint(ctx context.Context, ep *endpoint.DNSEndpoint) error {
	oldEp := &endpoint.DNSEndpoint{}
	epName := types.NamespacedName{
		Namespace: ep.Namespace,
		Name:      ep.Name,
	}
	err := r.Client.Get(ctx, epName, oldEp)
	if apierrors.IsNotFound(err) {
		fmt.Printf("%+v DNSEndpoint not found\n", epName)
		err = r.Client.Create(ctx, ep, &client.CreateOptions{})
		if err != nil {
			return err
		}
	}
	//TODO compare oldEp and ep
	return nil
}
func (r *ClusterReconciler) GenerateDNSEndpoint(ctx context.Context, req ctrl.Request) (*endpoint.DNSEndpoint, error) {
	ep := &endpoint.DNSEndpoint{}
	ep.ObjectMeta = metav1.ObjectMeta{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	cluster := &clusterv1.Cluster{}
	err := r.Client.Get(ctx, req.NamespacedName, cluster, &client.GetOptions{})
	if err != nil {
		return nil, err
	}
	endpoints := []*endpoint.Endpoint{}
	machines := &clusterv1.MachineList{}
	selectors := getSelectors(req.NamespacedName)
	err = r.List(ctx, machines, selectors...)
	if err != nil {
		return nil, err
	}
	ips := extractStatus(machines)
	record := &endpoint.Endpoint{
		DNSName:    cluster.Spec.ControlPlaneEndpoint.Host,
		RecordTTL:  endpoint.TTL(30),
		RecordType: "A",
		Targets:    ips,
	}
	endpoints = append(endpoints, record)
	ep.Spec = endpoint.DNSEndpointSpec{
		Endpoints: endpoints,
	}

	return ep, nil
}
func (r *ClusterReconciler) ClusterToEndpointMapFunc(ctx context.Context, o client.Object) []ctrl.Request {
	result := []ctrl.Request{}
	c, ok := o.(*clusterv1.Cluster)

	if !ok {
		panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
	}
	epName := c.Spec.ControlPlaneEndpoint.Host
	if epName == "" {
		fmt.Sprintf("ControlPlaneEndpoint.spec is not set")
		return nil
	}
	//get controller plane machines for cluster
	selectors := getSelectors(types.NamespacedName{
		Name:      c.Name,
		Namespace: c.Namespace,
	})

	machineList := &clusterv1.MachineList{}
	if err := r.Client.List(ctx, machineList, selectors...); err != nil {
		return nil
	}
	ips := extractStatus(machineList)
	fmt.Printf("IPs for cluster %s, %+v", c.Name, ips)
	return result
}

func getSelectors(n types.NamespacedName) []client.ListOption {
	x := []client.ListOption{
		client.InNamespace(n.Namespace),
		client.MatchingLabels{
			clusterv1.ClusterNameLabel:         n.Name,
			clusterv1.MachineControlPlaneLabel: "",
		},
	}
	return x

}
func extractStatus(l *clusterv1.MachineList) []string {
	targets := make([]string, 0)
	for _, m := range l.Items {
		if m.Status.Addresses != nil {
			a := getIPs(m.Status.Addresses)
			if a == "" {
				fmt.Printf("Machine %s/%s doesn't have IP yet", m.Namespace, m.Name)
				continue
			}
			targets = append(targets, a)
		}
	}
	return targets
}

func getIPs(al []clusterv1.MachineAddress) string {
	for _, a := range al {
		if a.Type == "ExternalIP" {
			return a.Address
		}
	}
	return ""
}
func (r ClusterReconciler) MachineToEndpointMapFunc(_ context.Context, o client.Object) []ctrl.Request {
	result := []ctrl.Request{}
	return result
}
