package controllers

import "github.com/astaxie/beego"

type GoodesControllers struct {
	beego.Controller
}

func (this*GoodesControllers)ShowIndex()  {

	userName:=this.GetSession("userName")
	if userName==nil {
		this.Data["userName"]=""
	}else {
		//userName是接口类型 需要断言
		this.Data["userName"]=userName.(string)
	}
	this.TplName="index.html"
}
