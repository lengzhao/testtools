package grpct

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/fullstorydev/grpcurl"
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
