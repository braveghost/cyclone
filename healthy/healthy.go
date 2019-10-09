package cyclone_healthy

import (
	"context"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
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
			res.Response.ApiInfo = append(res.Response.ApiInfo, ai)
		}

		return nil
	})
}

type HealthyHandler struct {
}

func (hh HealthyHandler) Healthy(ctx context.Context, req *CycloneRequest, res *CycloneResponse) error {
	InitResponse(defaultHealthyFunctionConfig.ServiceName, res)

	for _, l := range defaultHealthyFunctionConfig.Functions {
		GetHealthyInfo(l.FuserName, res, l.Func)
	}
	return nil
}

type HealthyHandlerConfig struct {
	ServiceName string
	Functions   []*HealthyFunction
}

type HealthyFunction struct {
	Func      HealthyFunc
	FuserName string
}

// 注册健康检查函数, srvName 服务名称, fuserName 熔断器配置名称, fns 接口健康检查函数
func RegistryHealthy(hhc *HealthyHandlerConfig) {
	defaultHealthyFunctionConfig = hhc

}
