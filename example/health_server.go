package main

import (
	"context"
	"fmt"
	"github.com/braveghost/cyclone"
	proto "github.com/braveghost/cyclone/example/proto"
	cyclone_healthy "github.com/braveghost/cyclone/healthy"
	"github.com/braveghost/joker"
	"github.com/braveghost/meteor/mode"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/service/grpc"
	"log"
)

type testCycloneHandler struct {
}

func (hh testCycloneHandler) Cyclone(ctx context.Context, req *proto.Request, res *proto.Response) error {
	res.Code = proto.Response_Fail
	return nil
}

func CarouselHealthy() (*cyclone_healthy.ApiInfo, error) {
	return nil, nil

}
func main() {

	joker.GetLogger("xxxx", mode.ModeLocal)

	service := grpc.NewService(
		micro.Name("test_healthy"),
	)
	// 注册健康检查函数
	cyclone_healthy.RegistryHealthy(
		&cyclone_healthy.HealthyHandlerConfig{
			ServiceName: "test_healthy",
			Functions: []*cyclone_healthy.HealthyFunction{
				{
					CarouselHealthy,
					"",
				},
			},
		})
	_ = proto.RegisterCycloneHandler(service.Server(), &testCycloneHandler{})
	srv, err := cyclone.NewServiceBuilder(service, nil, &cyclone.Setting{
		//srv, err := cyclone.NewServiceBuilder(service, nil, nil, &cyclone.Setting{
		// 运行态配置, 检查tag
		Masters:  2,
		Interval: 5,
		Tags:     map[string]string{"test_service": "miller"},
		Registry: &cyclone.RegistryConf{
			Registry:        "consul",
			RegistryAddress: []string{"127.0.0.1:8500"},
		},

		// 监控配置
		MonitorConfig: cyclone.McCountEqual("test_healthy",
			&cyclone.SrvConfigInfo{
				Name:  "test_healthy",
				Hosts: []string{"10.60.204.15:52303", "10.60.204.15:52360"},
			}),
	})

	if err == nil {
		srv.RegisterSameStart(func() {
			log.Println("same start")
		})
		if err := srv.Run(); err != nil {
			fmt.Println(err)
		}
	}
}
