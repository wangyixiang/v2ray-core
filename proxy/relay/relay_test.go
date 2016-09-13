package relay_test

import (
	"net"
	"testing"

	"v2ray.com/core/testing/servers/tcp"
	"v2ray.com/core/app"
	"v2ray.com/core/app/dispatcher"
	dispatcherimpl "v2ray.com/core/app/dispatcher/impl"
	"v2ray.com/core/app/proxyman"
	"v2ray.com/core/proxy/freedom"
	"v2ray.com/core/proxy"
	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/transport/internet"
	. "github.com/wangyixiang/v2ray-core/proxy/relay"
)

func TestRelay(t *testing.T) {
	tcpServerHeadStr := "Recieved: "
	tcpServer := &tcp.Server{
		MsgProcessor: func(data []byte) []byte {
			buffer := make([]byte, 0, 2048)
			buffer = append(buffer, []byte(tcpServerHeadStr)...)
			buffer = append(buffer, data...)
			return buffer
		},
	}

	_, err := tcpServer.Start()
	if err != nil {
		t.Fatal("tcp Server fails on starting.\n")
	}
	defer tcpServer.Close()

	space := app.NewSpace()
	space.BindApp(dispatcher.APP_ID, dispatcherimpl.NewDefaultDispatcher(space))
	ohm := proxyman.NewDefaultOutboundHandlerManager()
	ohm.SetDefaultHandler(
		freedom.NewFreedomConnection(
			&freedom.Config{},
			space,
			&proxy.OutboundHandlerMeta{
				StreamSettings: &internet.StreamSettings{
					Type: internet.StreamConnectionTypeRawTCP,
				},
			},
		),
	)
	space.BindApp(proxyman.APP_ID_OUTBOUND_MANAGER, ohm)

	data2Send := []byte("Data to be sent to remote")

	port := v2net.Port(13214)

	relayServer := NewRelayServer(
		&Config{
			TargetAddress: v2net.LocalHostIP,
			TargetPort: tcpServer.Port,
		}, space, &proxy.InboundHandlerMeta{
			Address: v2net.LocalHostIP,
			Port: port,
			StreamSettings: &internet.StreamSettings{
				Type: internet.StreamConnectionTypeRawTCP,
			},
		},
	)

	defer relayServer.Close()

	err = space.Initialize()

	if err != nil {
		t.Fatal("space initializes failed.\n")
	}

	err = relayServer.Start()

	if err != nil {
		t.Fatal("relayServer starts failed.\n")
	}

	if relayServer.Port() != port {
		t.Fatal("relayServer is using a port that not specified in config.\n")
	}

	tcpClient, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP: []byte{127,0,0,1},
		Port: int(port),
		Zone: "",
	})

	if err != nil {
		t.Fatal("connecting relayServer fails.\n")
	}

	wCount, err := tcpClient.Write(data2Send)

	if err != nil {
		t.Fatal("writing to relayServer fails.\n")
	}

	if wCount != len(data2Send) {
		t.Fatal("writen bytes is different on number with data2Send.")
	}

	response := make([]byte, 2048)
	rCount, err := tcpClient.Read(response)

	if err != nil {
		t.Fatal("reading to relayServer fails.\n")
	}
	tcpClient.Close()
	if tcpServerHeadStr + string(data2Send) != string(response[:rCount]) {
		t.Fail()
		t.Error("The data return from server is not as expected.\n")
	}

}