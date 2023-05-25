package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	gw "grpcdemo/proto/hello_http"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	endpoint := "127.0.0.1:50052"
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterHelloHttpHandlerFromEndpoint(ctx, mux, endpoint, opts)

	if err != nil {
		grpclog.Fatalln("register error")
	}

	fmt.Println("listen on 8080")
	http.ListenAndServe(":8080", mux)

}
