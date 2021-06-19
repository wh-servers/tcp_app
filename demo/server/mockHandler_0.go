package main

import (
	"context"
	"fmt"
)

func Mock_0_Process(ctx context.Context, reqData, respData interface{}) error {
	req := reqData.([]byte)
	resp := respData.(*[]byte)
	fmt.Println("receive msg: ", string(req))
	feedBack := "feed back from Mock_0 to client "
	data := []byte(feedBack)
	*resp = data
	return nil
}
