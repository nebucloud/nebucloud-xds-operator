package pkg

import (
	"context"

	"github.com/nebucloud/nebucloud-xds-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client provides a public interface to interact with xDS resources
type Client struct {
	kubeClient client.Client
}

// NewClient creates a new client for managing xDS resources
func NewClient(kubeClient client.Client) *Client {
	return &Client{
		kubeClient: kubeClient,
	}
}

// CreateListener creates a new Listener resource
func (c *Client) CreateListener(ctx context.Context, listener *v1alpha1.Listener) error {
	return c.kubeClient.Create(ctx, listener)
}

// UpdateListener updates an existing Listener resource
func (c *Client) UpdateListener(ctx context.Context, listener *v1alpha1.Listener) error {
	return c.kubeClient.Update(ctx, listener)
}

// DeleteListener deletes a Listener resource
func (c *Client) DeleteListener(ctx context.Context, name, namespace string) error {
	listener := &v1alpha1.Listener{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, listener)
	if err != nil {
		return err
	}
	return c.kubeClient.Delete(ctx, listener)
}

// GetListener gets a Listener resource
func (c *Client) GetListener(ctx context.Context, name, namespace string) (*v1alpha1.Listener, error) {
	listener := &v1alpha1.Listener{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, listener)
	if err != nil {
		return nil, err
	}
	return listener, nil
}
