package config

import (
	"fmt"
	"github.com/Nextsummer/micro-client/pkg/log"
	"github.com/Nextsummer/micro-client/pkg/queue"
	"github.com/Nextsummer/micro-client/pkg/utils"
	"strconv"
	"strings"
	"sync"
)

const (
	ControllerCandidateServers = "controller.candidate.servers" // List of machines for the server node
	ServiceName                = "service.name"                 // server name of the server node
	ServiceInstanceIp          = "service.instance.ip"          // ip address of the service instance
	ServiceInstancePort        = "service.instance.port"        // port of the service instance
	HeartbeatInterval          = "heartbeat.interval"           // Interval for sending heartbeats
)

var configurationOnce sync.Once
var configuration *Configuration

type Configuration struct {
	ControllerCandidates queue.Array[Server] // controller List of candidate nodes
	ServiceName          string
	ServiceInstanceIp    string
	ServiceInstancePort  int32
	HeartbeatInterval    int32
}

func GetConfigurationInstance() *Configuration {
	configurationOnce.Do(func() {
		configuration = &Configuration{
			ControllerCandidates: *queue.NewArray[Server](),
		}
	})
	return configuration
}

func (c *Configuration) Init(configPath string) {
	configMap := utils.LoadConfigurationFile(configPath)
	servers := validateServers(configMap[ControllerCandidateServers])
	serversArray := strings.Split(servers, ",")
	for _, server := range serversArray {
		serverSplit := strings.Split(server, ":")
		port, _ := strconv.ParseInt(serverSplit[1], 10, 32)
		c.ControllerCandidates.Put(Server{Address: serverSplit[0], Port: int32(port)})
	}

	c.ServiceName = validateServiceName(configMap[ServiceName])
	c.ServiceInstanceIp = validateServiceInstanceIp(configMap[ServiceInstanceIp])
	c.ServiceInstancePort = validateServiceInstancePort(configMap[ServiceInstancePort])
	c.HeartbeatInterval = validateHeartbeatInterval(configMap[HeartbeatInterval])
}

func validateServers(servers any) string {
	controllerCandidateServers := utils.ValidateStringType(servers, ControllerCandidateServers)
	controllerCandidateServersSplit := strings.Split(controllerCandidateServers, ",")
	if controllerCandidateServersSplit == nil || len(controllerCandidateServers) == 0 {
		log.Error.Fatalln("servers cannot be empty...")
	}

	for _, server := range controllerCandidateServersSplit {
		utils.ValidateRegexp(server, "(\\d+\\.\\d+\\.\\d+\\.\\d+):(\\d+)", ControllerCandidateServers)
	}
	return controllerCandidateServers
}

func validateServiceName(serviceName any) string {
	return utils.ValidateStringType(serviceName, ServiceName)
}

func validateServiceInstanceIp(serviceInstanceIp any) string {
	ip := utils.ValidateStringType(serviceInstanceIp, ServiceInstanceIp)
	utils.ValidateRegexp(ip, "(\\d+\\.\\d+\\.\\d+\\.\\d+)", ServiceInstanceIp)
	return ip
}

func validateServiceInstancePort(serviceInstancePort any) int32 {
	port := utils.ValidateStringType(serviceInstancePort, ServiceInstancePort)
	utils.ValidateRegexp(port, "(\\d+)", ServiceInstancePort)
	return utils.ValidateIntType(port, ServiceInstancePort)
}

func validateHeartbeatInterval(heartbeatInterval any) int32 {
	interval := utils.ValidateStringType(heartbeatInterval, HeartbeatInterval)
	utils.ValidateRegexp(interval, "(\\d+)", HeartbeatInterval)
	return utils.ValidateIntType(heartbeatInterval, HeartbeatInterval)
}

func PrintConfigLog() {
	configuration := GetConfigurationInstance()
	log.Info.Printf("[%s]=%s", ControllerCandidateServers, configuration.ControllerCandidates.Iter())
	log.Info.Printf("[%s]=%s", ServiceName, configuration.ServiceName)
	log.Info.Printf("[%s]=%s", ServiceInstanceIp, configuration.ServiceInstanceIp)
	log.Info.Printf("[%s]=%d", ServiceInstancePort, configuration.ServiceInstancePort)
	log.Info.Printf("[%s]=%d", HeartbeatInterval, configuration.HeartbeatInterval)
}

type Server struct {
	Id      int32  `json:"id"`
	Address string `json:"address"`
	Port    int32  `json:"port"`
}

func NewServer(address string, port int32) *Server {
	return &Server{Address: address, Port: port}
}

func (s *Server) GetRemoteSocketAddress() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

func (s *Server) SetId(id int32) {
	s.Id = id
}

func (s *Server) GetId() int32 {
	return s.Id
}

func (s *Server) String() string {
	return utils.ToJson(s)
}
