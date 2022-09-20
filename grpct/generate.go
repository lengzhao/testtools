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
	os.Mkdir(dir, os.ModePerm)
	log.Println("service:", serviceName)
	methods, err := grpcurl.ListMethods(descriptor, serviceName)
	if err != nil {
		log.Println("not found method:", serviceName, err)
		return err
	}
	log.Println("methods:", methods)
	for _, method := range methods {
		info, err := ShowMethod(descriptor, method)
		if err != nil {
			log.Println("fail to get description of method:", method, err)
			continue
		}
		fn := path.Join(dir, method+".json")
		err = ioutil.WriteFile(fn, []byte(info), os.ModePerm)
		log.Println("generate one file:", fn, err)
	}
	return nil
}

func ShowMethod(descriptor grpcurl.DescriptorSource, method string) (string, error) {
	dsc, err := descriptor.FindSymbol(method)
	if err != nil {
		log.Println("not found method:", method, err)
		return "", err
	}
	mth, ok := dsc.(*desc.MethodDescriptor)
	if !ok {
		log.Println("error type of method:", method)
		return "", err
	}
	var tc Case
	tc.Name = method
	tc.Method = method
	tc.Service = mth.GetService().GetFullyQualifiedName()
	tc.Headers = make([]string, 0)

	inMsg := dynamic.NewMessage(mth.GetInputType())
	updateDefaultValue(inMsg, descriptor, 20)
	dumpParam := jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
		OrigName:     true,
		Indent:       "  ",
	}
	inData, _ := inMsg.MarshalJSONPB(&dumpParam)
	tc.Request = grpcParam{data: inData}

	outMsg := dynamic.NewMessage(mth.GetOutputType())
	updateDefaultValue(outMsg, descriptor, 20)
	outData, _ := outMsg.MarshalJSONPB(&dumpParam)
	tc.Response = grpcParam{data: outData}
	caseData, _ := json.MarshalIndent(tc, "", "  ")

	return string(caseData), nil
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

func updateDefaultValue(msg *dynamic.Message, descriptor grpcurl.DescriptorSource, limit int) error {
	if limit <= 0 {
		return nil
	}
	fields := msg.GetKnownFields()
	for _, field := range fields {
		if field.GetMessageType() == nil {
			continue
		}
		// name := field.GetFullyQualifiedName()
		name := field.GetMessageType().GetFullyQualifiedName()

		childDesc, err := descriptor.FindSymbol(name)
		if err != nil {
			log.Println("fail to find symbol by name:", name, err)
			continue
		}
		cm, ok := childDesc.(*desc.MessageDescriptor)
		if !ok {
			log.Println("it is not MessageDescriptor:", name, err)
			continue
		}
		childMsg := dynamic.NewMessage(cm)
		if msg.GetMessageDescriptor().GetFullyQualifiedName() != name {
			updateDefaultValue(childMsg, descriptor, limit-1)
		}
		if field.IsRepeated() {
			msg.SetFieldByName(field.GetName(), []*dynamic.Message{childMsg})
		} else {
			msg.SetFieldByName(field.GetName(), childMsg)
		}

	}
	return nil
}
