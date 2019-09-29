package main

import (
	"context"
	"cyclone"
	proto "cyclone/example/proto"
	healthy "cyclone/healthy"
	"fmt"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
)

type testHealthyHandler struct {
}

func (hh testHealthyHandler) Healthy(ctx context.Context, req *healthy.CycloneRequest, res *healthy.CycloneResponse) error {
	res.Response = &healthy.ServiceStatus{
		Name: "healthy",
		Api: []*healthy.ApiInfo{
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
	_ = proto.RegisterHealthyHandler(service.Server(), &testHealthyHandler{})
	_ = proto.RegisterCycloneHandler(service.Server(), &testCycloneHandler{})
	srv, err := cyclone.NewServiceBuilder(service, nil, &cyclone.Setting{
		Threshold: 5,  // 计数器阈值, 溢出后表服务不可用
		Duration:  30, // 计数器统计时间周期, 距离当前多少秒内
		Masters:   2,
		Interval:  5,
		Tags:      map[string]string{"test_service": "miller"},
		Registry: &cyclone.RegistryConf{
			Registry:        "consul",
			RegistryAddress: []string{"127.0.0.1:8500"},
		},
	})
	if err == nil {
		if err := srv.Run(); err != nil {
			fmt.Println(err)
		}
	}
}
