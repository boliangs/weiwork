package router

import (
	"github.com/gin-gonic/gin"
	"sendswork/app/gateway/internal"
	"sendswork/app/gateway/middlewares"
)

func Router() *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.Cors())
	ping := r.Group("/ping")
	registerPing(ping)

	user := r.Group("/user")
	registerUser(user)

	school := r.Group("/school")
	school.Use(middlewares.LimiterHandler(20)).
		Use(middlewares.AuthUserCheck())
	registerSchool(school)

	yearBill := r.Group("/yearBill", middlewares.AuthUserCheck())
	registerYearBill(yearBill)
	r.GET("/yearBill/init", middlewares.WebsocketMiddleware(), internal.YearBillDataInit)

	return r
}

func registerPing(group *gin.RouterGroup) {
	group.GET("/user", internal.UserPing)
	group.GET("/boBing", internal.BoBingPing)
	group.GET("/yearBill", internal.YearBillPing)
}

func registerUser(group *gin.RouterGroup) {
	group.POST("/login", internal.UserLogin)
	group.POST("/jssdk", internal.WxJSSDK)
	group.POST("/school_login", internal.SchoolUserLogin)
	group.POST("/bill_login", internal.YearBillLogin)
	group.POST("/check_token", internal.CheckTokenExpire)
}

func registerSchool(group *gin.RouterGroup) {
	group.POST("/schedule", internal.SchoolSchedule)
	group.GET("/xuefen", internal.SchoolXueFen)
	group.GET("/gpa", internal.SchoolGpa)
	group.GET("/grade", internal.SchoolGrade)
}

func registerYearBill(group *gin.RouterGroup) {
	group.GET("/learn", internal.GetLearnData)
	group.GET("/pay", internal.GetPayData)
	group.GET("/rank", internal.GetRank)
	group.POST("/appraise", internal.Appraise)
}
