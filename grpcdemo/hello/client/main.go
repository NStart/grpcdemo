package main

import (
	"context"
	"crypto/tls"
	"fmt"
	pb "grpcdemo/proto/hello"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

const (
	Address = "127.0.0.1:50052"
)

func main() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	//NewClientTLSFromFile的第二个参数应该是填ssl证书的common name
	//但是自签名证书会报错，所以下面dial的时候多传了参数tlsConfig用于忽略主机名验证
	creds, err := credentials.NewClientTLSFromFile("../../keys/server.crt", "")
	if err != nil {
		grpclog.Fatalln(err)
	}

	//conn, err := grpc.Dial(Address, grpc.WithInsecure())
	conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(creds), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

	if err != nil {
		grpclog.Fatalln(err)
	}
	defer conn.Close()

	c := pb.NewHelloClient(conn)

	req := &pb.HelloRequest{Name: "grpc"}

	res, err := c.SayHello(context.Background(), req)

	if err != nil {
		grpclog.Fatalln(err)
	}

	fmt.Println(res.Message)

}
