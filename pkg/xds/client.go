package xds

import (
	"context"
	"fmt"
)

// Resource represents a generic xDS resource
type Resource struct {
	// Type is the xDS resource type (listener, cluster, route, etc.)
	Type string

	// Name is the name of the resource
	Name string

	// Data contains the resource configuration
	Data map[string]interface{}
}

// Client defines the interface for interacting with an xDS server
type Client interface {
	// UpdateResource updates or creates a resource on the xDS server
	UpdateResource(ctx context.Context, resource *Resource) error

	// DeleteResource deletes a resource from the xDS server
	DeleteResource(ctx context.Context, resourceType, resourceName string) error
}

// ClientOptions contains options for the xDS client
type ClientOptions struct {
	NodeID string
}

// ClientOption defines an option for the xDS client
type ClientOption func(*ClientOptions)

// WithNodeID sets the node ID for the xDS client
func WithNodeID(nodeID string) ClientOption {
	return func(o *ClientOptions) {
		o.NodeID = nodeID
	}
}

// NewClient creates a new xDS client
func NewClient(serverAddress string, opts ...ClientOption) Client {
	options := &ClientOptions{
		NodeID: "xds-operator",
	}

	for _, opt := range opts {
		opt(options)
	}

	return &defaultClient{
		serverAddress: serverAddress,
		nodeID:        options.NodeID,
		resources:     make(map[string]*Resource),
	}
}

// defaultClient is a simple implementation of the Client interface
type defaultClient struct {
	serverAddress string
	nodeID        string
	resources     map[string]*Resource
}

// UpdateResource updates or creates a resource on the xDS server
func (c *defaultClient) UpdateResource(ctx context.Context, resource *Resource) error {
	// For now, just store the resource in memory
	// In a real implementation, this would send the resource to the xDS server
	key := fmt.Sprintf("%s/%s", resource.Type, resource.Name)
	c.resources[key] = resource

	fmt.Printf("Updated xDS resource: %s/%s\n", resource.Type, resource.Name)
	return nil
}

// DeleteResource deletes a resource from the xDS server
func (c *defaultClient) DeleteResource(ctx context.Context, resourceType, resourceName string) error {
	// For now, just remove the resource from memory
	// In a real implementation, this would send a deletion request to the xDS server
	key := fmt.Sprintf("%s/%s", resourceType, resourceName)
	delete(c.resources, key)

	fmt.Printf("Deleted xDS resource: %s/%s\n", resourceType, resourceName)
	return nil
}
