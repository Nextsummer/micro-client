package client

import (
	"github.com/Nextsummer/micro-client/pkg/config"
	pkgrpc "github.com/Nextsummer/micro-client/pkg/grpc"
	"github.com/Nextsummer/micro-client/pkg/log"
	"github.com/Nextsummer/micro-client/pkg/queue"
	"github.com/Nextsummer/micro-client/pkg/utils"
	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map/v2"
	"io"
	"net"
	"sync"
)

type ServerConnectionManager struct {
	serverConnections cmap.ConcurrentMap[string, *ServerConnection]
}

func NewServerConnectionManager() *ServerConnectionManager {
	return &ServerConnectionManager{
		serverConnections: cmap.New[*ServerConnection](),
	}
}

func (s *ServerConnectionManager) addServerConnection(connection *ServerConnection) {
	conn := connection.conn
	s.serverConnections.Set(conn.RemoteAddr().String(), connection)
}

// Check whether a connection has been established to the server address
func (s ServerConnectionManager) hasConnected(server config.Server) bool {
	return s.serverConnections.Has(server.GetRemoteSocketAddress())
}

type ServerConnection struct {
	conn         net.Conn
	connectionId string
	nodeId       int32
}

func NewServerConnection(conn net.Conn) *ServerConnection {
	return &ServerConnection{
		conn:         conn,
		connectionId: uuid.New().String(),
	}
}

func (s *ServerConnection) readMessage() (*pkgrpc.MessageResponse, bool) {
	responseBodyBytes, err := utils.ReadByte(s.conn)

	if err == io.EOF {
		return nil, false
	}
	if err != nil {
		log.Error.Println("Client network io process decode server response failed, err: ", err)
		return nil, false
	}
	response := &pkgrpc.MessageResponse{}
	_ = utils.Decode(responseBodyBytes, response)

	return response, true
}

var serverMessageQueuesOnce sync.Once
var serverMessageQueues *ServerMessageQueues

type ServerMessageQueues struct {
	requestQueues  cmap.ConcurrentMap[string, *queue.Array[pkgrpc.MessageEntity]]
	responseQueues cmap.ConcurrentMap[string, *queue.Array[pkgrpc.MessageResponse]]
}

func GetServerMessageQueuesInstance() *ServerMessageQueues {
	serverMessageQueuesOnce.Do(func() {
		serverMessageQueues = &ServerMessageQueues{
			cmap.New[*queue.Array[pkgrpc.MessageEntity]](),
			cmap.New[*queue.Array[pkgrpc.MessageResponse]](),
		}
	})
	return serverMessageQueues
}

func (s *ServerMessageQueues) init(serverConnectionId string) {
	s.requestQueues.Set(serverConnectionId, queue.NewArray[pkgrpc.MessageEntity]())
	s.responseQueues.Set(serverConnectionId, queue.NewArray[pkgrpc.MessageResponse]())
}

func (s *ServerMessageQueues) putRequest(serverConnectionId string, request *pkgrpc.MessageEntity) {
	q, ok := s.requestQueues.Get(serverConnectionId)
	if !ok {
		return
	}
	q.Put(*request)
}

func (s *ServerMessageQueues) putResponse(serverConnectionId string, response *pkgrpc.MessageResponse) {
	q, ok := s.responseQueues.Get(serverConnectionId)
	if !ok {
		return
	}
	q.Put(*response)
}
