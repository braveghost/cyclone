package cyclone

import (
	"fmt"
	"testing"
)

func TestMonitorAddressFull(t *testing.T) {

	x := &MonitorConfig{
		Registry: &RegistryConf{"consul", []string{"127.0.0.1:8500"}},
		Type:     MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			&SrvConfigInfo{
				Name:  "go.micro.util.srv.banner",
				Hosts: []string{"127.0.0.1:54901"},
			},
		},
		Match: MatchTypeFull,
	}
	m, _ := NewMonitor("TestMonitorAddressFull", x)
	fmt.Println(m.Run())

}

func TestMonitorAddressIn(t *testing.T) {

	x := &MonitorConfig{
		Registry: &RegistryConf{"consul", []string{"127.0.0.1:8500"}},
		Type:     MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			&SrvConfigInfo{
				Name:  "go.micro.util.srv.banner",
				Hosts: []string{"27.0.0.1"},
			},
		},
		Match: MatchTypeIn,
	}
	m, _ := NewMonitor("TestMonitorAddressIn", x)
	fmt.Println(m.Run())

}

func TestMonitorAddressPrefix(t *testing.T) {

	x := &MonitorConfig{
		Registry: &RegistryConf{"consul", []string{"127.0.0.1:8500"}},
		Type:     MonitorTypeNode,
		Services: []*SrvConfigInfo{
			&SrvConfigInfo{
				Name:  "go.micro.util.srv.banner",
				Hosts: []string{"Destiny"},
			},
		},
		Match: MatchTypePrefix,
	}
	m, _ := NewMonitor("TestMonitorAddressIn", x)
	fmt.Println(m.Run())

}

func TestMonitorEqual(t *testing.T) {

	x := &MonitorConfig{
		Registry: &RegistryConf{"consul", []string{"127.0.0.1:8500"}},
		Type:     MonitorTypeCount,
		Services: []*SrvConfigInfo{
			&SrvConfigInfo{
				Name:  "go.micro.util.srv.banner",
				Hosts: []string{"127.0.0.1", "127.0.0.1"},
			},
		},
		Match: MatchTypeEqual,
	}
	m, _ := NewMonitor("TestMonitorEqual", x)
	fmt.Println(m.Run())

}

func TestMonitorScope(t *testing.T) {

	x := &MonitorConfig{
		Registry: &RegistryConf{"consul", []string{"127.0.0.1:8500"}},
		Type:     MonitorTypeCount,
		Services: []*SrvConfigInfo{
			&SrvConfigInfo{
				Name:   "go.micro.util.srv.banner",
				Hosts:  []string{"127.0.0.1"},
				Peak:   1,
				Valley: 1,
			},
		},
		Match: MatchTypeScope,
	}
	m, _ := NewMonitor("TestMonitorScope", x)
	fmt.Println(m.Run())
}
