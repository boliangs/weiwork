package config

import "time"

var (
	AppId     = "**********"
	AppSecret = "**********"
)

// id和密钥
var (
	Name     = "HQUadmin"
	Password = "hucr&sends"
)

var (
	StartTime int64
	EndTime   int64
	Reason    string
)

var (
	AccessToken          string
	ExpiresIn            int
	AccessTokenCreatTime time.Time
)
