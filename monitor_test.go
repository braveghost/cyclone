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
				Hosts: []string{"10.60.204.18:59016"},
			},
		},
		Match: MatchTypeFull,
	}

	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	//fmt.Println(m.Run())
	//

	for i := 0; i <= 10; i++ {

		srvs, err := m.GetHealth(&SrvHealthyConfig{
			Name:      "test_healthy",
			Duration:  10,
			Threshold: 5,
		})
		if err != nil {
			fmt.Println(err)
		} else {

			fmt.Println(srvs.Sick, srvs.Healthy)
		}
	}
	////m.removeNode(srvs[0].Name, srvs[0].Nodes[0])
	//
	//sss, _ := m.GetService("test_healthy")
	//fmt.Println(sss[0].Nodes)

}
func TestMonitorAddressFull(t *testing.T) {

	x := &MonitorConfig{
		Name: "go.micro.util.srv.banner",
		Type: MonitorTypeAddress,
		Services: []*SrvConfigInfo{
			{
				Name:  "go.micro.util.srv.banner",
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
				Name:      "test_healthy",
				Hosts:     []string{"127.0.0.1", "127.0.0.2"},
				Duration:  10,
				Threshold: 5,
			},
		},
		Match: MatchTypeEqual,
	}
	r, _ := NewRegistry(&RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := NewMonitor(r, x)
	fmt.Println(m.Run())

	for i := 0; i <= 10; i++ {
		fmt.Println(m.Run())
		//
		//srvs, err := m.GetHealth(&SrvHealthyConfig{
		//	Name:      "test_healthy",
		//	Duration:  10,
		//	Threshold: 5,
		//})
		//if err != nil {
		//	fmt.Println(err)
		//} else {
		//
		//	fmt.Println(srvs.Sick, srvs.Healthy)
		//}
	}
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
	for i := 0; i <= 10; i++ {
		fmt.Println(m.Run())
		//
		//srvs, err := m.GetHealth(&SrvHealthyConfig{
		//	Name:      "test_healthy",
		//	Duration:  10,
		//	Threshold: 5,
		//})
		//if err != nil {
		//	fmt.Println(err)
		//} else {
		//
		//	fmt.Println(srvs.Sick, srvs.Healthy)
		//}
	}
}
