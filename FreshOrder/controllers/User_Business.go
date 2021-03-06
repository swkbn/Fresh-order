package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"fresh/Fresh-order/FreshOrder/models"
	"regexp"
	"strconv"

	"github.com/astaxie/beego/utils"
	"github.com/gomodule/redigo/redis"
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
	//this.Redirect("/login",302)
	//注册成功发送激活码
	//注册成功的时候发送激活邮件
	config := `{"username":"601232044@qq.com","password":"jxtcakchoapvbfii","host":"smtp.qq.com","port":587}`
	emailSend := utils.NewEMail(config)
	emailSend.From = "601232044@qq.com"
	emailSend.To = []string{email}
	emailSend.Subject = "天天生鲜用户激活"
	emailSend.HTML = `<a href="http://192.168.189.11:8080/active?userId=`+strconv.Itoa(User.Id)+`">点击激活</a>`
	emailSend.Send()
	//注册成功页面业务
	this.Ctx.WriteString("注册成功请进入邮箱进行激活")
}
//激活操作
func (this*UserController)ActvieUser()  {
	userId,err:=this.GetInt("userId")
	if err!=nil {
		beego.Error("获取数据失败")
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Id=userId
	err=o.Read(&user)
	if err!=nil {
		beego.Error("用户名不存在")
		this.Redirect("/register",302)
		return
	}

	user.Active=1
	_,err=o.Update(&user)
	if err!=nil {
		beego.Error("更新数据失败")
		this.Redirect("/register",302)
	}
	//激活成功后进入登录页面
	this.Redirect("/login",302)

}
//显示登录页面
func (this *UserController)ShowLogin()  {

	UserName:=this.Ctx.GetCookie("UserName")
	if UserName!="" {
		this.Data["userName"]=UserName
		this.Data["checked"]="checked"
	}else {
		this.Data["userName"]=""
		this.Data["checked"]=""
	}
	this.TplName="login.html"
}
//登录业务处理
func (this*UserController)HanderlLogin()  {
	//获取数据
	username:=this.GetString("username")
	pwd:=this.GetString("pwd")
	if username==""||pwd=="" {
		beego.Error("用户名和密码输入不完整,请重新输入")
		this.Redirect("/login",302)
		return
	}
	//处理数据
	o:=orm.NewOrm()
	var users models.User
	users.UserName=username
	//查询
	err:=o.Read(&users,"UserName")
	if err!=nil {
		beego.Error("用户名不存在,请重新输入")
		this.Redirect("/login",302)
		return
	}
	if users.Pwd!=pwd {
		beego.Error("密码不正确,请重新输入")
		this.Redirect("/login",302)
		return
	}
	if users.Active==0 {
		beego.Error("没有激活,请进入注册时的邮箱进行激活")
		this.Redirect("/login",302)
		return
	}



	//记住用户名操作

	rember:=this.GetString("rember")
	if rember=="on" {
		this.Ctx.SetCookie("UserName",users.UserName,3600)
	}else {
		this.Ctx.SetCookie("UserName",users.UserName,-1)
	}
	//设置Session
	this.SetSession("userName",username)
	//跳转到首页
	this.Redirect("/",302)

}
//退出登录实现
func (this*UserController)Logout()  {

	this.DelSession("userName")
	this.Redirect("/",302)

}
//显示用户中心
func (this*UserController)ShowUsercenterinfo()  {
	//从session中获取用户名
	userName:=this.GetSession("userName")
	this.Data["userName"]=userName
	//获取
	var siet models.Receiver
	o:=orm.NewOrm()
	o.QueryTable("Receiver").RelatedSel("User").Filter("User__UserName",userName).Filter("IsDefault",true).One(&siet)
	this.Data["siet"]=siet
	//显示浏览记录
	//从redis中获取数据
	coon,err:=redis.Dial("tcp","192.168.189.11:6379")
	if err!=nil {
		beego.Error("连接redis错误")
		return
	}
	defer coon.Close()
	resp,err:=coon.Do("lrange","history_"+userName.(string),0,4)
	 //回复助手汉书，对类型进行转换
	 res,err:=redis.Ints(resp,err)
	if err!=nil {
		beego.Error("有问题")
		return
	}
	var goods []models.GoodsSKU
	for _,goodsId :=range res{
		var goodsSku models.GoodsSKU
		goodsSku.Id=goodsId
		o.Read(&goodsSku)
		goods=append(goods,goodsSku)
	}
	beego.Info(goods)
	this.Data["goods"]=goods

	//拼接显示页面
	this.Layout="layout.html"
	this.TplName="user_center_info.html"

}
//显示用户中心页面的订单页
func (this*UserController)ShowUserCenterOrder()  {

	//从session中获取用户名
	userName:=this.GetSession("userName").(string)
	this.Data["userName"]=userName

	var goods []map[string]interface{}
	//获取订单数据
	o := orm.NewOrm()
	//获取当前用户的订单数据
	var orderInfos []models.OrderInfo

	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__UserName",userName).All(&orderInfos)

	//获取订单商品数据
	for _,v := range orderInfos{
		temp := make(map[string]interface{})
		//查询所有订单商品
		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo__Id",v.Id).All(&orderGoods)

		//把数据塞进容器
		temp["orderInfo"] = v
		temp["orderGoods"] = orderGoods

		goods = append(goods,temp)
	}

	this.Data["goods"] = goods




	//用于界面拼接
	this.Layout="layout.html"
	this.TplName="user_center_order.html"
}
//显示用户中心内收货地址
func (this*UserController)ShowUserCenterSite()  {
	//从session中获取用户名
	userName:=this.GetSession("userName")
	this.Data["userName"]=userName
	//显示默认地址
	var DefaultSite models.Receiver

	o:=orm.NewOrm()
	o.QueryTable("Receiver").RelatedSel("User").Filter("User__UserName",userName).Filter("IsDefault",true).One(&DefaultSite)
	//把数据传输给前端
	this.Data["DefaultSite"]=DefaultSite



	//用于界面拼接
	this.Layout="layout.html"
	this.TplName="user_center_site.html"



}
//提交收货地址
func (this*UserController)HenderlUserCenterSite()  {
	//获取用户名
	userName:=this.GetSession("userName")
	//获取数据
	recipientName:=this.GetString("recipientName")		//收件人
	detailedAddress:=this.GetString("detailedAddress")		//详细地址
	zipCode:=this.GetString("zipCode")		//邮编
	mobilePhone:=this.GetString("mobilePhone")		//手机号
	//校验数据
	if userName==nil {
		beego.Error("获取用户名失败")
		return
	}
	if recipientName==""||detailedAddress==""||zipCode==""||mobilePhone=="" {
		beego.Error("地址信息输入不完整,请重新输入")
		this.TplName="user_center_site.html"
		return
	}
	//处理数据
	o:=orm.NewOrm()
	//用户对象
	var user models.User
	user.UserName=userName.(string)
	err:=o.Read(&user,"UserName")
	if err!=nil {
		beego.Error("读取数据失败")
		return
	}

	//判断是否有默认地址，如果有默认地址就把默认地址更改为非默认地址
	var DefaultSite  models.Receiver
	//查询当前用户的所有收件人地址
	qs:=o.QueryTable("Receiver").RelatedSel("User").Filter("User__Id",user.Id)
	//查询是否有默认地址
	err=qs.Filter("IsDefault",true).One(&DefaultSite)
	if err==nil {
		//把默认地址更改为非默认地址
		DefaultSite.IsDefault=false
		o.Update(&DefaultSite)
	}

	//地址对象
	var site models.Receiver
	//赋值操作
	site.Name=recipientName		//收件人
	site.ZipCode=detailedAddress	//地址
	site.Addre=zipCode				//邮编
	site.Phone=mobilePhone			//手机号
	site.IsDefault=true				//把新添加的地址更改为默认地址

	site.User=&user
	_,err=o.Insert(&site)
	if err!=nil {
		beego.Error("插入数据失败")
		return
	}
	this.Redirect("/goods/usercenterinfo",302)

}





