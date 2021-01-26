package service_router

import (
	"testing"
	"time"
)

var serverList = ServerList{
	Server{
		Host:        "192.168.1.1",
		Port:        8888,
		Protocol:    ServerProtocol_THRIFT,
		Weight:      10,
		ServiceName: "test",
		UpdateTime:  uint64(time.Now().UnixNano() / 1e6),
		Status:      ServerStatus_AVAILABLE,
	},
	Server{
		Host:        "192.168.1.2",
		Port:        9999,
		Protocol:    ServerProtocol_THRIFT,
		Weight:      1,
		ServiceName: "test",
		UpdateTime:  uint64(time.Now().UnixNano() / 1e6),
		Status:      ServerStatus_AVAILABLE,
	},
	Server{
		Host:        "192.168.1.3",
		Port:        6666,
		Protocol:    ServerProtocol_THRIFT,
		Weight:      8,
		ServiceName: "test",
		UpdateTime:  uint64(time.Now().UnixNano() / 1e6),
		Status:      ServerStatus_AVAILABLE,
	},
	Server{
		Host:        "192.168.10.1",
		Port:        8888,
		Protocol:    ServerProtocol_THRIFT,
		Weight:      10,
		ServiceName: "test",
		UpdateTime:  uint64(time.Now().UnixNano() / 1e6),
		Status:      ServerStatus_AVAILABLE,
	},
	Server{
		Host:        "192.168.10.2",
		Port:        8080,
		Protocol:    ServerProtocol_THRIFT,
		Weight:      8,
		ServiceName: "test",
		UpdateTime:  uint64(time.Now().UnixNano() / 1e6),
		Status:      ServerStatus_AVAILABLE,
	},
}

func BenchmarkLoadBalanceRandom(b *testing.B) {
	lb := NewLoadBalanceRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.Select(serverList)
	}
}

func BenchmarkLoadBalanceRoundrobin(b *testing.B) {
	lb := NewLoadBalanceRoundrobin()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.Select(serverList)
	}
}

func BenchmarkLoadBalanceLocalFirst(b *testing.B) {
	lb := NewLoadBalanceLocalFirst(BalanceLocalFirstConfig{
		LocalIp:   "192.168.1.1",
		DiffRange: 256,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.Select(serverList)
	}
}

func BenchmarkLoadBalanceIpRangeFirst(b *testing.B) {
	lb := NewLoadBalanceIpRangeFirst(BalanceLocalFirstConfig{
		LocalIp:   "192.168.1.1",
		DiffRange: 256,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.Select(serverList)
	}
}

func BenchmarkLoadBalanceConfigurableWeight(b *testing.B) {
	lb := NewLoadBalanceConfigurableWeight()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.Select(serverList)
	}
}

func BenchmarkLoadBalanceStaticWeight(b *testing.B) {
	lb := NewLoadBalanceStaticWeight()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.Select(serverList)
	}
}
