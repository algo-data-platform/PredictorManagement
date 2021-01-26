package service_router

import (
	"context"
	"log"
	"sync"
	"time"
)

type ServicePush interface {
	Push(server Server) error
}

type ConsulServicePush struct {
	registry Registry
}

var (
	consulServicePusherOnce     sync.Once
	consulServicePusherInstance *ConsulServicePush
)

func NewConsulServicePush(registry Registry) *ConsulServicePush {
	consulServicePusherOnce.Do(func() {
		consulServicePusherInstance = &ConsulServicePush{
			registry: registry,
		}
	})
	return consulServicePusherInstance
}

func (csp *ConsulServicePush) Push(server Server) error {
	return csp.registry.RegisterServer(server)
}

type ServicePusher struct {
	server Server
	push   ServicePush
	ticker *time.Ticker
	ttl    time.Duration
	cancel context.CancelFunc
}

func calculateTtl(ttl uint32) time.Duration {
	return time.Duration(2*ttl/3) * time.Millisecond
}

func NewServicePusher(server Server, push ServicePush, routerConfig ServiceRouterConfig) *ServicePusher {
	sh := &ServicePusher{
		server: server,
		push:   push,
		ttl:    calculateTtl(routerConfig.TtlInMs),
	}
	sh.start()
	return sh
}

func (sh *ServicePusher) ConfigNotify(serviceName string, serviceConfig ServiceConfig) {
	if serviceName != sh.server.ServiceName {
		return
	}
	ttl := calculateTtl(serviceConfig.Router.TtlInMs)
	if ttl > 0 && ttl != sh.ttl {
		sh.Restart(ttl)
	}
}

func (sh *ServicePusher) start() context.CancelFunc {
	// first do task, then create ticker
	sh.push.Push(sh.server)
	ctx, cancel := context.WithCancel(context.Background())
	sh.cancel = cancel
	ticker := time.NewTicker(sh.ttl)
	sh.ticker = ticker
	log.Printf("[Heartbeat] - %s ticker created with ttl %v\n", sh.server.ServiceName, sh.ttl)
	go func(ticker *time.Ticker, push ServicePush, server Server) {
		for {
			select {
			case <-ticker.C:
				push.Push(server)
				//log.Printf("HEARTBEAT: %s\n", server.ServiceName)
			case <-ctx.Done():
				ticker.Stop()
				log.Printf("[Heartbeat] - %s ticker stopped\n", sh.server.ServiceName)
				return
			}
		}
	}(ticker, sh.push, sh.server)
	return cancel
}

func (sh *ServicePusher) Restart(ttl time.Duration) {
	sh.ttl = ttl
	sh.ticker.Stop()
	sh.start()
}

func (sh *ServicePusher) Stop() {
	sh.cancel()
}

func (sh *ServicePusher) SetWeight(weight uint32) {
	sh.server.Weight = weight
}

func (sh *ServicePusher) SetShardList(shardList []uint32) {
	sh.server.ShardList = shardList
}

func (sh *ServicePusher) SetAvailableShardList(shardList []uint32) {
	sh.server.AvailableShardList = shardList
}

func (sh *ServicePusher) SetFollowerShardList(shardList []uint32) {
	sh.server.FollowerShardList = shardList
}

func (sh *ServicePusher) SetFollowerAvailableShardList(shardList []uint32) {
	sh.server.FollowerAvailableShardList = shardList
}

func (sh *ServicePusher) SetStatus(status string) {
	sh.server.Status = status
}
