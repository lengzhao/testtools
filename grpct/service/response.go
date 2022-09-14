package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/lengzhao/testtools/grpct"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultKey = "default"

type respItem struct {
	msg proto.Message
	err error
}

type responseMap struct {
	others map[string]respItem
}

type dynamicResponse struct {
	svcs      grpct.Services
	testcase  grpct.CaseSlice
	responses map[string]*responseMap
}

func RespFactoryByTestcase(svcs grpct.Services, caseSlice grpct.CaseSlice) (ResponseFactory, error) {
	out := dynamicResponse{svcs: svcs, testcase: caseSlice}
	out.responses = make(map[string]*responseMap)
	for _, it := range caseSlice {
		responses, ok := out.responses[it.Method]
		if !ok {
			responses = &responseMap{}
			responses.others = make(map[string]respItem)
			out.responses[it.Method] = responses
		}
		options := grpcurl.FormatOptions{
			EmitJSONDefaultFields: false,
			IncludeTextSeparator:  true,
			AllowUnknownFields:    false,
		}
		fileSource, ok := svcs[it.Service]
		if !ok {
			log.Println("not found service:", it.Service)
			continue
		}
		dsc, err := fileSource.FindSymbol(it.Method)
		if err != nil {
			log.Println("not found method:", it.Method, err)
			continue
		}
		mth, ok := dsc.(*desc.MethodDescriptor)
		if !ok {
			log.Println("error type of method:", it.Method)
			continue
		}

		var respMsg respItem

		if it.ErrorCode == codes.OK {
			out := bytes.NewReader(it.GetResponse())
			resp := grpcurl.MakeTemplate(mth.GetOutputType())

			rf, _, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, fileSource, out, options)
			if err != nil {
				log.Println("Failed to construct response parser and formatter for json", err)
				continue
			}
			err = rf.Next(resp)
			if err != nil {
				log.Println("fail to rf.Next response:", it.Method, err)
				continue
			}
			respMsg.msg = resp
		} else {
			respMsg.err = status.Error(it.ErrorCode, it.Error)
		}
		{
			reqData := it.GetRequest()
			if len(reqData) == 3 && string(reqData) == "\"*\"" {
				responses.others[defaultKey] = respMsg
				continue
			}
			in := bytes.NewReader(it.GetRequest())

			req := grpcurl.MakeTemplate(mth.GetInputType())

			rf, _, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, fileSource, in, options)
			if err != nil {
				log.Println("Failed to construct request parser and formatter for json", err)
				continue
			}
			err = rf.Next(req)
			if err != nil {
				log.Println("fail to rf.Next request:", it.Method, err)
				continue
			}
			data, err := proto.Marshal(req)
			if err != nil {
				log.Println("fail to marshal request:", it.Method, err)
				continue
			}

			h := SHA1(data)
			// log.Println("hash of request:", h)
			responses.others[h] = respMsg
		}

	}
	return out.handle, nil
}

func (d dynamicResponse) handle(ctx context.Context, methodName string, reqData []byte) (interface{}, error) {
	h := SHA1(reqData)
	method := strings.ReplaceAll(methodName, "/", ".")
	respMap, ok := d.responses[method[1:]]
	if !ok {
		log.Println("method not found:", method, methodName)
		return nil, status.Error(codes.Unimplemented, "not found method")
	}
	resp, ok := respMap.others[h]
	if !ok {
		resp, ok = respMap.others[defaultKey]
		if !ok {
			return nil, status.Error(codes.Unknown, "not found response")
		}
	}

	return resp.msg, resp.err
}
func SHA1(input []byte) string {
	c := sha1.New()
	c.Write(input)
	bytes := c.Sum(nil)
	return hex.EncodeToString(bytes)
}
