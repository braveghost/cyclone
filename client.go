package cyclone

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"strconv"
)

type matchConsul struct {
	client *consul.Client
}

func (m *matchConsul) initClient(rc *RegistryConf) error {
	addr := rc.RegistryAddress
	if len(addr) == 0 {
		return MonitorAddrISNullErr
	}

	config := consul.DefaultConfig()
	config.Address = addr[0] //consul server

	var err error
	m.client, err = consul.NewClient(config)
	if err != nil {
		return errors.Wrap(err, "create consul client error")
	}
	return nil

}

func (m *matchConsul) HealthService(name string) *srvBaseInfo {

	var sbi = &srvBaseInfo{}
	srvList, _, err := m.client.Catalog().Service(name, "", &consul.QueryOptions{
		AllowStale: true,
	})
	if err != nil {
		sbi.Err = err
		return sbi
	}

	sbi.Health = len(srvList)
	for _, l := range srvList {
		sbi.Active = append(sbi.Active, &srvInfo{
			Node:    l.Node,
			Address: l.Address + ":" + strconv.Itoa(l.ServicePort),
		})
	}

	return sbi
}
