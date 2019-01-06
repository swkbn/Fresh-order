package routers

import (
	"fresh/Fresh-order/FreshOrder/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.GoodesControllers{},"get:ShowIndex")
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HenderlRegister")
    //登录业务
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HanderlLogin")
    //激活用户
    beego.Router("/active",&controllers.UserController{},"get:ActvieUser")

}
