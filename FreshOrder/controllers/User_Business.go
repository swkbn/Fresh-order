package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"fresh/Fresh-order/FreshOrder/models"
	"regexp"
)

type UserController struct {
	beego.Controller
}
//显示注册页面
func (this *UserController)ShowRegister()  {
	this.TplName="register.html"

}
//注册业务实现
func (this *UserController)HenderlRegister()  {
	//获取数据
	user_name:=this.GetString("user_name")
	pwd:=this.GetString("pwd")
	cpwd:=this.GetString("cpwd")
	email:=this.GetString("email")
	//注册orm对象
	o:=orm.NewOrm()
	//注册操作对象
	var User models.User
	//校验数据
	if user_name==""||pwd==""||cpwd==""||email=="" {
		this.Data["errmasg"]="注册信息不能为空，请重新输入"
		this.TplName="register.html"
		return
	}
	//校验两次密码是否一致
	if pwd!=cpwd {
		this.Data["errmasg"]="两次密码不一致请重新输入"
		beego.Error("邮箱格式不正确")
		this.TplName="register.html"
		return
	}
	//正则表达式校验邮箱格式
	reg,err:=regexp.Compile(`^[A-Za-z\d]+([-_.][A-Za-z\d]+)*@([A-Za-z\d]+[-.])+[A-Za-z\d]{2,4}$`)
	if err!=nil {
		this.Data["errmasg"]="正则匹配错误"
		this.TplName="register.html"
		beego.Error("正则表达式错误")
		return
	}
	re:=reg.MatchString(email)
	if re==false {
		this.Data["errmasg"]="邮箱格式不正确，请重新输入"
		this.TplName="register.html"
		beego.Error("邮箱格式不正确")
		return
	}
	//赋值
	User.UserName=user_name
	User.Pwd=pwd
	User.Email=email
	_,err=o.Insert(&User)
	if err!=nil {
		this.Data["errmasg"]="用户名已存在，请重新输入"
		this.TplName="register.html"
		beego.Error("插入数据失败")
		return
	}
	this.Redirect("/login",302)
}
//显示登录页面
func (this *UserController)ShowLogin()  {
	this.TplName="login.html"
}

