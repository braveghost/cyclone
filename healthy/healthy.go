package cyclone_healthy

import (
	"errors"
	"github.com/afex/hystrix-go/hystrix"

	"github.com/micro/go-micro/server"
)

type HealthyFunc func() (*ApiInfo, error)

func RegistryHealthy(s server.Server, hdlr CycloneHealthyHandler, opts ...server.HandlerOption) error {
	return RegisterCycloneHealthyHandler(s, hdlr, opts...)
}

func InitResponse(srvName string, res *CycloneResponse) *CycloneResponse {
	res.Code = CycloneResponse_Healthy
	res.Response = &ServiceStatus{
		Name:    srvName,
		ApiInfo: []*ApiInfo{},
	}
	return &CycloneResponse{
		Code: CycloneResponse_Healthy,
		Response: &ServiceStatus{
			Name:    srvName,
			ApiInfo: []*ApiInfo{},
		},
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
