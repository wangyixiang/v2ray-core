package relay

import (
	"io"

	"v2ray.com/core/transport/internet"
	"v2ray.com/core/app"
	"v2ray.com/core/proxy"
	"v2ray.com/core/proxy/registry"
	"v2ray.com/core/app/dispatcher"
	"v2ray.com/core/common/log"
	v2net "v2ray.com/core/common/net"
	v2io "v2ray.com/core/common/io"
)

type Server struct {
	targetAddress    v2net.Address
	targetPort       v2net.Port
	accepting        bool
	tcpListener      *internet.TCPHub
	packetDispatcher dispatcher.PacketDispatcher
	config           *Config
	meta             *proxy.InboundHandlerMeta
}

func NewRelayServer(config *Config, space app.Space, meta *proxy.InboundHandlerMeta) *Server {
	s := &Server{
		targetAddress: config.TargetAddress,
		targetPort: config.TargetPort,
		accepting: false,
		config: config,
		meta: meta,
	}

	space.InitializeApplication(func() error {
		if !space.HasApp(dispatcher.APP_ID) {
			log.Error("Relay|Server: Dispatcher is not found in the space.")
			return app.ErrMissingApplication
		}
		s.packetDispatcher = space.GetApp(dispatcher.APP_ID).(dispatcher.PacketDispatcher)
		return nil
	})

	return s
}

func (relayServer *Server) Port() v2net.Port {
	return relayServer.meta.Port
}

func (relayServer *Server) Close() {
	if relayServer.tcpListener != nil {
		relayServer.tcpListener.Close()
		relayServer.tcpListener = nil
		relayServer.accepting = false
	}
}

func (relayServer *Server) Start() error {
	if relayServer.accepting == true {
		return nil
	}
	var err error
	relayServer.tcpListener, err = internet.ListenTCP(
		relayServer.meta.Address,
		relayServer.meta.Port,
		relayServer.handleConnection,
		relayServer.meta.StreamSettings,
	)
	if err != nil {
		log.Error("Relay|Server: Fail to listen on", relayServer.meta.Address, ":", relayServer.meta.Port, ":", err)
		return err
	}
	relayServer.accepting = true
	return nil

}

func (relayServer *Server) handleConnection(connection internet.Connection) {
	clientAddr := v2net.DestinationFromAddr(connection.RemoteAddr())
	targetAddr := v2net.TCPDestination(relayServer.targetAddress, relayServer.targetPort)
	relayServer.transport(connection, connection, &proxy.SessionInfo{
		Source: clientAddr,
		Destination: targetAddr,
	})
}

func (relayServer *Server) transport(reader io.Reader, writer io.Writer, session *proxy.SessionInfo) {
	ray := relayServer.packetDispatcher.DispatchToOutbound(relayServer.meta, session)
	input := ray.InboundInput()
	output := ray.InboundOutput()

	go func() {
		v2reader := v2io.NewAdaptiveReader(reader)
		defer v2reader.Release()

		v2io.Pipe(v2reader, input)
		input.Close()
	}()

	go func() {
		v2writer := v2io.NewAdaptiveWriter(writer)
		defer v2writer.Release()
		v2io.Pipe(output, v2writer)
		output.Close()
	}()
}

type ServerFactory struct{}

func (serverFactory *ServerFactory) StreamCapability() internet.StreamConnectionType {
	return internet.StreamConnectionTypeRawTCP
}

func (serverFactory *ServerFactory) Create(space app.Space, rawConfig interface{}, meta *proxy.InboundHandlerMeta) (proxy.InboundHandler, error) {
	return NewRelayServer(rawConfig.(*Config), space, meta), nil
}

func init() {
	registry.MustRegisterInboundHandlerCreator("yxrelay", new(ServerFactory))
}
