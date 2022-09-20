package grpct

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Services map[string]grpcurl.DescriptorSource

func LoadProtos(protoPath string, importPath []string) (Services, error) {
	services := make(Services)
	err := filepath.Walk(protoPath,
		func(fn string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if path.Ext(fn) != ".proto" {
				log.Println("ignore file(not .proto):", fn)
				return nil
			}

			var fileSource grpcurl.DescriptorSource
			fileSource, err = grpcurl.DescriptorSourceFromProtoFiles(importPath, fn)
			if err != nil {
				log.Println("Failed to process proto source files.", err)
				return nil
			}
			svcs, err := grpcurl.ListServices(fileSource)
			if err != nil {
				log.Println("Failed to list services", fn, err)
				return nil
			}
			for _, svc := range svcs {
				fmt.Println("service:", svc)
				services[svc] = fileSource
			}
			return nil
		})
	if err != nil {
		log.Println("Walk", protoPath, err)
		return nil, err
	}
	if len(services) > 0 {
		return services, nil
	}
	return nil, errors.New("not found service")
}

type InvokeCb func(stat *status.Status, response []byte)

// send grpc request by json data
// method = "helloworld.Greeter/SayHello"
func (svcs Services) Invoke(conn *grpc.ClientConn, method string, headers []string, reqData []byte, cb InvokeCb) error {
	var service string
	if strings.Contains(method, "/") {
		service = strings.Split(method, "/")[0]
	} else {
		array := strings.Split(method, ".")
		l := len(array)
		if l < 2 {
			return errors.New("unknow service name")
		}
		service = strings.Join(array[:l-1], ".")
	}

	in := bytes.NewReader(reqData)
	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: false,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    false,
	}
	fileSource, ok := svcs[service]
	if !ok {
		return errors.New("not found service")
	}

	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, fileSource, in, options)
	if err != nil {
		log.Println("Failed to construct request parser and formatter for json", err)
		return err
	}
	buff := bytes.Buffer{}
	h := &grpcurl.DefaultEventHandler{
		Out:            &buff,
		Formatter:      formatter,
		VerbosityLevel: 0,
	}
	ctx := context.Background()

	err = grpcurl.InvokeRPC(ctx, fileSource, conn, method, headers, h, rf.Next)
	if err != nil {
		log.Println("fail to do InvokeRPC:", method, err)
		return err
	}
	cb(h.Status, buff.Bytes())
	return nil
}
