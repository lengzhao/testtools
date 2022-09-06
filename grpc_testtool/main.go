package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Config struct {
	ImportPath   []string `json:"import_path,omitempty"`
	ProtoPath    string   `json:"proto_path,omitempty"`
	TestcasePath string   `json:"testcase_path,omitempty"`
}

var services map[string]grpcurl.DescriptorSource

var conf Config = Config{
	ImportPath:   []string{"./protos"},
	ProtoPath:    "./protos",
	TestcasePath: "./testcase",
}

func main() {
	address := flag.String("addr", "localhost:50051", "the grpc server address")
	confFile := flag.String("conf", "./conf.json", "config file")
	bGenerate := flag.Bool("gen", true, "generate testcase")

	flag.Parse()

	if len(*confFile) > 0 {
		data, err := ioutil.ReadFile(*confFile)
		if err != nil {
			log.Println("fail to read config file:", *confFile, err)
			return
		}
		err = json.Unmarshal(data, &conf)
		if err != nil {
			log.Println("fail to Unmarshal config file:", *confFile, err)
			return
		}
	}

	services = make(map[string]grpcurl.DescriptorSource)
	err := filepath.Walk(conf.ProtoPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			var fileSource grpcurl.DescriptorSource
			fileSource, err = grpcurl.DescriptorSourceFromProtoFiles(conf.ImportPath, path)
			if err != nil {
				log.Println("Failed to process proto source files.", err)
				return nil
			}
			svcs, err := grpcurl.ListServices(fileSource)
			if err != nil {
				log.Println("Failed to list services", path, err)
				return nil
			}
			for _, svc := range svcs {
				fmt.Println("service:", svc)
				services[svc] = fileSource
			}
			return nil
		})
	if err != nil {
		log.Println("Walk", conf.ProtoPath, err)
	}
	if *bGenerate {
		// array := []string{"helloworld.Greeter", "helloworld.HelloRequest"}
		array := []string{"helloworld.HelloRequest"}
		// for key, descSource := range services {
		for _, key := range array {
			descSource := services["helloworld.Greeter"]
			log.Println("service:", key)
			dsc, err := descSource.FindSymbol(key)
			if err != nil {
				log.Println("Failed to resolve symbol ", key, err)
				continue
			}
			fqn := dsc.GetFullyQualifiedName()
			log.Println("  fqn:", fqn)
			var elementType string
			switch d := dsc.(type) {
			case *desc.MessageDescriptor:
				elementType = "a message"
				parent, ok := d.GetParent().(*desc.MessageDescriptor)
				if ok {
					if d.IsMapEntry() {
						for _, f := range parent.GetFields() {
							if f.IsMap() && f.GetMessageType() == d {
								// found it: describe the map field instead
								elementType = "the entry type for a map field"
								dsc = f
								break
							}
						}
					} else {
						// see if it's a group
						for _, f := range parent.GetFields() {
							if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_GROUP && f.GetMessageType() == d {
								// found it: describe the map field instead
								elementType = "the type of a group field"
								dsc = f
								break
							}
						}
					}
				}
			case *desc.FieldDescriptor:
				elementType = "a field"
				if d.GetType() == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
					elementType = "a group field"
				} else if d.IsExtension() {
					elementType = "an extension"
				}
			case *desc.OneOfDescriptor:
				elementType = "a one-of"
			case *desc.EnumDescriptor:
				elementType = "an enum"
			case *desc.EnumValueDescriptor:
				elementType = "an enum value"
			case *desc.ServiceDescriptor:
				elementType = "a service"
			case *desc.MethodDescriptor:
				elementType = "a method"
			default:
				err = fmt.Errorf("descriptor has unrecognized type %T", dsc)
				log.Println("Failed to describe symbol ", key, err)
			}
			txt, err := grpcurl.GetDescriptorText(dsc, descSource)
			if err != nil {
				log.Println("Failed to describe symbol ", key, err)
			}
			fmt.Printf("%s is %s:\n", fqn, elementType)
			fmt.Println(txt)
			// all, err := grpcurl.GetAllFiles(descSource)
			// if err != nil {
			// 	log.Println("Failed to GetAllFiles ", key, err)
			// }
			// for i, it := range all {
			// 	dpr := it.AsFileDescriptorProto()
			// 	log.Println("index:", i)
			// 	src := dpr.GetSourceCodeInfo()
			// 	location := src.GetLocation()
			// 	for j, d := range location {
			// 		log.Println(" dependency:", j, d.String())
			// 	}
			// }
			// descSource.ListServices()
		}
		return
	}
	err = filepath.Walk(conf.TestcasePath,
		func(fn string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			log.Println("fn:", fn)
			if path.Ext(fn) != ".json" {
				log.Println("ignore file(not .json):", fn)
				return nil
			}
			var testcase Case
			data, err := ioutil.ReadFile(fn)
			if err != nil {
				log.Println("fail to read file:", fn, err)
				return nil
			}
			err = json.Unmarshal(data, &testcase)
			if err != nil {
				log.Println("fail to Unmarshal:", fn, err)
				return nil
			}
			svcName := testcase.GetServeceName()
			if svcName == "" {
				log.Println("unknow service name:", fn)
				return nil
			}
			svc := services[svcName]
			if svc == nil {
				log.Println("not found service from proto files:", fn, svcName)
				return nil
			}
			resp, err := invoke(svc, *address, testcase.Method, testcase.Headers, testcase.GetRequest())
			if !testcase.CompareResponse(resp, err) {
				log.Println("different response:", fn, string(resp))
			}
			// log.Println("response1:", string(resp))
			log.Println("pass:", testcase.Name)
			return nil
		})
	if err != nil {
		log.Println("Walk", conf.ProtoPath, err)
	}
}

func invoke(fileSource grpcurl.DescriptorSource, address, method string, headers []string, data []byte) ([]byte, error) {
	in := bytes.NewReader(data)
	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: false,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    false,
	}
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, fileSource, in, options)
	if err != nil {
		log.Println("Failed to construct request parser and formatter for json", err)
		return nil, err
	}
	buff := bytes.Buffer{}
	h := &grpcurl.DefaultEventHandler{
		Out:            &buff,
		Formatter:      formatter,
		VerbosityLevel: 0,
	}
	ctx := context.Background()
	cc, err := dial(ctx, address)
	if err != nil {
		log.Println("fail to dial:", address, err)
		return nil, err
	}

	err = grpcurl.InvokeRPC(ctx, fileSource, cc, method, headers, h, rf.Next)
	if err != nil {
		// log.Println("Error invoking method ", err)
		return nil, err
	}
	return buff.Bytes(), nil
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
