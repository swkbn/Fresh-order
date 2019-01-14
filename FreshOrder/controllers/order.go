package controllers

import (
	"github.com/astaxie/beego"
	"fresh/Fresh-order/FreshOrder/models"
	"github.com/astaxie/beego/orm"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
	"fmt"
	"github.com/smartwalle/alipay"

	"github.com/KenmyZhang/aliyun-communicate"
	"math/rand"
)

type OrderControllers struct {
	beego.Controller
}
//展示订单页
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
	//商品的ID
	this.Data["goodsIds"]=goodsIds
	//指定页面
	this.TplName="place_order.html"
}
//添加订单数据   向订单表中添加数据    向订单商品表添加数据
func(this*OrderControllers)AddOrder(){
	//获取数据
	addrId,err1 :=this.GetInt("addrId")
	payId,err2:=this.GetInt("payId")
	goodsIds := this.GetString("goodsIds")
	totalCount,err5 := this.GetInt("totalCount")
	totaPrice,err3 := this.GetInt("totalPrice")
	transPrice,err4 := this.GetInt("transPrice")
	//totalPay,err6 :=this.GetInt("totalPay")

	resp := make(map[string]interface{})
	defer errFunc(&this.Controller,resp)
	beego.Info(addrId,payId,goodsIds,totalCount,totaPrice,transPrice)
	//校验数据
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整"
		return
	}

	//处理数据
	//向订单表插入数据
	o := orm.NewOrm()
	var orderInfo models.OrderInfo
	//获取收件人信息
	var receiver models.Receiver
	receiver.Id = addrId
	o.Read(&receiver)
	//把收件人信息赋值给订单
	orderInfo.Receiver = &receiver

	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totaPrice
	orderInfo.TransitPrice = transPrice
	//获取用户信息插入订单
	userName := this.GetSession("userName")
	if userName == nil{
		resp["errno"] = 2
		resp["errmsg"] = "用户未登录"
		return
	}
	var user models.User
	user.UserName = userName.(string)
	o.Read(&user,"UserName")
	//时间+用户Id
	orderInfo.OrderId = time.Now().Format("20060102150405") + strconv.Itoa(user.Id)

	orderInfo.User = &user
	o.Insert(&orderInfo)

	//向订单商品表插入数据

	conn,_:=redis.Dial("tcp","192.168.189.11:6379")
	defer conn.Close()

	//对这个数据处理"[1 13 10]"
	ids:=strings.Split(goodsIds[1:len(goodsIds)-1]," ")
	//beego.Info("ids=",ids)
	for _,v := range ids{
		var orderGoods models.OrderGoods
		id,_:=strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)
		//获取商品数量
		count,_:=redis.Int(conn.Do("hget","cart_"+userName.(string),id))

		//给orderGoods赋值
		orderGoods.GoodsSKU = &goodsSku
		orderGoods.Count = count
		orderGoods.Price = count * goodsSku.Price
		orderGoods.OrderInfo = &orderInfo
		//插入操作
		o.Insert(&orderGoods)
	}

	//返回数据
	resp["errno"] = 5
	resp["errmsg"] = "OK"
}

//支付宝付款
func (this*OrderControllers)HendelPay()  {
	var privateKey = `MIIEpAIBAAKCAQEAyaRuIJ5R3h+DmfSuAMcJFENu0/u3wHmDDo7PSC/9mM/vGOfl
2WPjE8iliBZmBfBxRFPr3/w80H2FExpzOCe5WLouz83FxpDUYrBfp6xHtzUG092e
hJKUlnLZa0Bs4qveNaZ0eg9FYimablG3k060cbk1SHgbDVWWpcI3sjcQ19R31wDg
qn2SrwI1CJiz8IwwXVemTieOXdqTaFSN7AQa30MyYFSqqRW24gdtSstC0KnGcj5J
6XJNa4eUKo0LGDzU2bSH2Pr1qn2Qpbzk0AFvmAHmZLH5Ry48M2IVoudfMK13UOpm
BHTPmulixP+lh0y9eJgo4KWOdUelooni4/uDvQIDAQABAoIBAATOUHmijFz471AK
DuOh2suK1+dhhn2l58O/D52u1yJ/QjmbvVSzFsRv8dIOhpv5oRl5zpNmFaT6eON9
q+VYvQgqV9dIFkCnTwiTH5SFfKgXMXR3QcHzJGt5jUkLHg1A/2jT8M6/8m1mhHHA
rNlr9M0JFwYFJs/ojFCjEmTC/znFocjAXi9dwvfZgyJkg92peXlYSYIoudIM5bw8
ZES7RfjXtqDaZNIbYL2bsEz9iEJx17Acoo3GXeJ8WUmpUNhtELDCU1Tf8u584Yd2
WYmcGNJZ6LlgAdW3nIR11CQCfAyyTvGadjjF6IW7venzLugxXxiOHW9zOenje8Wy
0zwBrdkCgYEA5ldLFl1ZkzuvoyQQdQAMmCSou9dTJyNy0eMC0954JqTqI/BqKkiQ
HldIJ8QL805aVF+OWrniffo3YIylypzXVjPVvLStQHAgLJx3WnB1l4FXewaqCbv3
ZlZtRtPjM4ZzQs8KBUU+yxBWcuLhwfTEk1Hu6QgQbO8/sUnFpJI8m1sCgYEA4Bq6
Ocqw4/ryXh5HoT7ydf9cLJqFFSeamQR6ngHST/0aiUN/Xoh6o9pUGdHWYOKpkJCA
87Ulqe944lMRx+ZE3GxXZ0xpRgmOOlntM9g7I2fWKpVOifnEulFDYLMNVr0uItJH
/JH0k4Olrbii1ETNHbUMkd0QJ5GaM75IMBaYQMcCgYArQGT3FBxHy0NVrOXyMkor
H2cXrn0MsllTE/9p7TI+f2T/zpsAyZNWPylrXiKoUyQfB7phSto+sYdId+CBxSWi
KCWQQ5TsrqE7/z1iHA/YnQ7iKQQww7zW2I+4Zv0Ypbxq5RmKl9AMrUquU+/0TZPD
3fSwiTUcX2hkT+fu2Q7MVQKBgQCxaTdIDQggU0eP7tSx+A0mELQ9s03rw2CGBp+z
eqmuHSbmx4KLqeu8z1iI4C+gn4+xHFSZmixo7WV7dlu7LrYQ8cv3wOwOP/5Sf2Jj
Cqk2jDtllrGIVSzCexal9Nl4c2eUtXe7oShHp45/io2NEbJ39B4xUxo42PGESP0I
5Lo/fQKBgQCsLDbqBqegJIySYIruuLeniD3UvOt2qJvnp/h71JGSV7eM2q8N2+pg
cO9hcnwKu1gydQhxI5P32Td+nkLYNtikQMh/H6uJe5DWlKhEblwAZfMJUMhfFs3z
oBMRXisszSg7vKIs4pD3XO8h97k/rVRRpN3nn3aikIUbcQfDYNEgTA==
`

	var appId = "2016092400588580"
	var aliPublicKey = `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyaRuIJ5R3h+DmfSuAMcJ FENu0/u3wHmDDo7PSC/9mM/vGOfl2WPjE8iliBZmBfBxRFPr3/w80H2FExpzOCe5 WLouz83FxpDUYrBfp6xHtzUG092ehJKUlnLZa0Bs4qveNaZ0eg9FYimablG3k060 cbk1SHgbDVWWpcI3sjcQ19R31wDgqn2SrwI1CJiz8IwwXVemTieOXdqTaFSN7AQa 30MyYFSqqRW24gdtSstC0KnGcj5J6XJNa4eUKo0LGDzU2bSH2Pr1qn2Qpbzk0AFv mAHmZLH5Ry48M2IVoudfMK13UOpmBHTPmulixP+lh0y9eJgo4KWOdUelooni4/uD vQIDAQAB`

	//业务代码
	var client = alipay.New(appId, aliPublicKey, privateKey, false)
	//获取订单信息
	//userName:=this.GetSession("userName")
	ids,_:=this.GetInt("ids")
	beego.Info(ids)

	rand.Seed(time.Now().UnixNano())
	code2:=rand.Intn(10000)
	code:=strconv.Itoa(code2)

	//alipay.trade.page.pay
	//回调支付宝支付界面配置
	var p = alipay.AliPayTradePagePay{}
	p.NotifyURL = "http://192.168.189.11:8080/payOk?ids="+strconv.Itoa(ids)
	p.ReturnURL = "http://192.168.189.11:8080/payOk?ids="+strconv.Itoa(ids)
	p.Subject = "天天生鲜"
	p.OutTradeNo = code
	p.TotalAmount = "3000.00"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	//调用支付宝支付界面
	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}

	var payURL = url.String()

	this.Redirect(payURL,302)



}
//支付成功
func (this*OrderControllers)PayOK()  {
	trade_no:=this.GetString("trade_no")
	if trade_no!="" {
		ids,_:=this.GetInt("ids")
		o:=orm.NewOrm()
		var oderinfo models.OrderInfo
		oderinfo.Id=ids
		o.Read(&oderinfo)
		oderinfo.Orderstatus=1
		o.Update(&oderinfo)
	}


	this.Redirect("/goods/usercenterorder",302)
}
//SDK接入,短信发送

func(this*OrderControllers)SendMsg(){

	rand.Seed(time.Now().UnixNano())
	code2:=rand.Intn(10000)
	code:=strconv.Itoa(code2)
	var (
		gatewayUrl      = "http://dysmsapi.aliyuncs.com/"
		accessKeyId     = "LTAIN9gZtWEmkc1e"
		accessKeySecret = "H7wFlnWODmifC7DHgps21wfO5GRn1e"
		phoneNumbers    = "18851509790"  //要发送的电话号码
		signName        = "天天生鲜"     //签名名称
		templateCode    = "SMS_149101793"  //模板号
		templateParam   = "{\"code\":\""+code+"\"}"//验证码
	)

	smsClient := aliyunsmsclient.New(gatewayUrl)
	result, err := smsClient.Execute(accessKeyId, accessKeySecret, phoneNumbers, signName, templateCode, templateParam)
	//fmt.Println("Got raw response from server:", string(result.RawResponse))
	if err != nil {
		beego.Info("配置有问题")
	}

	if result.IsSuccessful() {
		beego.Error("短信成功")
	} else {
		beego.Error("短信失败")
	}
	//this.TplName = "SMS.html"
	this.Redirect("/",302)
}