package routers

import (
	"fresh/Fresh-order/FreshOrder/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	//设置关卡
	beego.InsertFilter("/goods/*",beego.BeforeExec,filterFunc)
    beego.Router("/", &controllers.GoodesControllers{},"get:ShowIndex")
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HenderlRegister")
    //登录业务
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HanderlLogin")
    //激活用户
    beego.Router("/active",&controllers.UserController{},"get:ActvieUser")
    //退出登录
    beego.Router("/logout",&controllers.UserController{},"get:Logout")
    //用户中心业务
    beego.Router("/goods/usercenterinfo",&controllers.UserController{},"get:ShowUsercenterinfo")
    //个人中心内全部订单业务
    beego.Router("/goods/usercenterorder",&controllers.UserController{},"get:ShowUserCenterOrder")
	//个人中心内收货地址的业务
	beego.Router("/goods/usercentersite",&controllers.UserController{},"get:ShowUserCenterSite")
}

func filterFunc(ctx*context.Context)  {
	//获取session
	userName:=ctx.Input.Session("userName")
	if userName==nil {
		ctx.Redirect(302,"/login")
		return
	}

}
