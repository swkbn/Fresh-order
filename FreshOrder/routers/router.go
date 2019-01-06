package routers

import (
	"fresh/Fresh-order/FreshOrder/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
