package types

type CodeInfo struct {
	Code string `json:"code"` //code
}

type UrlInfo struct {
	Url string `json:"url"`
}

type ExpireInfo struct {
	Code    string `json:"code"`
	Expired int64  `json:"expired"`
}
type TokenInfo struct { //检验token是否过期的接口使用
	Token string `json:"token"`
}
