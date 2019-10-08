package main

import (
	"context"
	"cyclone"
	proto "cyclone/example/proto"
	healthy "cyclone/healthy"
	"fmt"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/service/grpc"
)

type testHealthyHandler struct {
}

func (hh testHealthyHandler) Healthy(ctx context.Context, req *healthy.CycloneRequest, res *healthy.CycloneResponse) error {
	res.Response = &healthy.ServiceStatus{
		Name: "healthy",
		ApiInfo: []*healthy.ApiInfo{
			{
				Api:   "healthy",
				Error: "",
			},
		},
	}
	res.Code = healthy.CycloneResponse_Healthy
	return nil
}

type testCycloneHandler struct {
}

func (hh testCycloneHandler) Cyclone(ctx context.Context, req *proto.Request, res *proto.Response) error {
	res.Code = proto.Response_Fail
	return nil
}

func main() {
	service := grpc.NewService(
		micro.Name("test_healthy"),
	)
	_ = proto.RegisterCycloneHandler(service.Server(), &testCycloneHandler{})
	srv, err := cyclone.NewServiceBuilder(service, nil, &testHealthyHandler{}, &cyclone.Setting{
		Masters:  2,
		Interval: 5,
		Tags:     map[string]string{"test_service": "miller"},
		Registry: &cyclone.RegistryConf{
			Registry:        "consul",
			RegistryAddress: []string{"127.0.0.1:8500"},
		},
		MonitorConfig: &cyclone.MonitorConfig{
			Name: "test_healthy",
			Type: cyclone.MonitorTypeCount,
			Services: []*cyclone.SrvConfigInfo{
				{
					Name:  "test_healthy",
					Hosts: []string{"10.60.204.15:52303", "10.60.204.15:52360"},
				},
			},
			Match: cyclone.MatchTypeEqual,
		},
	})
	if err == nil {
		if err := srv.Run(); err != nil {
			fmt.Println(err)
		}
	}
}
