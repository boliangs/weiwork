package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"net/http"
	"sendswork/app/gateway/rpc"
	"sendswork/app/gateway/types"
	"sendswork/config"
	userPb "sendswork/idl/pb/user"
	"sendswork/utils"
	"strconv"
	"time"
)

var userKey = []byte("wqrggyyds")

func UserLogin(c *gin.Context) {

	//-----------当前接入的是群众登录接口

	json := types.ExpireInfo{}
	c.BindJSON(&json)
	code := json.Code
	logrus.Info(code)
	userReq := userPb.UserLoginRequest{Code: code}
	//resp, err := rpc.UserLogin(c, &userReq)
	resp, err := rpc.MassesLogin(c, &userReq)
	if err != nil {
		if err.Error() == "rpc error: code = Unknown desc =找不到学号，请绑定桑梓微助手！" {
			types.ResponseErrorWithMsg(c, http.StatusForbidden, errors.New("找不到学号，请绑定桑梓微助手！"))
			return
		}
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	user, err := utils.AnalyseMassesToken(resp.Token) //解析token
	nowTime := time.Now()

	ctx := context.WithValue(c.Request.Context(), "nowTime", nowTime)
	c.Request = c.Request.WithContext(ctx)
	if json.Expired == 0 {
		json.Expired = 86400
	}
	expireTime := nowTime.Add(time.Duration(json.Expired) * time.Second).Unix() //计算过期时间
	user.Exp = expireTime
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	resp.Token, err = token.SignedString(userKey)
	response := gin.H{
		"token":   resp.Token,
		"expired": user.Exp,
	}
	types.ResponseSuccess(c, response)
}

func SchoolUserLogin(c *gin.Context) {
	json := types.ExpireInfo{}
	c.BindJSON(&json)
	code := json.Code
	userReq := userPb.UserLoginRequest{Code: code}
	//resp, err := rpc.UserLogin(c, &userReq)
	resp, err := rpc.SchoolUserLoginRpc(c, &userReq)
	if err != nil {
		if err.Error() == "rpc error: code = Unknown desc =登录失败请绑定桑梓微助手" {
			types.ResponseErrorWithMsg(c, http.StatusForbidden, errors.New("找不到学号，请绑定桑梓微助手！"))
			return
		}
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	user, err := utils.AnalyseUserToken(resp.Token) //解析token
	nowTime := time.Now()

	ctx := context.WithValue(c.Request.Context(), "nowTime", nowTime)
	c.Request = c.Request.WithContext(ctx)
	if json.Expired == 0 {
		json.Expired = 86400
	}
	expireTime := nowTime.Add(time.Duration(json.Expired) * time.Second).Unix() //计算过期时间
	user.Exp = expireTime
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	resp.Token, err = token.SignedString(userKey)
	response := gin.H{
		"token":   resp.Token,
		"expired": user.Exp,
	}
	types.ResponseSuccess(c, response)
}
func CheckTokenExpire(c *gin.Context) { //为前端提供单独检验token的接口
	json := types.TokenInfo{}
	c.BindJSON(&json)
	token := json.Token
	_, err := utils.AnalyseUserToken(token)
	if err != nil {
		if err.Error() == "Token已过期" {
			response := gin.H{
				"effective": true,
				"expired":   true,
			}
			types.ResponseSuccess(c, response)
			return
		}
		response := gin.H{
			"effective": false,
			"expired":   false,
		}
		types.ResponseSuccess(c, response)
		return
	}
	response := gin.H{
		"effective": true,
		"expired":   false,
	}
	types.ResponseSuccess(c, response)
}
func UserGetJssdk(c *gin.Context) {
	json := types.UrlInfo{}
	c.ShouldBindJSON(&json)
	resp, err := rpc.UserJsTicketRpc(c.Request.Context())
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
	}
	jsApiTicket := resp.JsTicket
	// 生成一个 13 位的随机字符串作为 nonceStr
	nonceStr, err := utils.RandString(13)
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err.Error())
		return
	}
	// 获取当前时间戳（单位为秒）作为 timestamp
	timestamp := time.Now().Unix()
	s := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s", jsApiTicket, nonceStr, strconv.FormatInt(timestamp, 10), json.Url)
	signature := utils.GetSHA1(s)

	// 创建一个 map 类型的变量，用来存储 JS-SDK 配置信息
	data := make(map[string]interface{})

	// 将 appId, timestamp, nonceStr, signature 添加到 map 中
	data["appId"] = config.AppId
	data["timestamp"] = timestamp
	data["nonceStr"] = nonceStr
	data["signature"] = signature
	types.ResponseSuccess(c, data)
}

func WxJSSDK(c *gin.Context) {
	json := types.UrlInfo{}
	c.ShouldBindJSON(&json)
	req := userPb.WxJSSDKRequest{Url: json.Url}
	resp, err := rpc.WxJSSDKRpc(c, &req)
	if err != nil {
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, err)
	}
	types.ResponseSuccess(c, resp)
}

func YearBillLogin(c *gin.Context) {
	json := types.CodeInfo{}
	c.ShouldBindJSON(&json)
	code := json.Code
	logrus.Info(code)
	userReq := userPb.UserLoginRequest{Code: code}
	resp, err := rpc.YearBillLoginRpc(c, &userReq)
	if err != nil {
		if err.Error() == "rpc error: code = Unknown desc =找不到学号，请绑定桑梓微助手！" {
			types.ResponseErrorWithMsg(c, http.StatusForbidden, errors.New("找不到学号，请绑定桑梓微助手！"))
			return
		}
		types.ResponseErrorWithMsg(c, types.CodeServerBusy, "请绑定桑梓微助手！")
		return
	}
	types.ResponseSuccess(c, resp)
}
