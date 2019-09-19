package consul

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

// Client provides an interface for getting data out of Consul
type Client interface {
	// Get a Service from consul
	Service(string, string) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
	// Register a service with local agent
	Register(string, string, int) error
	// Deregister a service with local agent
	DeRegister(string) error
	// Catalog ...
	Catalog() *consul.Catalog
}

type client struct {
	consul *consul.Client
}

// NewConsulClient returns a Client interface for given consul address
func NewConsulClient(addr string) (Client, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &client{consul: c}, nil
}

func (c *client) Catalog() *consul.Catalog {
	return c.consul.Catalog()
}

// Register a service with consul local agent
func (c *client) Register(name string, address string, port int) error {
	reg := &consul.AgentServiceRegistration{
		Address: address,
		ID:      name,
		Name:    name,
		Port:    port,
	}
	return c.consul.Agent().ServiceRegister(reg)
}

// DeRegister a service with consul local agent
func (c *client) DeRegister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service return a service
func (c *client) Service(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.consul.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, errors.Wrapf(err, "service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "error get service info from consul")
	}
	return addrs, meta, nil
}
