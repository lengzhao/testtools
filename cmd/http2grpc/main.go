package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/lengzhao/testtools/grpct"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	endponit := flag.String("endponit", "localhost:50051", "the grpc server address(endpoint)")
	addr := flag.String("addr", ":8080", "listen address")
	prefix := flag.String("pre", "/ov/", "the prefix of url path")
	importPath := flag.String("import", "./protos", "import path of proto, split with ','")
	protoPath := flag.String("proto", "./protos", "proto path")
	genPath := flag.String("gen", "", "testcase path, new testcase with null value")

	flag.Parse()

	imps := strings.Split(*importPath, ",")
	svcs, err := grpct.LoadProtos(*protoPath, imps)
	if err != nil {
		log.Fatal("fail to load protos:", err)
	}
	if len(*genPath) > 0 {
		os.Mkdir(*genPath, os.ModePerm)
		grpct.GenerateAll(*genPath, svcs)
		return
	}

	cc, err := dial(context.Background(), *endponit)
	if err != nil {
		log.Fatal("fail to dial ", *endponit, err)
	}

	http.HandleFunc(*prefix, func(w http.ResponseWriter, req *http.Request) {
		log.Println("path:", req.URL.Path)
		u := req.URL.Path
		if len(*prefix) > 1 {
			u = strings.TrimLeft(u, *prefix)
		}
		array := strings.Split(u, "/")
		if len(array) != 2 {
			log.Println("unknow service:", req.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reqData, err := ioutil.ReadAll(req.Body)
		if err != nil {
			reqData = []byte("{}")
		}

		err = svcs.Invoke(cc, u, nil, reqData, func(stat *status.Status, response []byte) {
			if stat.Code() != codes.OK {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, stat.String())
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
		})
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, err.Error())
	})
	http.ListenAndServe(*addr, nil)
}

func dial(ctx context.Context, address string) (*grpc.ClientConn, error) {
	dialTime := 10 * time.Second

	ctx, cancel := context.WithTimeout(ctx, dialTime)
	defer cancel()
	var opts []grpc.DialOption

	var creds credentials.TransportCredentials

	network := "tcp"

	cc, err := grpcurl.BlockingDial(ctx, network, address, creds, opts...)
	if err != nil {
		log.Println("Failed to dial target host ", address, err)
		return nil, err
	}
	return cc, nil
}
