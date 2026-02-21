package consul

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

const ttl = time.Second * 5

type Service struct {
	client *api.Client
}

func NewRegistry(addr, serviceName string) (*Service, error) {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Service{client: client}, nil
}

func (s *Service) Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return err
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	return s.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      instanceID,
		Name:    serviceName,
		Address: host,
		Port:    portNum,
		Check: &api.AgentServiceCheck{
			CheckID:                        instanceID,
			TLSSkipVerify:                  true,
			TTL:                            ttl.String(),
			DeregisterCriticalServiceAfter: ttl.String(),
		},
	})
}

func (s *Service) Deregister(ctx context.Context, instanceID string, serviceName string) error {
	log.Printf("deregistering service %s", instanceID)
	return s.client.Agent().CheckDeregister(instanceID)
}

func (s *Service) HealthCheck(instanceID string, serviceName string) error {
	return s.client.Agent().UpdateTTL(instanceID, "online", api.HealthPassing)
}

func (s *Service) Discover(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := s.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var instances []string
	for _, entry := range entries {
		instances = append(instances, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}

	return instances, nil
}
