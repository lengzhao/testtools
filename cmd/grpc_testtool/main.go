package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/lengzhao/testtools/grpct"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	address := flag.String("addr", "localhost:50051", "the grpc server address")
	importPath := flag.String("import", "./protos", "import path of proto, split with ','")
	protoPath := flag.String("proto", "./protos", "proto path")
	testcasePath := flag.String("testcase", "./testcase", "testcase path(include json files)")
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
	caseList, err := grpct.LoadTestcases(*testcasePath)
	if err != nil {
		log.Fatal("fail to load testcase from ", *testcasePath, err)
	}

	cc, err := dial(context.Background(), *address)
	if err != nil {
		log.Fatal("fail to dial ", *address, err)
	}
	for _, cs := range caseList {
		log.Println("-*-start to run testcase:", cs.Name)
		err := grpct.RunCase(cc, svcs, cs, nil)
		if err != nil {
			log.Fatalf("fail to run case:%s %s\n", cs.Name, err)
		} else {
			log.Println("---success to run case:", cs.Name)
		}
	}
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
