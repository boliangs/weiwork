package service

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"sendswork/app/user/database/cache"
	"sendswork/app/user/database/dao"
	"sendswork/app/user/database/models"
	"sendswork/config"
	userPb "sendswork/idl/pb/user"
	"sendswork/utils"
	"sendswork/utils/school"
	"sync"
	"time"
)

type UserSrv struct {
	userPb.UnimplementedUserServiceServer
}

var UserSrvIns *UserSrv

var UserSrvOnce sync.Once

func GetUserSrv() *UserSrv {
	UserSrvOnce.Do(func() {
		UserSrvIns = &UserSrv{}
	})
	return UserSrvIns
}

func (*UserSrv) UserPing(ctx context.Context, empty *emptypb.Empty) (resp *userPb.UserPingResponse, err error) {
	resp = new(userPb.UserPingResponse)
	resp.Message = "user微服务ping通"
	return
}

func (*UserSrv) UserLogin(ctx context.Context, req *userPb.UserLoginRequest) (resp *userPb.UserLoginResponse, err error) {
	resp = new(userPb.UserLoginResponse)
	wxLoginResp, err := utils.WXLogin(req.Code)
	if err != nil || wxLoginResp.OpenId == "" {
		return
	}
	openid := wxLoginResp.OpenId
	//openid := "oHCMZuJVYI3k1NrgjGaFxZ3a5pk8"
	//处理缓存
	userDao := dao.NewUserDao(ctx)
	data, cnt, err := userDao.ExistUserByOpenid(openid)
	if err != nil {
		return
	}
	if cnt == 0 {
		//stResp := utils.GetStuNumResp{}
		var stResp *utils.GetStuNumResp
		wxInfoResp := &utils.WeChatUser{}
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			stResp, err = utils.GetStuNum(openid)
			//stResp.Stuid = "2125102013"
			if err != nil {
				err = errors.New("找不到学号，请绑定桑梓微助手！")
				return
			}
		}()
		go func() {
			defer wg.Done()
			wxInfoResp, err = utils.GetWeChatInfo(openid, wxLoginResp.AccessToken)
			if err != nil {
				err = fmt.Errorf("%w %w", err, errors.New("获取头像昵称错误！"))
			}
			return
		}()
		wg.Wait()
		if err != nil {
			return
		}
		user := models.User{
			OpenId:   openid,
			StuNum:   stResp.Stuid,
			IsAdmin:  0,
			Avatar:   wxInfoResp.HeadImgURL,
			Nickname: wxInfoResp.Nickname,
		}
		err = userDao.CreateUser(&user)
		if err != nil {
			return
		}
		token := utils.GenerateUserToken(openid, stResp.Stuid)
		resp.Token = token
		return
	}
	token := utils.GenerateUserToken(openid, data[0].StuNum)
	resp.Token = token
	return
}

func (*UserSrv) MassesLogin(ctx context.Context, req *userPb.UserLoginRequest) (resp *userPb.UserLoginResponse, err error) {
	resp = new(userPb.UserLoginResponse)
	wxLoginResp, err := utils.WXLogin(req.Code)
	if err != nil || wxLoginResp.OpenId == "" {
		return
	}
	openid := wxLoginResp.OpenId
	userDao := dao.NewUserDao(ctx)
	data, cnt, err := userDao.ExistUserByOpenid(openid)
	if err != nil {
		return
	}
	if cnt == 0 {
		wxInfoResp, err := utils.GetWeChatInfo(openid, wxLoginResp.AccessToken)
		if err != nil {
			err = fmt.Errorf("%w %w", err, errors.New("获取头像昵称错误！"))
		}
		user := models.User{
			OpenId:   openid,
			IsAdmin:  0,
			Avatar:   wxInfoResp.HeadImgURL,
			Nickname: wxInfoResp.Nickname,
		}
		err = userDao.CreateUser(&user)
		if err != nil {
			return resp, err
		}
		token := utils.GenerateMassesToken(openid, wxInfoResp.Nickname)
		resp.Token = token
		return resp, err
	}
	token := utils.GenerateMassesToken(openid, data[0].Nickname)
	resp.Token = token
	return
}

func (*UserSrv) UserJsTicket(ctx context.Context, empty *emptypb.Empty) (resp *userPb.UserJsTicketResponse, err error) {
	resp = new(userPb.UserJsTicketResponse)
	cacheRDB := cache.NewRDBCache(ctx)
	jsApiTicket, err := cacheRDB.GetValue("jsapi")
	if err != nil {
		accessResp, err := utils.GetAccessToken()
		if err != nil {
			return resp, err
		}
		access_token := accessResp.AccessToken
		if err != nil {
			return resp, err
		}
		jsapiResp, err := utils.GetJsApiTicket(access_token)
		if err != nil {
			return resp, err
		}
		jsApiTicket := jsapiResp.Ticket
		cacheRDB.SaveValue("jsapi", jsApiTicket, 1*time.Hour)
		if err != nil {
			return resp, err
		}
	}
	resp.JsTicket = jsApiTicket
	return
}

func (*UserSrv) SchoolUserLogin(ctx context.Context, req *userPb.UserLoginRequest) (resp *userPb.UserLoginResponse, err error) {
	resp = new(userPb.UserLoginResponse)
	wxLoginResp, err := utils.WXLogin(req.Code)
	if err != nil || wxLoginResp.OpenId == "" {
		return
	}
	openid := wxLoginResp.OpenId
	var stResp *utils.GetStuNumResp
	stResp, err = utils.GetStuNum(openid)
	if err != nil {
		err = errors.New("找不到学号，请绑定桑梓微助手！")
		return
	}
	cacheRDB := cache.NewRDBCache(ctx)
	jwcCertificate := cacheRDB.GetJwcCertificate(stResp.Stuid)
	if jwcCertificate.Emaphome_WEU == "" {
		aes := utils.NewEncryption()
		aes.SetKey(utils.AcquiesceKey2)
		jwcCertificate.GsSession, err = school.GetGsSession(aes.AesEncoding(openid))
		if err != nil {
			return
		}
		aes.SetKey(utils.AcquiesceKey3)
		err, jwcCertificate.GsSession = aes.AesDecoding(jwcCertificate.GsSession)
		if err != nil {
			return
		}
		jwc := school.NewJwc()
		jwc.GsSession = jwcCertificate.GsSession
		jwcCertificate.Emaphome_WEU, err = jwc.GetEmaphome_WEU()
		err = cacheRDB.SaveJwcCertificate(jwcCertificate, stResp.Stuid)
		if err != nil {
			return
		}
	}
	token := utils.GenerateUserToken(openid, stResp.Stuid)
	resp.Token = token
	return
}

func (*UserSrv) WxJSSDK(ctx context.Context, req *userPb.WxJSSDKRequest) (resp *userPb.WxJSSDKResponse, err error) {
	resp = new(userPb.WxJSSDKResponse)
	cacheRDB := cache.NewRDBCache(ctx)
	jsApiTicket, err := cacheRDB.GetValue("jsapi")
	if err != nil {
		accessResp, err := utils.GetAccessToken()
		if err != nil {
			return resp, err
		}
		access_token := accessResp.AccessToken
		if err != nil {
			return resp, err
		}
		jsapiResp, err := utils.GetJsApiTicket(access_token)
		if err != nil {
			return resp, err
		}
		jsApiTicket = jsapiResp.Ticket
		cacheRDB.SaveValue("jsapi", jsApiTicket, 1*time.Hour)
		if err != nil {
			return resp, err
		}
	}
	nonceStr, err := utils.RandString(13)
	if err != nil {
		return
	}
	timestamp := time.Now().Unix()
	resp.Signature = utils.GenerateSignature(jsApiTicket, timestamp, nonceStr, req.Url)
	resp.AppId = config.AppId
	resp.Timestamp = timestamp
	resp.NonceStr = nonceStr
	return
}
