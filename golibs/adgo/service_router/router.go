package service_router

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/algo-data-platform/predictor/golibs/adgo/common/cityhash"
	"github.com/algo-data-platform/predictor/golibs/adgo/common/tally"
	promreporter "github.com/algo-data-platform/predictor/golibs/adgo/common/tally/prometheus"
	"github.com/hashicorp/consul/api"
)

const (
	LoadBalanceMethod_RANDOM              = "random"
	LoadBalanceMethod_ROUNDROBIN          = "roundrobin"
	LoadBalanceMethod_LOCALFIRST          = "localfirst"
	LoadBalanceMethod_CONSISTENT          = "consistent"
	LoadBalanceMethod_CONFIGURABLE_WEIGHT = "configurable_weight"
	LoadBalanceMethod_ACTIVE_WEIGHT       = "active_weight"
	LoadBalanceMethod_USER_DEFINED        = "user_defined"
	LoadBalanceMethod_IPRANGEFIRST        = "iprangefirst"
	LoadBalanceMethod_STATIC_WEIGHT       = "static_weight"
)

const (
	ShardType_ALL      = "0"
	ShardType_LEADER   = "1"
	ShardType_FOLLOWER = "2"
)

type BalanceLocalFirstConfig struct {
	LocalIp   string
	DiffRange int
}

type ServerAddress struct {
	Host string
	Port uint16
}

type ClientOption struct {
	ServiceName         string
	ShardId             int64
	ShardType           string
	Protocol            string
	Loadbalance         string
	TargetServerAddress ServerAddress
	UserBalance         LoadBalance
	LocalFirstConfig    BalanceLocalFirstConfig
}

type ServiceRouterConfig struct {
	TtlInMs           uint32 `json:"TtlInMs"`
	LoadBalanceMethod string `json:"LoadBalance"`
	TotalShards       uint32 `json:"TotalShards"`
	PullInterval      uint32 `json:"PullInterval"`
}

type ServiceConfig struct {
	Router  ServiceRouterConfig
	Configs map[string]string
}

type Multiplexer func(string, http.Handler)

type RouterConfig struct {
	Consul      api.Config
	ProjectName string
	LocalIp     string
	HttpPort    int
	Multiplexer Multiplexer
}

// Get default ServiceRouterConfig
func DefaultServiceRouterConfig() ServiceRouterConfig {
	return ServiceRouterConfig{
		TtlInMs:           300000,
		LoadBalanceMethod: LoadBalanceMethod_RANDOM,
		TotalShards:       0,
		PullInterval:      3,
	}
}

func ClientOptionFactory(serviceName string, protocol string) ClientOption {
	config := DefaultClientOption()
	config.ServiceName = serviceName
	config.Protocol = protocol
	return config
}

// Get Default ClientOption
func DefaultClientOption() ClientOption {
	return ClientOption{
		ShardId:          4294967295,
		ShardType:        ShardType_LEADER,
		Loadbalance:      LoadBalanceMethod_LOCALFIRST,
		LocalFirstConfig: DefaultBalanceLocalFirstConfig(),
	}
}

// Get default BalanceLocalFirstConfig
func DefaultBalanceLocalFirstConfig() BalanceLocalFirstConfig {
	return BalanceLocalFirstConfig{
		DiffRange: 256,
	}
}

type Router struct {
	registry        Registry
	routerDb        RouterDb
	metrics         tally.Scope
	counters        map[string]tally.Counter
	configPullers   map[string]*ServicePuller
	discoverPullers map[string]*ServicePuller
	servicePushers  map[string]*ServicePusher
	balancers       map[uint64]LoadBalance
	counterLock     sync.RWMutex
	configLock      sync.RWMutex
	discoverLock    sync.RWMutex
	pusherLock      sync.RWMutex
	balancerLock    sync.RWMutex
}

var (
	routerInstance *Router
	routerOnce     sync.Once
)

func (router *Router) getServiceKey(server Server) string {
	return server.ServiceName + server.Host + strconv.Itoa(int(server.Port))
}

func (router *Router) GetOrCreateServicePusher(server Server) *ServicePusher {
	key := router.getServiceKey(server)
	router.pusherLock.RLock()
	sh, ok := router.servicePushers[key]
	router.pusherLock.RUnlock()
	if !ok {
		// upgrade
		router.pusherLock.Lock()
		defer router.pusherLock.Unlock()
		sh, ok = router.servicePushers[key]
		if !ok {
			pusher := NewConsulServicePush(router.registry)
			routerConfig, _ := router.GetServiceRouterConfig(server.ServiceName)
			sh = NewServicePusher(server, pusher, *routerConfig)
			router.registry.SubscribeConfig(server.ServiceName, sh)
			router.servicePushers[key] = sh
		}
	}
	return sh
}

func (router *Router) GetOrCreateConfigPuller(serviceName string) *ServicePuller {
	router.configLock.RLock()
	puller, ok := router.configPullers[serviceName]
	router.configLock.RUnlock()
	if !ok {
		// upgrade
		router.configLock.Lock()
		defer router.configLock.Unlock()
		puller, ok = router.configPullers[serviceName]
		if !ok {
			// subscribe
			router.registry.SubscribeConfig(serviceName, router.routerDb)
			serviceConfigPull := GetConsulServiceConfigPull(router.registry)
			serviceConfigPull.Pull(serviceName)
			routerConfig := router.routerDb.GetServiceRouterConfig(serviceName)
			puller = NewServicePuller(serviceName, serviceConfigPull, routerConfig, ServicePullerType_Config)
			// subscribe
			router.registry.SubscribeConfig(serviceName, puller)
			router.configPullers[serviceName] = puller
		}
	}
	return puller
}

func (router *Router) GetOrCreateDiscoverPuller(serviceName string) *ServicePuller {
	router.discoverLock.RLock()
	puller, ok := router.discoverPullers[serviceName]
	router.discoverLock.RUnlock()
	if !ok {
		// upgrade
		router.discoverLock.Lock()
		defer router.discoverLock.Unlock()
		puller, ok = router.discoverPullers[serviceName]
		if !ok {
			// subscribe
			router.registry.SubscribeService(serviceName, router.routerDb)
			serviceInfoPull := GetConsulServiceInfoPull(router.registry)
			routerConfig, _ := router.GetServiceRouterConfig(serviceName)
			puller = NewServicePuller(serviceName, serviceInfoPull, *routerConfig, ServicePullerType_Server)
			router.discoverPullers[serviceName] = puller
		}
	}
	return puller
}

func (router *Router) GetOrCreateBalancer(option ClientOption) LoadBalance {
	var (
		balanceId   uint64
		method      string      = option.Loadbalance
		protocol    string      = option.Protocol
		serviceName string      = option.ServiceName
		userBalance LoadBalance = option.UserBalance
	)
	balanceId = cityhash.CityHash64WithSeed(method, uint64(len(method)), 0)
	balanceId = cityhash.CityHash64WithSeed(serviceName, uint64(len(serviceName)), balanceId)
	balanceId = cityhash.CityHash64WithSeed(protocol, uint64(len(protocol)), balanceId)
	if method == LoadBalanceMethod_USER_DEFINED {
		if nil != userBalance {
			customBalancerName := reflect.TypeOf(userBalance).Name()
			balanceId = cityhash.CityHash64WithSeed(customBalancerName, uint64(len(customBalancerName)), balanceId)
		} else {
			method = LoadBalanceMethod_RANDOM
		}
	}
	router.balancerLock.RLock()
	balancer, ok := router.balancers[balanceId]
	router.balancerLock.RUnlock()
	if !ok {
		router.balancerLock.Lock()
		defer router.balancerLock.Unlock()
		balancer, ok = router.balancers[balanceId]
		if !ok {
			switch method {
			case LoadBalanceMethod_RANDOM:
				balancer = NewLoadBalanceRandom()
			case LoadBalanceMethod_ROUNDROBIN:
				balancer = NewLoadBalanceRoundrobin()
			case LoadBalanceMethod_LOCALFIRST:
				balancer = NewLoadBalanceLocalFirst(option.LocalFirstConfig)
			case LoadBalanceMethod_IPRANGEFIRST:
				balancer = NewLoadBalanceIpRangeFirst(option.LocalFirstConfig)
			case LoadBalanceMethod_STATIC_WEIGHT:
				balancer = NewLoadBalanceStaticWeight()
			case LoadBalanceMethod_CONFIGURABLE_WEIGHT:
				balancer = NewLoadBalanceConfigurableWeight()
			case LoadBalanceMethod_USER_DEFINED:
				balancer = userBalance
			default:
				balancer = NewLoadBalanceRandom()
			}
			router.balancers[balanceId] = balancer
		}
	}
	return balancer
}

func (router *Router) Discover(option ClientOption) (*Server, bool) {
	router.GetOrCreateDiscoverPuller(option.ServiceName)
	// routerDb 只返回副本，所以loadbalancer可以返回引用
	serverList, ok := router.routerDb.SelectServers(option.ServiceName, option.Protocol, option.ShardId, option.ShardType)
	if !ok {
		counter := router.getCounter(ROUTER_METRICS_SELECT_ADDRESS_TAGS_ADDR_VAL_NONE)
		counter.Inc(1)
		return nil, false
	}
	balancer := router.GetOrCreateBalancer(option)
	res, ok := balancer.Select(serverList)
	counter := router.getCounter(res.Host + ":" + strconv.Itoa(int(res.Port)))
	counter.Inc(1)
	return res, ok
}

func (router *Router) GetServerList(option ClientOption) (ServerList, bool) {
	router.GetOrCreateDiscoverPuller(option.ServiceName)
	serverList, ok := router.routerDb.SelectServers(option.ServiceName, option.Protocol, option.ShardId, option.ShardType)
	if !ok {
		return nil, false
	}
	return serverList, ok
}

func (router *Router) GetServiceRouterConfig(serviceName string) (*ServiceRouterConfig, bool) {
	router.GetOrCreateConfigPuller(serviceName)
	config := router.routerDb.GetServiceRouterConfig(serviceName)
	return &config, true
}

func (router *Router) GetConfigs(name string) (map[string]string, bool) {
	router.GetOrCreateConfigPuller(name)
	return router.routerDb.GetServiceConfig(name)
}

func (router *Router) SetRouterDb(db RouterDb) {
	router.routerDb = db
}

func (router *Router) GetRouterDb() RouterDb {
	return router.routerDb
}

func (router *Router) RegisterServer(server Server) {
	router.GetOrCreateServicePusher(server)
}

func (router *Router) GetRegistry() Registry {
	return router.registry
}

func (router *Router) getCounter(key string) tally.Counter {
	router.counterLock.RLock()
	counter, ok := router.counters[key]
	router.counterLock.RUnlock()
	if !ok {
		router.counterLock.Lock()
		counter, ok = router.counters[key]
		if !ok {
			tags := map[string]string{ROUTER_METRICS_SELECT_ADDRESS_TAGS_ADDR: key}
			counter = router.metrics.Tagged(tags).Counter(ROUTER_METRICS_SELECT_ADDRESS)
			router.counters[key] = counter
		}
		router.counterLock.Unlock()
	}
	return counter
}

func (router *Router) GetMetrics() tally.Scope {
	return router.metrics
}

func (router *Router) initVersionMetrics(scope tally.Scope) {
	scope.Tagged(map[string]string{"type": ROUTER_MAJOR_VERSION}).Gauge(ROUTER_VERSION).Update(MAJOR_VERSION)
	scope.Tagged(map[string]string{"type": ROUTER_MINOR_VERSION}).Gauge(ROUTER_VERSION).Update(MINOR_VERSION)
	scope.Tagged(map[string]string{"type": ROUTER_REVISION}).Gauge(ROUTER_VERSION).Update(REVISION)
}

func (router *Router) initMetricsReporter(config *RouterConfig, reporter promreporter.Reporter) {
	var port = config.HttpPort
	if port < 0 {
		port = 0
	}
	if nil == config.Multiplexer {
		serveMux := &http.ServeMux{}
		listen, _ := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if port == 0 {
			port = listen.Addr().(*net.TCPAddr).Port
		}
		serveMux.Handle(METRICS_URI, reporter.JsonHTTPHandler())
		serveMux.Handle(PROMETHEUS_URI, reporter.HTTPHandler())
		http := &http.Server{Handler: serveMux}
		go http.Serve(listen)
	} else {
		config.Multiplexer(METRICS_URI, reporter.JsonHTTPHandler())
		config.Multiplexer(PROMETHEUS_URI, reporter.HTTPHandler())
	}
	// registry http server
	server := Server{
		ServiceName: config.ProjectName,
		Host:        config.LocalIp,
		Port:        uint16(port),
		Protocol:    ServerProtocol_HTTP,
		Status:      ServerStatus_AVAILABLE,
	}
	router.RegisterServer(server)
	if _, ok := router.registry.GetValue(GetRouterConfigPath(&server)); !ok {
		// registry router config
		router.registry.PutRouterConfig(server, DefaultServiceRouterConfig())
	}
}

// singleton
func GetRouter(config *RouterConfig) *Router {
	routerOnce.Do(func() {
		registry, _ := NewConsulRegistry(config)
		routerDb := GetRouterDb()
		reporter := promreporter.NewReporter(promreporter.Options{})
		scope, _ := tally.NewRootScope(tally.ScopeOptions{
			Prefix:         PREFIX,
			CachedReporter: reporter,
			Separator:      promreporter.DefaultSeparator,
			Tags:           map[string]string{"project": config.ProjectName},
		}, time.Second)
		routerInstance = &Router{
			routerDb:        routerDb,
			registry:        registry,
			metrics:         scope,
			counters:        make(map[string]tally.Counter),
			discoverPullers: make(map[string]*ServicePuller),
			configPullers:   make(map[string]*ServicePuller),
			servicePushers:  make(map[string]*ServicePusher),
			balancers:       make(map[uint64]LoadBalance),
		}
		routerInstance.initVersionMetrics(scope)
		routerInstance.initMetricsReporter(config, reporter)
	})
	return routerInstance
}
