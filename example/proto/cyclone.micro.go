// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: cyclone.proto

package cyclone_test

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/any"
	math "math"
)

import (
	context "context"
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Cyclone service

type CycloneService interface {
	Cyclone(ctx context.Context, in *Request, opts ...client.CallOption) (*Response, error)
}

type cycloneService struct {
	c    client.Client
	name string
}

func NewCycloneService(name string, c client.Client) CycloneService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "cyclone.test"
	}
	return &cycloneService{
		c:    c,
		name: name,
	}
}

func (c *cycloneService) Cyclone(ctx context.Context, in *Request, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "Cyclone.Cyclone", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Cyclone service

type CycloneHandler interface {
	Cyclone(context.Context, *Request, *Response) error
}

func RegisterCycloneHandler(s server.Server, hdlr CycloneHandler, opts ...server.HandlerOption) error {
	type cyclone interface {
		Cyclone(ctx context.Context, in *Request, out *Response) error
	}
	type Cyclone struct {
		cyclone
	}
	h := &cycloneHandler{hdlr}
	return s.Handle(s.NewHandler(&Cyclone{h}, opts...))
}

type cycloneHandler struct {
	CycloneHandler
}

func (h *cycloneHandler) Cyclone(ctx context.Context, in *Request, out *Response) error {
	return h.CycloneHandler.Cyclone(ctx, in, out)
}
