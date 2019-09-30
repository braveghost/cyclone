package main

import (
	"context"
	proto "cyclone/example/proto"
	healthy "cyclone/healthy"
	"fmt"
	"github.com/micro/go-micro"

	"cyclone"

	cli "github.com/micro/go-grpc"
)

var (
	FuserName = "Banner.Carousel"
)
var iClient healthy.CycloneHealthyService

func GetHealthyClient() (healthy.CycloneHealthyService, error) {
	if iClient == nil {
		reg, _ := cyclone.NewRegistry(&cyclone.RegistryConf{
			Registry:        "consul",
			RegistryAddress: []string{"127.0.0.1:8500"},
		})

		srv := cli.NewService(
			micro.Registry(reg),
		).Client()
		srv.Init()
		iClient = healthy.NewCycloneHealthyService("test_healthy", srv)
	}
	return iClient, nil
}


var iiClient proto.CycloneService


func GetCycloneClient() (proto.CycloneService, error) {
	if iiClient == nil {

		reg, _ := cyclone.NewRegistry(&cyclone.RegistryConf{
			Registry:        "consul",
			RegistryAddress: []string{"127.0.0.1:8500"},
		})

		srv := cli.NewService(
			micro.Registry(reg),
		).Client()
		srv.Init()
		iiClient = proto.NewCycloneService("test_healthy", srv)
	}
	return iiClient, nil
}
func main() {

	x := &cyclone.MonitorConfig{
		Registry: &cyclone.RegistryConf{"consul", []string{"127.0.0.1:8500"}},
		Type:     cyclone.MonitorTypeAddress,
		Services: []*cyclone.SrvConfigInfo{
			{
				Name:  "go.micro.util.srv.zipcode",
				Hosts: []string{"127.0.0.1:54901"},
			},
		},
		Match: cyclone.MatchTypeFull,
	}
	m, _ := cyclone.NewMonitor("TestMonitorAddressFull", x)
	fmt.Println(m.Run())

	iCli, _ := GetHealthyClient()

	innerRes, innerErr := iCli.Healthy(
		context.Background(),
		&healthy.CycloneRequest{
		},

		//func(option *client.CallOptions) {
		//	option.Address = []string{"10.xxx.xxx.15:63372"}
		//},
	)
	fmt.Println(innerRes, innerErr)
	iiCli, _ := GetCycloneClient()

	iinnerRes, iinnerErr := iiCli.Cyclone(
		context.Background(),
		&proto.Request{
		},

		//func(option *client.CallOptions) {
		//	option.Address = []string{"10.xxx.xxx.15:63372"}
		//},
	)
	fmt.Println("xxxxxx",iinnerRes, iinnerErr)

}



