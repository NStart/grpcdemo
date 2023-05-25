package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	r "runtime"
	"strings"

	//模块名开头/目录/包名
	pb "grpcdemo/proto/hello_http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

const (
	Address = "127.0.0.1:50052"
)

type hellohttpService struct {
	pb.UnimplementedHelloHttpServer
}

var HellohttpService = hellohttpService{}

func (h hellohttpService) SayHello(ctx context.Context, in *pb.HelloHTTPRequest) (*pb.HelloHTTPResponse, error) {
	resp := &pb.HelloHTTPResponse{}
	resp.Message = "hello " + in.Name
	return resp, nil
}

func main() {
	endpoint := "127.0.0.1:50052"
	conn, err := net.Listen("tcp", endpoint)
	if err != nil {
		grpclog.Fatalln(err)
	}

	//grpc tls server
	creds, err := credentials.NewServerTLSFromFile("../../keys/server.crt", "../../keys/server.key")
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHelloHttpServer(grpcServer, HellohttpService)

	//gw server
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	ctx := context.Background()
	dcreds, err := credentials.NewClientTLSFromFile("../../keys/server.crt", "")
	if err != nil {
		grpclog.Fatalln(err)
	}
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))}
	gmux := runtime.NewServeMux()
	if err := pb.RegisterHelloHttpHandlerFromEndpoint(ctx, gmux, endpoint, dopts); err != nil {
		grpclog.Fatalln(err)
	}

	// http服务
	mux := http.NewServeMux()
	mux.Handle("/", gmux)

	srv := http.Server{
		Addr:      endpoint,
		Handler:   grpcHandlerFunc(grpcServer, mux),
		TLSConfig: getTLSConfig(),
	}

	fmt.Printf("grpc and http listen on %s\n", endpoint)
	if err = srv.Serve(tls.NewListener(conn, srv.TLSConfig)); err != nil {
		grpclog.Fatalln(err)
	}

	return
}

func getTLSConfig() *tls.Config {
	cert, err := ioutil.ReadFile("../../keys/server.crt")
	if err != nil {
		grpclog.Fatalln(err)
	}

	key, err := ioutil.ReadFile("../../keys/server.key")
	if err != nil {
		grpclog.Fatalln(err)
	}

	var demoKeypair *tls.Certificate
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		grpclog.Fatalln(err)
	}
	demoKeypair = &pair
	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{*demoKeypair},
		NextProtos:         []string{http2.NextProtoTLS},
	}
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func printLineNum() {
	_, _, line, ok := r.Caller(1)
	if ok {
		fmt.Println("current line is :", line)
	} else {
		fmt.Println("fail to print line number")
	}
}
