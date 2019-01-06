package routers

import (
	"fresh/Fresh-order/FreshOrder/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HenderlRegister")
    //登录业务
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin")

}
