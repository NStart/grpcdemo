package main

import (
	"context"
	"crypto/tls"
	"fmt"
	pb "grpcdemo/proto/hello"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

const (
	Address = "127.0.0.1:50052"

	OpenTLS = true
)

type customerCredential struct {
}

func (c customerCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appid":  "101010",
		"appkey": "i am key",
	}, nil
}

func (c customerCredential) RequireTransportSecurity() bool {
	return OpenTLS
}

func main() {
	var err error
	var opts []grpc.DialOption

	if OpenTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}

		//NewClientTLSFromFile的第二个参数应该是填ssl证书的common name
		//但是自签名证书会报错，所以下面dial的时候多传了参数tlsConfig用于忽略主机名验证
		creds, err := credentials.NewClientTLSFromFile("../../keys/server.crt", "")
		if err != nil {
			grpclog.Fatalln(err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	//指定自定义认证
	opts = append(opts, grpc.WithPerRPCCredentials(new(customerCredential)))

	//指定客户端interceptor
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor))

	//conn, err := grpc.Dial(Address, grpc.WithInsecure())
	conn, err := grpc.Dial(Address, opts...)

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

func interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	fmt.Println(ctx, method, req, reply, cc, opts, time.Since(start))
	return err
}
