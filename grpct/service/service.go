package service

import (
	"context"
	"log"

	"github.com/lengzhao/proxy/grpc_proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rawService struct {
	respFac ResponseFactory
}

type ResponseFactory func(ctx context.Context, methodName string, reqData []byte) (interface{}, error)

func GetServerOptions(factory ResponseFactory) []grpc.ServerOption {
	var out []grpc.ServerOption
	h := rawService{respFac: factory}

	out = append(out, grpc.UnknownServiceHandler(h.handler))
	out = append(out, grpc.CustomCodec(grpc_proxy.ProxyCodec{}))
	return out
}

func GetStreamHadler(factory ResponseFactory) grpc.StreamHandler {
	h := rawService{respFac: factory}
	return h.handler
}

func (s *rawService) handler(srv interface{}, serverStream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}

	for i := 0; ; i++ {
		f := &grpc_proxy.ProxyData{}
		err := serverStream.RecvMsg(f)
		if err != nil {
			break
		}
		data, _ := f.Marshal()
		resp, err := s.respFac(serverStream.Context(), fullMethodName, data)
		if err != nil {
			log.Println("get error from factory:", fullMethodName, err)
			return err
		}
		err = serverStream.SendMsg(resp)
		if err != nil {
			log.Println("fail to response:", err)
			return err
		}
		log.Println("success to response:", fullMethodName)
	}

	return nil
}
