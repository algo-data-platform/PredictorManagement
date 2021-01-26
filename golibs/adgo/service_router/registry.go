package service_router

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

var (
	ErrInvalidServer = errors.New("invalid server")
)

// 服务协议
const (
	ServerProtocol_HTTP   = "http"
	ServerProtocol_THRIFT = "thrift"
	ServerProtocol_REDIS  = "redis"
)

// 服务状态
const (
	ServerStatus_AVAILABLE   = "available"
	ServerStatus_UNAVAILABLE = "unavailable"
	ServerStatus_UNKNOWN     = "unknown"
)

type ServerList []Server

// 服务
type Server struct {
	Host                       string            `json:"Host"`
	Port                       uint16            `json:"port"`
	ServiceName                string            `json:"ServiceName"`
	Protocol                   string            `json:"Protocol"`
	Status                     string            `json:"Status"`
	UpdateTime                 uint64            `json:"UpdateTime"`
	Weight                     uint32            `json:"Weight"`
	ShardList                  []uint32          `json:"ShardList"`
	AvailableShardList         []uint32          `json:"AvailableShardList"`
	FollowerShardList          []uint32          `json:"FollowerShardList"`
	FollowerAvailableShardList []uint32          `json:"FollowerAvailableShardList"`
	OtherSettings              map[string]string `json:"OtherSettings"`
}

// ServiceSubscriber interface
type ServiceSubscriber interface {
	ServiceNotify(serviceName string, services ServerList)
}

// ConfigSubscriber interface
type ConfigSubscriber interface {
	ConfigNotify(serviceName string, serviceConfig ServiceConfig)
}

type Registry interface {
	SubscribeService(serviceName string, subscriber ServiceSubscriber)
	UnsubscribeService(serviceName string)
	SubscribeConfig(serviceName string, subscriber ConfigSubscriber)
	UnsubscribeConfig(serviceName string)
	Discover(serviceName string)
	GetConfig(serviceName string)
	GetValue(key string) (string, bool)
	PutRouterConfig(server Server, config ServiceRouterConfig) error
	PutServiceConfigs(server Server, config map[string]string) error
	RegisterServer(server Server) error
	UnregisterServer(server Server) error
	UnregisterConfig(server Server) error
	Available(server Server) error
	Unavailable(server Server) error
}

// ConsulRegistry class
type ConsulRegistry struct {
	client                 *api.Client
	lastIndexes            sync.Map
	serviceSubscribers     map[string][]ServiceSubscriber
	configSubscribers      map[string][]ConfigSubscriber
	serviceSubscribersLock sync.RWMutex
	configSubscribersLock  sync.RWMutex
}

// Constructor
func NewConsulRegistry(config *RouterConfig) (*ConsulRegistry, error) {
	client, err := api.NewClient(&(config.Consul))
	if err != nil {
		return nil, err
	}
	return &ConsulRegistry{
		client:             client,
		serviceSubscribers: make(map[string][]ServiceSubscriber),
		configSubscribers:  make(map[string][]ConfigSubscriber),
	}, nil
}

// subscribe service
func (cr *ConsulRegistry) SubscribeService(serviceName string, subscriber ServiceSubscriber) {
	cr.serviceSubscribersLock.RLock()
	arr, ok := cr.serviceSubscribers[serviceName]
	cr.serviceSubscribersLock.RUnlock()
	if !ok {
		cr.serviceSubscribersLock.Lock()
		arr = make([]ServiceSubscriber, 1)
		arr[0] = subscriber
		cr.serviceSubscribers[serviceName] = arr
		cr.serviceSubscribersLock.Unlock()
	} else {
		cr.serviceSubscribersLock.Lock()
		cr.serviceSubscribers[serviceName] = append(arr, subscriber)
		cr.serviceSubscribersLock.Unlock()
	}
}

func (cr *ConsulRegistry) UnsubscribeService(serviceName string) {
	delete(cr.serviceSubscribers, serviceName)
}

// subscribe config
func (cr *ConsulRegistry) SubscribeConfig(serviceName string, subscriber ConfigSubscriber) {
	cr.configSubscribersLock.RLock()
	arr, ok := cr.configSubscribers[serviceName]
	cr.configSubscribersLock.RUnlock()
	if !ok {
		cr.configSubscribersLock.Lock()
		arr = make([]ConfigSubscriber, 1)
		arr[0] = subscriber
		cr.configSubscribers[serviceName] = arr
		cr.configSubscribersLock.Unlock()
	} else {
		cr.configSubscribersLock.Lock()
		cr.configSubscribers[serviceName] = append(arr, subscriber)
		cr.configSubscribersLock.Unlock()
	}
}

func (cr *ConsulRegistry) UnsubscribeConfig(serviceName string) {
	delete(cr.serviceSubscribers, serviceName)
}

// check if service changed
func (cr *ConsulRegistry) checkUpdate(key string, index uint64) bool {
	var lastIndex uint64 = 0
	loaded, ok := cr.lastIndexes.Load(key)
	if ok {
		lastIndex = loaded.(uint64)
	}
	if index <= lastIndex {
		return false
	}
	cr.lastIndexes.Store(key, index)
	return true
}

// ConsulRegistry::discover
func (cr *ConsulRegistry) Discover(serviceName string) {
	server := &Server{
		ServiceName: serviceName,
	}
	kv := cr.client.KV()
	query := &api.QueryOptions{}
	result, meta, err := kv.List(GetNodesPath(server), query)
	if err != nil {
		log.Println(err)
		return
	}
	cr.processDiscover(server, &result, meta)
}

func (cr *ConsulRegistry) processDiscover(server *Server, kvPairs *api.KVPairs, meta *api.QueryMeta) {
	var (
		indexPath   string = GetNodesPath(server)
		isUpdate    bool   = cr.checkUpdate(indexPath, meta.LastIndex)
		length      int    = len(*kvPairs)
		serviceName string = server.ServiceName
	)

	if length == 0 || !isUpdate {
		// log.Printf("The number of %s is %d, updated is %t\n", indexPath, length, isUpdate)
		// TODO fallback
		return
	}

	// get subscriber of current server
	cr.serviceSubscribersLock.RLock()
	subscribers, ok := cr.serviceSubscribers[serviceName]
	cr.serviceSubscribersLock.RUnlock()
	if ok {
		size := len(subscribers)
		wg := sync.WaitGroup{}
		wg.Add(size)
		for _, subscriber := range subscribers {
			go func(subscriber ServiceSubscriber) {
				var servers = make(ServerList, length)
				for idx, pair := range *kvPairs {
					server := &Server{}
					json.Unmarshal(pair.Value, server)
					servers[idx] = *server
				}
				// notify subscriber
				subscriber.ServiceNotify(serviceName, servers)
				wg.Done()
			}(subscriber)
		}
		wg.Wait()
	}
}

func (cr *ConsulRegistry) GetValue(key string) (string, bool) {
	kv := cr.client.KV()
	query := &api.QueryOptions{}
	result, _, err := kv.Get(key, query)
	if err != nil || result == nil {
		return "", false
	} else {
		return string(result.Value), true
	}
}

// ConsulRegistry::getConfig
func (cr *ConsulRegistry) GetConfig(serviceName string) {
	server := &Server{
		ServiceName: serviceName,
	}

	kv := cr.client.KV()
	query := &api.QueryOptions{}
	result, meta, err := kv.List(GetConfigPath(server), query)
	if err != nil {
		log.Println(err)
		return
	}

	cr.processConfig(server, &result, meta)
}

func (cr *ConsulRegistry) processConfig(server *Server, kvPairs *api.KVPairs, meta *api.QueryMeta) {
	var (
		indexPath        string = GetConfigPath(server)
		routerConfigPath string = GetRouterConfigPath(server)
		configsPath      string = GetConfigsPath(server)
		isUpdate         bool   = cr.checkUpdate(indexPath, meta.LastIndex)
		serviceName      string = server.ServiceName
		length           int    = len(*kvPairs)
	)

	if length == 0 || !isUpdate {
		// log.Printf("The number of %s is %d, updated is %t\n", indexPath, length, isUpdate)
		// TODO fallback
		return
	}
	cr.configSubscribersLock.RLock()
	subscribers, ok := cr.configSubscribers[serviceName]
	cr.configSubscribersLock.RUnlock()
	if ok {
		size := len(subscribers)
		wg := sync.WaitGroup{}
		wg.Add(size)
		for _, subscriber := range subscribers {
			go func(subscriber ConfigSubscriber) {
				serviceConfig := ServiceConfig{}
				configs := make(map[string]string)
				for _, sc := range *kvPairs {
					if sc.Key == routerConfigPath && len(sc.Value) != 0 {
						// router config
						routerConfig := ServiceRouterConfig{}
						json.Unmarshal(sc.Value, &routerConfig)
						serviceConfig.Router = routerConfig
					} else if strings.Contains(sc.Key, configsPath) && len(sc.Value) != 0 {
						// configs
						configs[sc.Key] = string(sc.Value)
					}
				}
				serviceConfig.Configs = configs
				subscriber.ConfigNotify(serviceName, serviceConfig)
				wg.Done()
			}(subscriber)
		}
		wg.Wait()
	}
}

func (cr *ConsulRegistry) checkServerValid(server *Server) bool {
	return len(server.ServiceName) != 0
}

func (cr *ConsulRegistry) RegisterServer(server Server) error {
	if !cr.checkServerValid(&server) {
		return ErrInvalidServer
	}
	// updatetime
	now := time.Now().UnixNano() / 1e6
	server.UpdateTime = uint64(now)

	if nil == server.ShardList {
		server.ShardList = make([]uint32, 0)
	}

	if nil == server.AvailableShardList {
		server.AvailableShardList = make([]uint32, 0)
	}

	if nil == server.FollowerShardList {
		server.FollowerShardList = make([]uint32, 0)
	}

	if nil == server.FollowerAvailableShardList {
		server.FollowerAvailableShardList = make([]uint32, 0)
	}

	if nil == server.OtherSettings {
		server.OtherSettings = make(map[string]string)
	}

	json, err := json.Marshal(server)
	if err != nil {
		return err
	}
	kv := cr.client.KV()
	_, err = kv.Put(&api.KVPair{
		Key:   GetNodePath(&server),
		Value: json,
	}, nil)
	if err != nil {
		log.Panicf("Register server %s failed, error is %v\n", server.ServiceName, err)
		return err
	}

	return nil
}

func (cr *ConsulRegistry) UnregisterServer(server Server) error {
	if !cr.checkServerValid(&server) {
		return ErrInvalidServer
	}
	kv := cr.client.KV()
	_, err := kv.Delete(GetNodePath(&server), nil)
	if err != nil {
		log.Panicf("Unregister server %s failed, error is %v\n", server.ServiceName, err)
		return err
	}
	return nil
}

func (cr *ConsulRegistry) UnregisterConfig(server Server) error {
	if !cr.checkServerValid(&server) {
		return ErrInvalidServer
	}
	kv := cr.client.KV()
	_, err := kv.DeleteTree(GetConfigPath(&server), nil)
	if err != nil {
		log.Panicf("Unregister config %s failed, error is %v\n", server.ServiceName, err)
		return err
	}
	return nil
}

func (cr *ConsulRegistry) Available(server Server) error {
	server.Status = ServerStatus_AVAILABLE
	return cr.RegisterServer(server)
}

func (cr *ConsulRegistry) Unavailable(server Server) error {
	server.Status = ServerStatus_UNAVAILABLE
	return cr.UnregisterServer(server)
}

func (cr *ConsulRegistry) PutRouterConfig(server Server, config ServiceRouterConfig) error {
	j, err := json.Marshal(config)
	if err != nil {
		return err
	}
	kv := cr.client.KV()
	_, err = kv.Put(&api.KVPair{
		Key:   GetRouterConfigPath(&server),
		Value: j,
	}, nil)
	return err
}

func (cr *ConsulRegistry) PutServiceConfigs(server Server, configs map[string]string) error {
	kv := cr.client.KV()
	path := GetConfigsPath(&server)
	if nil == configs {
		_, err := kv.Put(&api.KVPair{
			Key:   path,
			Value: []byte{},
		}, nil)
		return err
	} else {
		for k, v := range configs {
			j, err := json.Marshal(v)
			if err != nil {
				return err
			}
			_, err = kv.Put(&api.KVPair{
				Key:   path + k,
				Value: j,
			}, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
