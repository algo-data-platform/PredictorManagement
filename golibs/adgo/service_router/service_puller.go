package service_router

import (
	"context"
	"log"
	"sync"
	"time"
)

// service pull interface
type ServicePull interface {
	Pull(serviceName string)
}

type ConsulServiceInfoPull struct {
	registry Registry
}

var (
	serviceInfoPull *ConsulServiceInfoPull
	sipOnce         sync.Once
)

// singleton
func GetConsulServiceInfoPull(registry Registry) *ConsulServiceInfoPull {
	sipOnce.Do(func() {
		serviceInfoPull = &ConsulServiceInfoPull{
			registry: registry,
		}
	})

	return serviceInfoPull
}

func (sip *ConsulServiceInfoPull) Pull(serviceName string) {
	sip.registry.Discover(serviceName)
}

type ConsulServiceConfigPull struct {
	registry Registry
}

var (
	serviceConfigPull *ConsulServiceConfigPull
	scpOnce           sync.Once
)

func GetConsulServiceConfigPull(registry Registry) *ConsulServiceConfigPull {
	scpOnce.Do(func() {
		serviceConfigPull = &ConsulServiceConfigPull{
			registry: registry,
		}
	})

	return serviceConfigPull
}

func (scp *ConsulServiceConfigPull) Pull(serviceName string) {
	scp.registry.GetConfig(serviceName)
}

const (
	ServicePullerType_Config = "CONFIG_PULLER"
	ServicePullerType_Server = "SERVER_PULLER"
)

type ServicePuller struct {
	serviceName string
	pullerType  string
	interval    time.Duration
	servicePull ServicePull
	ticker      *time.Ticker
	cancel      context.CancelFunc
}

func calculateInterval(interval uint32) time.Duration {
	return time.Duration(2*interval*1000/3) * time.Millisecond
}

func NewServicePuller(serviceName string, servicePull ServicePull, routerConfig ServiceRouterConfig, pullerType string) *ServicePuller {
	puller := &ServicePuller{
		interval:    calculateInterval(routerConfig.PullInterval),
		pullerType:  pullerType,
		serviceName: serviceName,
		servicePull: servicePull,
	}
	puller.start()
	return puller
}

func (sp *ServicePuller) ConfigNotify(serviceName string, serviceConfig ServiceConfig) {
	if serviceName != sp.serviceName {
		return
	}
	interval := calculateInterval(serviceConfig.Router.PullInterval)
	if interval > 0 && interval != sp.interval {
		sp.Restart(interval)
	}
}

func (sp *ServicePuller) start() context.CancelFunc {
	// first do task, then create ticker
	sp.servicePull.Pull(sp.serviceName)
	ctx, cancel := context.WithCancel(context.Background())
	sp.cancel = cancel
	ticker := time.NewTicker(sp.interval)
	sp.ticker = ticker
	log.Printf("[%s] - %s ticker created with interval %v\n", sp.pullerType, sp.serviceName, sp.interval)
	go func(serviceName string, servicePull ServicePull, ctx context.Context) {
		for {
			select {
			case <-ticker.C:
				servicePull.Pull(serviceName)
				//log.Printf("PULL: %s\n", serviceName)
			case <-ctx.Done():
				ticker.Stop()
				log.Printf("[%s] - %s ticker stopped\n", sp.pullerType, sp.serviceName)
				return
			}
		}
	}(sp.serviceName, sp.servicePull, ctx)
	return cancel
}

func (sp *ServicePuller) Restart(interval time.Duration) {
	sp.interval = interval
	sp.ticker.Stop()
	sp.start()
}

func (sp *ServicePuller) Stop() {
	sp.cancel()
}
