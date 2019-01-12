package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/orm"
	"strconv"
	"fresh/Fresh-order/FreshOrder/models"
)

type CartControllers struct {
	beego.Controller
}
//添加购物车
func (this*CartControllers)HendelCart()  {
	//获取数据
	goodId,err1:=this.GetInt("goodId")
	count,err2:=this.GetInt("count")
	beego.Info(goodId,count)
	resp:=make(map[string]interface{})
	if err1!=nil||err2!=nil {
		beego.Error("获取数据失败",err2,err1)
		return
	}
	userName:=this.GetSession("userName")
	if userName==nil {
		resp["errno"]=1
		resp["errmsg"]="用户未登录"
		this.Data["json"]=resp
		//指定接受方式
		this.ServeJSON()
		return
	}
	//处理数据
	//使用redis存储数据
	conn,err:=redis.Dial("tcp","192.168.189.11:6379")
	if err!=nil {
		resp["errno"]=2
		resp["errmsg"]="redis连接失败"
		this.Data["json"]=resp
		//指定接受方式
		this.ServeJSON()
	}
	defer conn.Close()
	//往redis中存入数据
	conn.Do("hset","cart_"+userName.(string),goodId,count)
	//返回数据
	resp["errno"]=5
	resp["errmsg"]="OK"
	this.Data["json"]=resp
	this.ServeJSON()
}
//展示购物车页面

func (this*CartControllers)ShowCart()  {


	//获取数据
	userName:=this.GetSession("userName")
	if userName==nil {
		beego.Error("用户未登录")
		this.Redirect("/login",302)
		return
	}
	//连接redis
	conn,err:=redis.Dial("tcp","192.168.189.11:6379")
	if err!=nil {
		beego.Error("连接redis失败")
		this.Redirect("/",302)
		return
	}
	defer conn.Close()
	//从redis中读取数据
	res,err:=conn.Do("hgetall","cart_"+userName.(string))
	//回复助手函数
	kvsp,err:=redis.IntMap(res,err)
	if err!=nil {
		beego.Error("获取数据失败")
		this.Redirect("/",302)
		return 
	}
	//需要一个大容器
	var goods []map[string]interface{}

	o:=orm.NewOrm()
	//合计多少前
	var  Hj int
	var Gj  int

	//遍历获取到的map中的数据
	for key,value:=range kvsp{
		var goos models.GoodsSKU
		temp  :=make(map[string]interface{})
		goodsId,_:=strconv.Atoi(key)
		goos.Id=goodsId
		o.Read(&goos)
		temp["GoodsSku"]=goos
		temp["Count"]=value
		//小计
		temp["Xj"] = goos.Price*value
		//共计
		Gj+=1
		//合计
		Hj+=goos.Price*value
		//把这边的容器放入到大容器中
		goods=append(goods,temp)
	}
	//把数据传递给前端
	this.Data["goods"]=goods
	this.Data["Gj"]=Gj
	this.Data["Hj"]=Hj

	this.Data["userName"]=userName
	//指定页面
	this.TplName="cart.html"

}
//购物车内数量的处理
func (this*CartControllers)UpdateCart()  {
	//获取数据
	goodsId,err1:=this.GetInt("goodsId")
	count,err2:=this.GetInt("count")
	userName:=this.GetSession("userName")
	//定义一个返回json的容器
	resp:=make(map[string]interface{})
	//校验数据
	if err1!=nil||err2!=nil {
		resp["errson"]=1
		resp["errmsg"]="请求格式不正确"
		this.Data["json"]=resp
		this.ServeJSON()
		return
	}
	if userName==nil {
		resp["errson"]=2
		resp["errmsg"]="用户未登录"
		this.Data["json"]=resp
		this.ServeJSON()
		return
	}
	//处理数据
	//连接redis
	coon,err:=redis.Dial("tcp","192.168.189.11:6379")
	if err!=nil {
		resp["errson"]=3
		resp["errmsg"]="连接redis失败"
		this.Data["json"]=resp
		this.ServeJSON()
		return
	}
	defer coon.Close()
	//更新数据
	coon.Do("hset","cart_"+userName.(string),goodsId,count)
	//返回数据
	resp["errson"]=5
	resp["errmsg"]="OK"
	this.Data["json"]=resp
	this.ServeJSON()
	return

}

//处理错误函数
func errFunc(this*CartControllers,resp map[string]interface{} )  {
	this.Data["json"]=resp
	this.ServeJSON()
}

//购物车内商品的删除
func (this*CartControllers)DeleteCart()  {
	//获取数据
	goodsId,err:=this.GetInt("goodsId")
	resp:=make(map[string]interface{})
	defer errFunc(this,resp)
	if err!=nil {
		resp["errno"]=1
		resp["errmsg"]="获取数据失败"
		return
	}
	userName:=this.GetSession("userName")
	if userName==nil {
		resp["errno"]=2
		resp["errmsg"]="用户未登录"
		return
	}
	//连接redis
	coon,err:=redis.Dial("tcp","192.168.189.11:6379")

	if err!=nil {
		resp["errno"]=3
		resp["errmsg"]="连接redis失败"
		return
	}
	defer coon.Close()
	//删除redis中的数据
	coon.Do("hdel","cart_"+userName.(string),goodsId)
	resp["errno"]=5
	resp["errmsg"]="OK"
}
