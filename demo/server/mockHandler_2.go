package main

import (
	"context"
	"fmt"

	app_pb "github.com/wh-servers/tcp_app/gen"
)

func Mock_2_Process(ctx context.Context, reqData, respData interface{}) error {
	req, ok := reqData.(*app_pb.Mock2Request)
	if !ok {
		return fmt.Errorf("wrong req format")
	}
	resp, ok := respData.(*app_pb.Mock2Response)
	if !ok {
		return fmt.Errorf("wrong resp format")
	}
	fmt.Println("receive msg: ", req)
	resp.Keyword = "feed back from Mock_2 handler to client "
	return nil
}
