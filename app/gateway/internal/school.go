package internal

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
	"sendswork/app/gateway/rpc"
	"sendswork/app/gateway/types"
)

func SchoolSchedule(c *gin.Context) {
	jsonValue := types.SemesterInfo{}
	c.ShouldBindJSON(&jsonValue)
	if _, exists := c.Get("stu_num"); !exists {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, errors.New("token无stu_num"))
		return
	}
	md := metadata.Pairs(
		"semester", jsonValue.Semester,
		"stu_num", c.Value("stu_num").(string),
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := rpc.SchoolScheduleRpc(ctx)
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	types.ResponseSuccess(c, resp)
}

func SchoolGpa(c *gin.Context) {
	if _, exists := c.Get("stu_num"); !exists {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, errors.New("token无stu_num"))
		return
	}
	md := metadata.Pairs("stu_num", c.Value("stu_num").(string))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := rpc.SchoolGpaRpc(ctx)
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	types.ResponseSuccess(c, resp)
}

func SchoolGrade(c *gin.Context) {
	if _, exists := c.Get("stu_num"); !exists {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, errors.New("token无stu_num"))
		return
	}
	md := metadata.Pairs("stu_num", c.Value("stu_num").(string))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := rpc.SchoolGradeRpc(ctx)
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	types.ResponseSuccess(c, resp)
}

func SchoolXueFen(c *gin.Context) {
	if _, exists := c.Get("stu_num"); !exists {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, errors.New("token无stu_num"))
		return
	}
	md := metadata.Pairs("stu_num", c.Value("stu_num").(string))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := rpc.SchoolXuefenRpc(ctx)
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	types.ResponseSuccess(c, resp)
}
