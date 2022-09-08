package grpct

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type grpcParam struct {
	data []byte
}

func (p grpcParam) MarshalJSON() ([]byte, error) {
	if len(p.data) > 0 {
		return p.data, nil
	}
	return nil, errors.New("empty param")
}

func (p *grpcParam) UnmarshalJSON(data []byte) error {
	p.data = data
	return nil
}

func Generate(dir, serviceName string, descriptor grpcurl.DescriptorSource) error {
	dumpParam := jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
		OrigName:     true,
		Indent:       "  ",
	}
	os.Mkdir(dir, os.ModePerm)
	log.Println("service:", serviceName)
	methods, err := grpcurl.ListMethods(descriptor, serviceName)
	if err != nil {
		log.Println("not found method:", serviceName, err)
		return err
	}
	log.Println("methods:", methods)
	for _, method := range methods {
		dsc, err := descriptor.FindSymbol(method)
		if err != nil {
			log.Println("not found method:", method, err)
			continue
		}
		mth, ok := dsc.(*desc.MethodDescriptor)
		if !ok {
			log.Println("error type of method:", method)
			continue
		}
		var tc Case
		tc.Name = method
		tc.Method = method
		tc.Service = serviceName

		inData, _ := dynamic.NewMessage(mth.GetInputType()).MarshalJSONPB(&dumpParam)
		tc.Request = grpcParam{data: inData}

		outData, _ := dynamic.NewMessage(mth.GetOutputType()).MarshalJSONPB(&dumpParam)
		tc.Response = grpcParam{data: outData}
		caseData, _ := json.MarshalIndent(tc, "", "  ")
		fn := path.Join(dir, method+".json")
		err = ioutil.WriteFile(fn, caseData, os.ModePerm)
		log.Println("generate one file:", fn, err)
	}
	return nil
}

func GenerateAll(dir string, svcs Services) error {
	for key, svc := range svcs {
		subDir := path.Join(dir, key)
		err := Generate(subDir, key, svc)
		if err != nil {
			log.Println("fail to generate service:", key)
		}
	}
	return nil
}
