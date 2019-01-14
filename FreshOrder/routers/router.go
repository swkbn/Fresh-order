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
	beego.Router("/goods/usercentersite",&controllers.UserController{},"get:ShowUserCenterSite;post:HenderlUserCenterSite")
	//商品详情
	beego.Router("/detail",&controllers.GoodesControllers{},"get:ShowDetail")
	//商品分类页面
	beego.Router("/list",&controllers.GoodesControllers{},"get:ShowList")
	//搜索
	beego.Router("/search",&controllers.GoodesControllers{},"post:HanderlSearch")


	//加入购物车
	beego.Router("/addCart",&controllers.CartControllers{},"post:HendelCart;get:ShowCart")
	//购物车内数据改变
	beego.Router("/updateCart",&controllers.CartControllers{},"post:UpdateCart")
	//删除购物车内商品
	beego.Router("/deleteCart",&controllers.CartControllers{},"post:DeleteCart")
	//结算订单
	beego.Router("/goods/order",&controllers.OrderControllers{},"post:ShowOrder")
	//提交订单业务
	beego.Router("/addOrder",&controllers.OrderControllers{},"post:AddOrder")
	//去支付
	beego.Router("/aliPay",&controllers.OrderControllers{},"get:HendelPay")
	//支付成功后跳转
	beego.Router("/payOk",&controllers.OrderControllers{},"get:PayOK")
	//发送短信
	beego.Router("/payDX",&controllers.OrderControllers{},"get:SendMsg")

}

func filterFunc(ctx*context.Context)  {
	//获取session
	userName:=ctx.Input.Session("userName")
	if userName==nil {
		ctx.Redirect(302,"/login")
		return
	}

}
