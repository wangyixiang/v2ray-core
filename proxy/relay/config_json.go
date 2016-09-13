// +build json

package relay

import (
	"encoding/json"
	"errors"

	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/proxy/registry"
	"v2ray.com/core/common/log"
)

func (this *Config) UnmarshalJSON(data []byte) error {
	type RelayConfig struct {
		Address *v2net.AddressPB `json:"targetAddress"`
		Port    v2net.Port        `json:"targetPort"`
	}
	log.Warning("relay UnmarshalJSON")
	rawConfig := new(RelayConfig)
	if err := json.Unmarshal(data, rawConfig); err != nil {
		return errors.New("yxrelay: Failed to parse config: " + err.Error())
	}

	if rawConfig.Address != nil {
		this.TargetAddress = rawConfig.Address.AsAddress()
	}
	this.TargetPort = rawConfig.Port
	return nil
}

func init() {
	log.Debug("RegisterInboundConfig yxrelay")
	registry.RegisterInboundConfig("yxrelay", func() interface{} {
		return new(Config)
	})
}