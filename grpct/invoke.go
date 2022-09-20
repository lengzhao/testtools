package grpct

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type CompareFunc func(respStat *status.Status, respData, hopeData []byte) error

func RunCase(conn *grpc.ClientConn, svcs Services, testcase Case, comp CompareFunc) error {
	var result error
	err := svcs.Invoke(conn, testcase.Method, testcase.Headers, testcase.GetRequest(), func(status *status.Status, response []byte) {
		if comp != nil {
			result = comp(status, response, testcase.GetResponse())
			return
		}
		result = testcase.CompareResponse(response, status)
	})
	if err != nil {
		log.Println("fail to do InvokeRPC:", testcase.Name, err)
		return err
	}
	return result
}
