//go:build unit

package test

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/Kuadrant/multicluster-gateway-controller/pkg/apis/v1alpha1"
	"github.com/Kuadrant/multicluster-gateway-controller/pkg/dns"
	"github.com/Kuadrant/multicluster-gateway-controller/pkg/traffic"
	testutil "github.com/Kuadrant/multicluster-gateway-controller/test/util"
)

type FakeCertificateService struct {
	controlClient client.Client
}

func (s *FakeCertificateService) CleanupCertificates(_ context.Context, _ traffic.Interface) error {
	return nil
}

func (s *FakeCertificateService) EnsureCertificate(_ context.Context, host string, _ v1.Object) error {
	if host == testutil.FailEnsureCertHost {
		return fmt.Errorf(testutil.FailEnsureCertHost)
	}
	return nil
}

func (s *FakeCertificateService) GetCertificateSecret(ctx context.Context, host string, namespace string) (*corev1.Secret, error) {
	if host == testutil.FailGetCertSecretHost {
		return &corev1.Secret{}, fmt.Errorf(testutil.FailGetCertSecretHost)
	}
	tlsSecret := &corev1.Secret{ObjectMeta: v1.ObjectMeta{
		Name:      host,
		Namespace: namespace,
	}}
	if err := s.controlClient.Get(ctx, client.ObjectKeyFromObject(tlsSecret), tlsSecret); err != nil {
		return nil, err
	}
	return tlsSecret, nil
}

func NewTestCertificateService(client client.Client) *FakeCertificateService {
	return &FakeCertificateService{controlClient: client}
}

type FakeHostService struct {
	controlClient client.Client
}

func (h *FakeHostService) GetDNSRecordsFor(_ context.Context, _ traffic.Interface) ([]*v1alpha1.DNSRecord, error) {
	return nil, nil
}

func (h *FakeHostService) CleanupDNSRecords(_ context.Context, _ traffic.Interface) error {
	return nil
}

func (h *FakeHostService) CreateDNSRecord(_ context.Context, subDomain string, _ *v1alpha1.ManagedZone, _ v1.Object) (*v1alpha1.DNSRecord, error) {
	if subDomain == testutil.FailCreateDNSSubdomain {
		return nil, fmt.Errorf(testutil.FailCreateDNSSubdomain)
	}
	record := v1alpha1.DNSRecord{}
	return &record, nil
}

func (h *FakeHostService) GetDNSRecord(ctx context.Context, subDomain string, managedZone *v1alpha1.ManagedZone, _ v1.Object) (*v1alpha1.DNSRecord, error) {
	if subDomain == testutil.FailFetchDANSSubdomain {
		return &v1alpha1.DNSRecord{}, fmt.Errorf(testutil.FailFetchDANSSubdomain)
	}
	record := &v1alpha1.DNSRecord{
		ObjectMeta: v1.ObjectMeta{
			Name:      managedZone.Spec.DomainName,
			Namespace: managedZone.GetNamespace(),
		},
	}

	if err := h.controlClient.Get(ctx, client.ObjectKeyFromObject(record), record); err != nil {
		return nil, err
	}
	return record, nil
}

func (h *FakeHostService) AddEndpoints(_ context.Context, gateway traffic.Interface, _ *v1alpha1.DNSRecord) error {
	hosts := gateway.GetHosts()
	for _, host := range hosts {
		if host == testutil.FailEndpointsHostname {
			return fmt.Errorf(testutil.FailEndpointsHostname)
		}
	}
	return nil
}

func NewTestHostService(client client.Client) *FakeHostService {
	return &FakeHostService{controlClient: client}
}

type FakeGatewayPlacer struct{}

func (p *FakeGatewayPlacer) GetClusterGateway(_ context.Context, _ *v1beta1.Gateway, _ string) (dns.ClusterGateway, error) {
	return dns.ClusterGateway{}, nil
}

func (p *FakeGatewayPlacer) Place(_ context.Context, upstream *v1beta1.Gateway, _ *v1beta1.Gateway, _ ...v1.Object) (sets.Set[string], error) {
	if upstream.Labels == nil {
		return nil, nil
	}
	if *upstream.Spec.Listeners[0].Hostname == testutil.FailPlacementHostname {
		return nil, fmt.Errorf(testutil.FailPlacementHostname)
	}
	targetClusters := sets.Set[string](sets.NewString())
	targetClusters.Insert(testutil.Cluster)
	return targetClusters, nil
}

func (p *FakeGatewayPlacer) GetPlacedClusters(_ context.Context, gateway *v1beta1.Gateway) (sets.Set[string], error) {
	if gateway.Labels == nil {
		return nil, nil
	}
	placedClusters := sets.Set[string](sets.NewString())
	placedClusters.Insert(testutil.Cluster)
	return placedClusters, nil
}

func (p *FakeGatewayPlacer) GetClusters(_ context.Context, _ *v1beta1.Gateway) (sets.Set[string], error) {
	return nil, nil
}

func (p *FakeGatewayPlacer) ListenerTotalAttachedRoutes(_ context.Context, _ *v1beta1.Gateway, listenerName string, _ string) (int, error) {
	if listenerName == testutil.ValidTestHostname {
		return 1, nil
	}
	return 0, nil
}

func (p *FakeGatewayPlacer) GetAddresses(_ context.Context, _ *v1beta1.Gateway, _ string) ([]v1beta1.GatewayAddress, error) {
	t := v1beta1.IPAddressType
	return []v1beta1.GatewayAddress{
		{
			Type:  &t,
			Value: "1.1.1.1",
		},
	}, nil
}

func NewTestGatewayPlacer() *FakeGatewayPlacer {
	return &FakeGatewayPlacer{}
}
