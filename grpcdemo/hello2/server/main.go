package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	//模块名开头/目录/包名
	pb "grpcdemo/proto/hello"

	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

const (
	Address = "127.0.0.1:50052"
)

type helloService struct {
	pb.UnimplementedHelloServer
}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := &pb.HelloResponse{}
	resp.Message = "hello " + in.Name
	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalln(err)
	}

	var opts []grpc.ServerOption

	creds, err := credentials.NewServerTLSFromFile("../../keys/server.crt", "../../keys/server.key")

	if err != nil {
		grpclog.Fatalln(err)
	}

	opts = append(opts, grpc.Creds(creds))

	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	s := grpc.NewServer(opts...)

	pb.RegisterHelloServer(s, HelloService)

	// 开启trace
	go startTrace()

	fmt.Println("Listen on :" + Address)
	//grpclog.Println("Listen on:" + Address)

	s.Serve(listen)

}

func init() {
	grpc.EnableTracing = true
}

func auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "无token认证信息")
	}

	var (
		appid  string
		appkey string
	)

	if val, ok := md["appid"]; ok {
		appid = val[0]
	}

	if val, ok := md["appkey"]; ok {
		appkey = val[0]
	}

	if appid != "101010" || appkey != "i am key" {
		return grpc.Errorf(codes.Unauthenticated, "Token 认证信息无效，appid = %s, appkey = %", appid, appkey)
	}

	return nil
}

func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (interface{}, error) {
	err := auth(ctx)
	if err != nil {
		return nil, err
	}

	//继续处理请求
	return handle(ctx, req)
}

func startTrace() {
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
	go http.ListenAndServe(":50051", nil)
	fmt.Println("Trace listen on 50051")
}
