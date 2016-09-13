package relay

import v2net "v2ray.com/core/common/net"

type Config struct {
	TargetAddress v2net.Address
	TargetPort    v2net.Port
}