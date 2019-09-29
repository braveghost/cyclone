package main

import (
	"context"
	proto "cyclone/proto"
	"fmt"
	"github.com/micro/go-micro"

	"cyclone"

	cli "github.com/micro/go-grpc"
)

var (
	FuserName = "Banner.Carousel"
)
var iClient proto.HealthyService

func GetClient() (proto.HealthyService, error) {
	if iClient == nil {
		reg, _ := cyclone.NewRegistry(&cyclone.RegistryConf{
			Registry:        "consul",
			RegistryAddress: []string{"127.0.0.1:8500"},
		})

		srv := cli.NewService(
			micro.Registry(reg),
		).Client()
		srv.Init()
		iClient = proto.NewHealthyService("test_healthy", srv)
	}
	return iClient, nil
}
func main() {
	iCli, _ := GetClient()

	innerRes, innerErr := iCli.Healthy(
		context.Background(),
		&proto.Request{
		},

		//func(option *client.CallOptions) {
		//	option.Address = []string{"10.xxx.xxx.15:63372"}
		//},
	)
	fmt.Println(innerRes, innerErr)

}
