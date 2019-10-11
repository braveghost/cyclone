package main

import (
	"fmt"
	"github.com/braveghost/cyclone"
)

func main() {

	x := &cyclone.MonitorConfig{
		Name: "test_healthy",
		Type: cyclone.MonitorTypeCount,
		Services: []*cyclone.SrvConfigInfo{
			{
				Name: "test_healthy",
				Peak: 1,
			},
		},
		Match: cyclone.MatchTypeEqual,
	}
	r, _ := cyclone.NewRegistry(&cyclone.RegistryConf{"consul", []string{"127.0.0.1:8500"}})
	m, _ := cyclone.NewMonitor(r, x)
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
