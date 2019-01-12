package controllers

import (
	"github.com/astaxie/beego"
	"fresh/Fresh-order/FreshOrder/models"
	"github.com/astaxie/beego/orm"
	"strconv"
	"github.com/gomodule/redigo/redis"
)

type OrderControllers struct {
	beego.Controller
}

func (this*OrderControllers)ShowOrder()  {
	//获取数据
	goodsIds:=this.GetStrings("goodsId")
	userName:=this.GetSession("userName").(string)
	//校验数据
	if len(userName)==0 {
		beego.Error("用户未登录")
		this.Redirect("/login",302)
		return
	}
	if len(goodsIds)==0 {
		beego.Error("获取数据失败")
		this.Redirect("/addCart",302)
		return
	}
	//处理数据
	o:=orm.NewOrm()
	//大容器
	var goods []map[string]interface{}
	//连接redis
	coon,err:=redis.Dial("tcp","192.168.189.11:6379")
	if err!=nil {
		beego.Error("reids连接失败")
		this.Redirect("/addCart",302)
		return
	}
	//总金额
	var mter int
	//商品数量
	sum:=0
	//商品信息
	for _,value:=range goodsIds{
		//定义一个容器
		temp:=make(map[string]interface{})
		var goodster  models.GoodsSKU
		Id,err:=strconv.Atoi(value)
		if err!=nil {
			beego.Error("转换数据失败")
			return
		}
		goodster.Id=Id
		//读取商品的基本信息
		o.Read(&goodster)
		//从redis中获取数据(获取到数量)
		pre,err:=coon.Do("hget","cart_"+userName,value)
		//回复助手函数
		count,err:=redis.Int(pre,err)
		if err!=nil {
			beego.Error("获取数据失败")
			return
		}
		//计算小计
		totPrice:=goodster.Price * count
		//把数量存入容器中
		temp["count"]=count
		//赋值给容器商品基本信息
		temp["place"]=goodster
		//把小计也放入容器中
		temp["totPrice"]=totPrice
		//赋值给大容器
		goods=append(goods,temp)
		//计算总金额
		mter+=totPrice
		//计算总商品数
		sum+=1
	}
	//地址信息
	var rece []models.Receiver
	//读取地址信息
	o.QueryTable("Receiver").RelatedSel("User").Filter("User__UserName",userName).All(&rece)
	//返回数据
	//地址信息
	this.Data["rece"]=rece
	//商品信息
	this.Data["goods"]=goods
	//用户名
	this.Data["userName"]=userName
	//总金额
	this.Data["mter"]=float32(mter)
	//实赴金额
	this.Data["lmter"]=mter+10
	//商品总数
	this.Data["sum"]=sum
	//指定页面
	this.TplName="place_order.html"
}
