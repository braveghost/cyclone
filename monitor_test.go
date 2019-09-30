package cyclone

import (
	"fmt"
	"testing"
)

func TestMonitorDelNodeFull(t *testing.T) {
	x := &MonitorConfig{
		Name: "test_healthy",
		Type: MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			{
				Name:  "test_healthy",
				Hosts: []string{"10.60.204.15:52303", "10.60.204.15:52360"},
			},
		},
		Match: MatchTypeFull,
	}

	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)

	srvs, _ := m.GetService("test_healthy")
	fmt.Println(srvs[0].Nodes)
	m.removeNode(srvs[0].Name, srvs[0].Nodes[0])

	sss, _ := m.GetService("test_healthy")
	fmt.Println(sss[0].Nodes)

}
func TestMonitorAddressFull(t *testing.T) {

	x := &MonitorConfig{
		Name: "test_healthy",
		Type: MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			{
				Name:  "test_healthy",
				Hosts: []string{"10.60.204.15:52303", "10.60.204.15:52360"},
			},
		},
		Match: MatchTypeFull,
	}

	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	fmt.Println(m.Run())

}

func TestMonitorAddressIn(t *testing.T) {

	x := &MonitorConfig{
		Name: "test_healthy",

		Type: MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			{
				Name:  "test_healthy",
				Hosts: []string{"0.60.204.15"},
			},
		},
		Match: MatchTypeIn,
	}
	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	fmt.Println(m.Run())

}

func TestMonitorAddressPrefix(t *testing.T) {

	x := &MonitorConfig{
		Name: "test_healthy",
		Type: MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			{
				Name:  "test_healthy",
				Hosts: []string{"10.60.204.15"},
			},
		},
		Match: MatchTypePrefix,
	}
	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	fmt.Println(m.Run())

}

func TestMonitorEqual(t *testing.T) {
	x := &MonitorConfig{
		Name: "test_healthy",
		Type: MonitorTypeCount,
		Services: []*SrvConfigInfo{
			{
				Name:  "test_healthy",
				Hosts: []string{"127.0.0.1", "127.0.0.2"},
			},
		},
		Match: MatchTypeEqual,
	}
	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	fmt.Println(m.Run())

}

func TestMonitorScope(t *testing.T) {

	x := &MonitorConfig{
		Name: "test_healthy",
		Type: MonitorTypeCount,
		Services: []*SrvConfigInfo{
			{
				Name:   "test_healthy",
				Peak:   2,
				Valley: 1,
			},
		},
		Match: MatchTypeScope,
	}
	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	fmt.Println(m.Run())
}
