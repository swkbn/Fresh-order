package models

import ("github.com/astaxie/beego/orm"

		_"github.com/go-sql-driver/mysql"

)

type User struct {
	Id  int	`orm:"unique"`
	UserName string		//用户名
	Pwd string			//密码
	Email string		//邮箱
	Power int	`orm:"default(0)"`		//权限
	Active int	`orm:"default(0)"`		//激活状态
	Receiver []*Receiver `orm:"reverse(many)"`
}

type Receiver struct {
	Id int
	Name string			//收件人名字
	ZipCode string		//地址
	Addre string		//邮编
	Phone string		//手机号码
	IsDefault bool `orm:"default(false)"`
	//关联
	User * User `orm:"rel(fk)"`
}

func init()  {
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/dailyfresh?charset=utf8")
	orm.RegisterModel(new(User),new(Receiver))

	orm.RunSyncdb("default",false,true)

}
