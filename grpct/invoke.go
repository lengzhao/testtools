package grpct

import (
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/fullstorydev/grpcurl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type CompareFunc func(respStat *status.Status, respData, hopeData []byte) error

func RunCase(conn *grpc.ClientConn, svcs Services, testcase Case, comp CompareFunc) error {
	in := bytes.NewReader(testcase.GetRequest())
	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: false,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    false,
	}
	fileSource, ok := svcs[testcase.Service]
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

	err = grpcurl.InvokeRPC(ctx, fileSource, conn, testcase.Method, testcase.Headers, h, rf.Next)
	if err != nil {
		log.Println("fail to do InvokeRPC:", testcase.Name, err)
		return err
	}
	if comp != nil {
		return comp(h.Status, buff.Bytes(), testcase.GetResponse())
	}
	return testcase.CompareResponse(buff.Bytes(), h.Status)
}
