package cyclone_healthy

import (
	"context"
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/braveghost/joker"
	"google.golang.org/grpc/peer"
	"net"
	"os"
	"strings"
)

var (
	defaultHealthyFunctionConfig *HealthyHandlerConfig
)

type HealthyFunc func() (*ApiInfo, error)

func InitResponse(srvName string, res *CycloneResponse) {
	res.Code = CycloneResponse_Healthy
	res.Response = &ServiceStatus{
		Name:    srvName,
		ApiInfo: []*ApiInfo{},
	}
}

func GetHealthyInfo(name string, res *CycloneResponse, fn HealthyFunc) {
	var (
		ai *ApiInfo
	)

	_ = hystrix.Do(name, func() error {
		var innerErr error
		ai, innerErr = fn()
		if innerErr != nil {
			return innerErr
		}
		return nil
	}, func(err error) error {
		if err != nil {
			if errors.Is(err, hystrix.ErrTimeout) || errors.Is(err, hystrix.ErrCircuitOpen) || errors.Is(err, hystrix.ErrMaxConcurrency) {
				res.Code = CycloneResponse_Zombies
			} else {
				res.Code = CycloneResponse_Sick
			}
			if ai != nil {
				res.Response.ApiInfo = append(res.Response.ApiInfo, ai)
			}
		}

		return nil
	})
}

type HealthyHandler struct {
}

func (hh HealthyHandler) Healthy(ctx context.Context, req *CycloneRequest, res *CycloneResponse) error {
	if defaultHealthyFunctionConfig != nil {
		InitResponse(defaultHealthyFunctionConfig.ServiceName, res)

		for _, l := range defaultHealthyFunctionConfig.Functions {
			GetHealthyInfo(l.FuserName, res, l.Func)
		}

	}
	return nil
}

// todo 关闭的监控
func (hh HealthyHandler) Close(ctx context.Context, req *CycloneRequest, res *CycloneCloseResponse) error {
	ip, err := getClientIP(ctx)
	joker.Warnw("Cyclone.HealthyHandler.Close.Warning", "ip", ip, "err", err)
	os.Exit(0)
	return nil
}

func getClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}

type HealthyHandlerConfig struct {
	ServiceName string             // 服务名称
	Functions   []*HealthyFunction // 接口健康检查函数集合
}

type HealthyFunction struct {
	Func      HealthyFunc // 健康检查函数
	FuserName string      // 熔断器配置名称
}

// 注册健康检查函数
func RegistryHealthy(hhc *HealthyHandlerConfig) {
	if hhc == nil {
		if defaultHealthyFunctionConfig == nil {
			defaultHealthyFunctionConfig = &HealthyHandlerConfig{}
		}
	} else {
		defaultHealthyFunctionConfig = hhc
	}
}
