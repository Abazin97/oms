package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Service interface {
	Register(ctx context.Context, instanceID string, serverName string, hostPort string) error
	Deregister(ctx context.Context, instanceID string, serviceName string) error
	Discover(ctx context.Context, serviceName string) ([]string, error)
	HealthCheck(instanceID string, serviceName string) error
}

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
