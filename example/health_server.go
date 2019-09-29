package main

import (
	"context"
	"cyclone"
	proto "cyclone/proto"
	"fmt"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
)

type testHealthyHandler struct {
}

func (hh testHealthyHandler) Healthy(ctx context.Context, req *proto.Request, res *proto.Response) error {
	res.Msg = "SUCCESS"
	res.Code = proto.Response_Healthy
	return nil
}

func main() {
	service := grpc.NewService(
		micro.Name("test_healthy"),
	)
	_ = proto.RegisterHealthyHandler(service.Server(), &testHealthyHandler{})
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
