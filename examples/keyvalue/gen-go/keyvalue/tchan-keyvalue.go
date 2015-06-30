// Autogenerated by thrift-gen. Do not modify.
package keyvalue

import (
	"fmt"

	athrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/uber/tchannel/golang/thrift"
)

// Interfaces for the service and client for the services defined in the IDL.

type TChanAdmin interface {
	HealthCheck(ctx thrift.Context) (string, error)
	ClearAll(ctx thrift.Context) error
}

type TChanKeyValue interface {
	Get(ctx thrift.Context, key string) (string, error)
	HealthCheck(ctx thrift.Context) (string, error)
	Set(ctx thrift.Context, key string, value string) error
}

type TChanBaseService interface {
	HealthCheck(ctx thrift.Context) (string, error)
}

// Implementation of a client and service handler.

type tchanAdminClient struct {
	client thrift.TChanClient
}

func NewTChanAdminClient(client thrift.TChanClient) TChanAdmin {
	return &tchanAdminClient{client: client}
}

func (c *tchanAdminClient) HealthCheck(ctx thrift.Context) (string, error) {
	var resp HealthCheckResult
	args := HealthCheckArgs{}
	success, err := c.client.Call(ctx, "Admin", "HealthCheck", &args, &resp)
	if err == nil && !success {
	}

	return resp.GetSuccess(), err
}

func (c *tchanAdminClient) ClearAll(ctx thrift.Context) error {
	var resp ClearAllResult
	args := ClearAllArgs{}
	success, err := c.client.Call(ctx, "Admin", "clearAll", &args, &resp)
	if err == nil && !success {
		if e := resp.NotAuthorized; e != nil {
			err = e
		}
	}

	return err
}

type tchanAdminServer struct {
	handler TChanAdmin
}

func NewTChanAdminServer(handler TChanAdmin) thrift.TChanServer {
	return &tchanAdminServer{handler}
}

func (s *tchanAdminServer) Service() string {
	return "Admin"
}

func (s *tchanAdminServer) Methods() []string {
	return []string{
		"HealthCheck",
		"clearAll",
	}
}

func (s *tchanAdminServer) Handle(ctx thrift.Context, methodName string, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	switch methodName {
	case "HealthCheck":
		return s.handleHealthCheck(ctx, protocol)
	case "clearAll":
		return s.handleClearAll(ctx, protocol)
	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanAdminServer) handleHealthCheck(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HealthCheckArgs
	var res HealthCheckResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.HealthCheck(ctx)

	if err != nil {
		return false, nil, err
	}
	res.Success = &r

	return err == nil, &res, nil
}

func (s *tchanAdminServer) handleClearAll(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req ClearAllArgs
	var res ClearAllResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.ClearAll(ctx)

	if err != nil {
		switch v := err.(type) {
		case *NotAuthorized:
			res.NotAuthorized = v
		default:
			return false, nil, err
		}
	}

	return err == nil, &res, nil
}

type tchanKeyValueClient struct {
	client thrift.TChanClient
}

func NewTChanKeyValueClient(client thrift.TChanClient) TChanKeyValue {
	return &tchanKeyValueClient{client: client}
}

func (c *tchanKeyValueClient) Get(ctx thrift.Context, key string) (string, error) {
	var resp GetResult
	args := GetArgs{
		Key: key,
	}
	success, err := c.client.Call(ctx, "KeyValue", "Get", &args, &resp)
	if err == nil && !success {
		if e := resp.NotFound; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanKeyValueClient) HealthCheck(ctx thrift.Context) (string, error) {
	var resp HealthCheckResult
	args := HealthCheckArgs{}
	success, err := c.client.Call(ctx, "KeyValue", "HealthCheck", &args, &resp)
	if err == nil && !success {
	}

	return resp.GetSuccess(), err
}

func (c *tchanKeyValueClient) Set(ctx thrift.Context, key string, value string) error {
	var resp SetResult
	args := SetArgs{
		Key:   key,
		Value: value,
	}
	success, err := c.client.Call(ctx, "KeyValue", "Set", &args, &resp)
	if err == nil && !success {
	}

	return err
}

type tchanKeyValueServer struct {
	handler TChanKeyValue
}

func NewTChanKeyValueServer(handler TChanKeyValue) thrift.TChanServer {
	return &tchanKeyValueServer{handler}
}

func (s *tchanKeyValueServer) Service() string {
	return "KeyValue"
}

func (s *tchanKeyValueServer) Methods() []string {
	return []string{
		"Get",
		"HealthCheck",
		"Set",
	}
}

func (s *tchanKeyValueServer) Handle(ctx thrift.Context, methodName string, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	switch methodName {
	case "Get":
		return s.handleGet(ctx, protocol)
	case "HealthCheck":
		return s.handleHealthCheck(ctx, protocol)
	case "Set":
		return s.handleSet(ctx, protocol)
	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanKeyValueServer) handleGet(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req GetArgs
	var res GetResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.Get(ctx, req.Key)

	if err != nil {
		switch v := err.(type) {
		case *KeyNotFound:
			res.NotFound = v
		default:
			return false, nil, err
		}
	}
	res.Success = &r

	return err == nil, &res, nil
}

func (s *tchanKeyValueServer) handleHealthCheck(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HealthCheckArgs
	var res HealthCheckResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.HealthCheck(ctx)

	if err != nil {
		return false, nil, err
	}
	res.Success = &r

	return err == nil, &res, nil
}

func (s *tchanKeyValueServer) handleSet(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req SetArgs
	var res SetResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.Set(ctx, req.Key, req.Value)

	if err != nil {
		return false, nil, err
	}

	return err == nil, &res, nil
}

type tchanBaseServiceClient struct {
	client thrift.TChanClient
}

func NewTChanBaseServiceClient(client thrift.TChanClient) TChanBaseService {
	return &tchanBaseServiceClient{client: client}
}

func (c *tchanBaseServiceClient) HealthCheck(ctx thrift.Context) (string, error) {
	var resp HealthCheckResult
	args := HealthCheckArgs{}
	success, err := c.client.Call(ctx, "baseService", "HealthCheck", &args, &resp)
	if err == nil && !success {
	}

	return resp.GetSuccess(), err
}

type tchanBaseServiceServer struct {
	handler TChanBaseService
}

func NewTChanBaseServiceServer(handler TChanBaseService) thrift.TChanServer {
	return &tchanBaseServiceServer{handler}
}

func (s *tchanBaseServiceServer) Service() string {
	return "baseService"
}

func (s *tchanBaseServiceServer) Methods() []string {
	return []string{
		"HealthCheck",
	}
}

func (s *tchanBaseServiceServer) Handle(ctx thrift.Context, methodName string, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	switch methodName {
	case "HealthCheck":
		return s.handleHealthCheck(ctx, protocol)
	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanBaseServiceServer) handleHealthCheck(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req HealthCheckArgs
	var res HealthCheckResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.HealthCheck(ctx)

	if err != nil {
		return false, nil, err
	}
	res.Success = &r

	return err == nil, &res, nil
}
