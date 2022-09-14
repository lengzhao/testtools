package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/lengzhao/testtools/grpct"
	"github.com/lengzhao/testtools/grpct/service"
	"google.golang.org/grpc"
)

type Config struct {
	ImportPath   []string `json:"import_path,omitempty"`
	ProtoPath    string   `json:"proto_path,omitempty"`
	TestcasePath string   `json:"testcase_path,omitempty"`
}

func main() {
	port := flag.Int("port", 50051, "The server port")
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
	factory, err := service.RespFactoryByTestcase(svcs, caseList)
	if err != nil {
		log.Fatal("fail to new factory ", *testcasePath, err)
	}

	ops := service.GetServerOptions(factory)
	server := grpc.NewServer(ops...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
