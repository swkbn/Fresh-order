package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"fresh/Fresh-order/FreshOrder/models"
	"github.com/gomodule/redigo/redis"
	"math"
)

type GoodesControllers struct {
	beego.Controller
}
//首页展示业务
func (this*GoodesControllers)ShowIndex()  {
	//从session中获取用户名
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
	//获取所有商品类型
	o.QueryTable("GoodsType").All(&goodesTypes)
	//把数据传递给前端
	this.Data["goodesTypes"]=goodesTypes

	//论波图展示
	var goodsBanner []models.IndexGoodsBanner
	//根据表名查询表内所有的信息
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanner)
	//把论波图数据传递给前端
	this.Data["goodsBanner"]=goodsBanner

	//促销商品展示
	var goodsPromotion [] models.IndexPromotionBanner
	//从促销商品表中根据给的index进行排序获取所有
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&goodsPromotion)
	//把促销产品传递给前端
	this.Data["goodsPromotion"]=goodsPromotion
	//商品详情展示
	//创建一个存商品的容器（这个容器是map类型的切片，map的氏的数据类型是interface）
	var goods []map[string]interface{}
	//把所有的商品类型插入到容器中
	//遍历商品类型
	for _,v:=range goodesTypes{
		temp:=make(map[string]interface{})
		//把商品类型存入到一个类型：map中
		temp["goodsType"]=v
		//把类型添加到容器中
		goods=append(goods,temp)
	}
	//遍历一个有类型的大容器
	for _,v:=range goods{
		//获取到类型对应的商品
		qs:=o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType",v["goodsType"])
		var goodsText[] models.IndexTypeGoodsBanner
		//获取文字商品（筛选获取需要的商品）
		qs.Filter("DisplayType",0).OrderBy("Index").All(&goodsText)
		//获取图片商品
		var goodsImage[] models.IndexTypeGoodsBanner
		qs.Filter("DisplayType",1).OrderBy("Index").All(&goodsImage)
		//插入容器中
		v["goodsText"]=goodsText
		v["goodsImage"]=goodsImage
	}
	//打印是否获取到数据
	//beego.Info(goods)
	//传递给前端
	this.Data["goods"]=goods

	//显示页面
	this.TplName="index.html"
}

//商品详情页面

func (this*GoodesControllers)ShowDetail()  {



	//获取数据
	//获取商品的id
	goodId,err:=this.GetInt("goodsId")
	if err!=nil {
		beego.Error("获取数据失败")
		this.Redirect("/",302)
		return
	}
	var goods models.GoodsSKU
	goods.Id=goodId
	o:=orm.NewOrm()
	//获取商品的详情数据，关连goods表和goodstype表是为了两个表中的数据，根据id筛选选择的是哪一个
	//one是只有一个商品详情
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
	//连接商品类型表，是同类型的新品，根据时间排序，获取两个从0开始
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType",goods.GoodsType).OrderBy("Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"]=newGoods

//添加历史浏览记录
	userName:=this.GetSession("userName")
	if userName!=nil {
		//用redis中的list存储
		//连接redis
		coon,err:=redis.Dial("tcp","192.168.189.11:6379")
		if err!=nil {
			beego.Error("连接redis失败")
			return
		}
		defer coon.Close()
		//插入之前需要把重复的删掉（如果没有这个数据是删不掉的，所以是先删除后添加）
		coon.Do("lrem","history_"+userName.(string),0,goodId)
		//插入数据
		coon.Do("lpush","history_"+userName.(string),goodId)
	}

	if userName==nil {
		this.Data["userName"]=""
	}else {
		//userName是接口类型 需要断言
		this.Data["userName"]=userName.(string)
	}

	this.TplName="detail.html"

}

//类型详情页面
func (this*GoodesControllers)ShowList()  {
	//从session中获取用户名
	userName:=this.GetSession("userName")

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
	//是为了查看对应类型id内的所有数据
	qs:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId)
	//获取所有

	//返回数据

	//分页处理
	//获取总数据数
	count,err:=qs.Count()
	if err!=nil {
		beego.Error("获取数量失败")
		return
	}
	//计算总页数
	pageSize:=1
	pageCount:=math.Ceil(float64(count)/float64(pageSize))
	//获取当前页码
	pageIndex,err:=this.GetInt("pageIndex")
	if err!=nil {
		pageIndex=1
	}
	pages:=pageEditor(int(pageCount),pageIndex)
	this.Data["pages"]=pages
	//计算出开始的下表位置


	start := (pageIndex -1 )*pageSize
	//排序
	sort := this.GetString("sort")
	if sort == ""{
		//默认排序
		qs.Limit(pageSize,start).All(&goods)
		this.Data["sort"]=""
	}else if sort == "price"{
		//价格排序
		qs.OrderBy("-Price").Limit(pageSize,start).All(&goods)
		this.Data["sort"]= "price"
	}else{
		//销量排序
		qs.OrderBy("Sales").Limit(pageSize,start).All(&goods)
		this.Data["sort"] = "sale"
	}



	//实现页码显示  上一页下一页页码处理
	var preIndex,nextIndex int
	//上一页实现
	if pageIndex == 1{
		preIndex = 1
	}else {
		preIndex = pageIndex - 1

	}
	//下一页实现
	if pageIndex == int(pageCount) {
		nextIndex = int(pageCount)
	}else {
		nextIndex = pageIndex + 1

	}
	//新品推荐
	var newGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).OrderBy("Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"]=newGoods
	//商品类型展示
	var goodesTypes [] models.GoodsType
	//获取所有商品类型
	o.QueryTable("GoodsType").All(&goodesTypes)
	//把数据传递给前端
	this.Data["goodesTypes"]=goodesTypes
//////////////////////////////////
	this.Data["pageIndex"]=pageIndex
	this.Data["goods"]=goods
	this.Data["preIndex"] = preIndex
	this.Data["nextIndex"] = nextIndex
	//传递给前端
	//获取当前类型传递给前端
  var goodsTypeser models.GoodsType
  	goodsTypeser.Id=typeId					//给前端传递一个商品类型
  	o.Read(&goodsTypeser)
	this.Data["GoodsType"]=goodsTypeser
	this.Data["typeId"]=typeId

	if userName==nil {
		this.Data["userName"]=""
	}else {
		//userName是接口类型 需要断言
		this.Data["userName"]=userName.(string)
	}
	//显示页面
	this.TplName="list.html"
}

//封装一个分页处理函数

func pageEditor(pageCount int,pageIndex int) []int {
	//存放页码的切片
	var pages []int
	//如果总页数小于5
	if pageCount<5 {
		pages=make([]int,pageCount)
		//就存和总页数相同的页数
		for i:=1;i<=pageCount;i++ {
			pages[i-1]=i
		}
		//如果当前页面小于等于3时
	}else if pageIndex<=3{
		pages =make([]int,5)
		//就显示前5页，把前5页的页码存放进切片中
		for i:=1;i<=5 ;i++  {
			pages[i-1]=i
		}
		//如果当前页书大于总页书减2为了显示最后5个
	}else if pageIndex>=pageCount-2 {
		pages=make([]int,5)
		//把最后5个存入到切片中
		for i:=1;i<=5 ;i++  {
			pages[i-1]=pageCount-5+i
		}
	}else {
		//这个是为了显示让当前页在中间
		pages=make([]int,5)
		for i:=1;i<=5;i++ {
			pages[i-1]=pageIndex-3+i
		}
	}
	return pages

}

//商品搜索
func (this*GoodesControllers)HanderlSearch()  {
	//从session中获取用户名
	userName:=this.GetSession("userName")
	if userName==nil {
		this.Data["userName"]=""
	}else {
		//userName是接口类型 需要断言
		this.Data["userName"]=userName.(string)
	}
	//获取数据
	searchName:=this.GetString("searchName")

	if searchName=="" {
		this.Redirect("/",302)
	}

	var goods []models.GoodsSKU

	o:=orm.NewOrm()
	o.QueryTable("GoodsSKU").Filter("Name__contains",searchName).All(&goods)
	//返回数据给前端
	this.Data["goods"]=goods
	//商品类型展示
	var goodesTypes [] models.GoodsType
	//获取所有商品类型
	o.QueryTable("GoodsType").All(&goodesTypes)
	//把数据传递给前端
	this.Data["goodesTypes"]=goodesTypes
	//指定页面
	this.TplName="search.html"

}
