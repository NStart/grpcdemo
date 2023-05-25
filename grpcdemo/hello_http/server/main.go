package main

import (
	"context"
	"fmt"
	"net"

	//模块名开头/目录/包名
	pb "grpcdemo/proto/hello_http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	Address = "127.0.0.1:50052"
)

type helloService struct {
	pb.UnimplementedHelloHttpServer
}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloHTTPRequest) (*pb.HelloHTTPResponse, error) {
	resp := &pb.HelloHTTPResponse{}
	resp.Message = "hello " + in.Name
	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalln(err)
	}

	//creds, err := credentials.NewServerTLSFromFile("../../keys/server.crt", "../../keys/server.key")

	if err != nil {
		grpclog.Fatalln(err)
	}

	//s := grpc.NewServer(grpc.Creds(creds))
	s := grpc.NewServer()

	pb.RegisterHelloHttpServer(s, HelloService)

	fmt.Println("Listen on :" + Address)
	//grpclog.Println("Listen on:" + Address)

	s.Serve(listen)

}
