package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"fresh/Fresh-order/FreshOrder/models"

	"github.com/gomodule/redigo/redis"
)

type GoodesControllers struct {
	beego.Controller
}
//首页展示业务
func (this*GoodesControllers)ShowIndex()  {

	userName:=this.GetSession("userName")
	if userName==nil {
		this.Data["userName"]=""
	}else {
		//userName是接口类型 需要断言
		this.Data["userName"]=userName.(string)
	}
	//商品分类展示
	//获取数据库内信息
	o:=orm.NewOrm()
	var goodesTypes [] models.GoodsType
	o.QueryTable("GoodsType").All(&goodesTypes)
	this.Data["goodesTypes"]=goodesTypes

	//论波图展示

	var goodsBanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanner)
	this.Data["goodsBanner"]=goodsBanner

	//促销商品展示
	var goodsPromotion [] models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&goodsPromotion)
	this.Data["goodsPromotion"]=goodsPromotion
	//商品详情展示
	//创建一个存商品的容器
	var goods []map[string]interface{}
	//把所有的商品类型插入到容器中
	for _,v:=range goodesTypes{
		temp:=make(map[string]interface{})
		temp["goodsType"]=v
		//把类型添加到容器中
		goods=append(goods,temp)
	}
	for _,v:=range goods{
		//获取到类型对应的商品
		qs:=o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType",v["goodsType"])
		var goodsText[] models.IndexTypeGoodsBanner
		//获取文字商品
		qs.Filter("DisplayType",0).OrderBy("Index").All(&goodsText)
		//获取图片商品
		var goodsImage[] models.IndexTypeGoodsBanner
		qs.Filter("DisplayType",1).OrderBy("Index").All(&goodsImage)
		//插入容器中
		v["goodsText"]=goodsText
		v["goodsImage"]=goodsImage
	}
	//打印是否获取到数据
	beego.Info(goods)
	//传递给前端
	this.Data["goods"]=goods

	//显示页面
	this.TplName="index.html"
}

//商品详情页面

func (this*GoodesControllers)ShowDetail()  {
	//获取数据
	goodId,err:=this.GetInt("goodsId")
	if err!=nil {
		beego.Error("获取数据失败")
		this.Redirect("/",302)
		return
	}
	var goods models.GoodsSKU
	goods.Id=goodId
	o:=orm.NewOrm()
	o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",goodId).One(&goods)
	//o.Read(&goods)
	//
	this.Data["goods"]=goods
	//获取类型数据
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"]=goodsTypes
	//获取新品数据
	var newGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType",goods.GoodsType).OrderBy("Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"]=newGoods

//添加历史浏览记录
	userName:=this.GetSession("userName")
	if userName!=nil {
		//用redis中的list存储
		coon,err:=redis.Dial("tcp","192.168.189.11:6379")
		if err!=nil {
			beego.Error("连接redis失败")
			return
		}
		defer coon.Close()
		//插入之前需要把重复的删掉
		coon.Do("lrem","history_"+userName.(string),0,goodId)
		//插入数据
		coon.Do("lpush","history_"+userName.(string),goodId)
	}


	this.TplName="detail.html"
}

//类型详情页面

func (this*GoodesControllers)ShowList()  {
	//获取数据
	typeId,err:=this.GetInt("typeId")
	//校验数据
	if err!=nil {
		beego.Error("获取数据失败")
		return
	}
	//处理数据
	o:=orm.NewOrm()
	var goods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).All(&goods)
	//返回数据
	this.Data["goods"]=goods
	//显示页面
	this.TplName="list.html"
}

