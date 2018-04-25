package main

import (
	"net/http"
	"log"
	"time"
	"github.com/jinzhu/gorm"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	_"github.com/go-sql-driver/mysql"
	"strconv"
)

// 定义数据库连接实例db对象
var db *gorm.DB

type Pack struct {
	Id string
	Userphone string `gorm:"not null"`
	Username string `gorm:"not null"`
	State  int `gorm:"default:0"`
	Havedial int `gorm:"default:0"`
	Havemess int `gorm:"default:0"`
	Intime time.Time
	Outtime time.Time
}
// 修改默认表名
//func (Pack) TableName() string {
//	return "tb_pack"
//}

// 获取总数及分页记录
func getList(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	r.ParseForm()
	uphone := r.FormValue("uphone")
	page := r.FormValue("page")
	size := r.FormValue("size")

	// 定义返回的数据结构
	result := make(map[string]interface{})

	// 符合查询条件的总数
	if len(page) == 0 && len(size) == 0 {
		var count int

		rs := db.Model(&Pack{}).Where("1=1").Count(&count)
		if len(uphone) != 0 {
			rs = rs.Where("userphone like ?", "%"+uphone+"%").Count(&count)
		}

		result["count"] = count		// 符合条件的记录总数
	}else {
		// 页码和每页个数
		page_int, _ := strconv.Atoi(page)
		size_int, _ := strconv.Atoi(size)

		// 定义pack切片，存储数据集
		var packs []*Pack
		//err := db.Find(&packs)			// 查询全部

		if len(uphone) != 0 {
			db.Where("userphone like ?", "%"+uphone+"%").Offset((page_int - 1) * size_int).Limit(size_int).Find(&packs)
		} else {
			db.Offset((page_int - 1) * size_int).Limit(size_int).Find(&packs)
		}
		result["dt"] = packs // 符合提交的分页列表
	}

	result["success"] = true

	// 返回json结果
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 根据id查询
func getPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	id := ps.ByName("id")
	var pack Pack
	err := db.First(&pack,id)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if err != nil {
		result["pack"] = pack
		result["success"] = true
	}else{
		result["success"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 新增记录
func addPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()
	id := r.FormValue("id")
	uphone := r.FormValue("uphone")
	uname := r.FormValue("uname")

	pack := Pack{
		Id:id,
		Userphone:uphone,
		Username:uname,
		Intime:time.Now(),
		Outtime:time.Now(),
	}
	// 执行新增操作
	err := db.Create(&pack)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if err != nil {
		result["success"] = true
	}else{
		result["success"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 修改记录
func modPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()

	uphone := r.FormValue("uphone")
	uname := r.FormValue("uname")
	state := r.FormValue("state")
	havedial := r.FormValue("havedial")
	havemess := r.FormValue("havemess")

	var pack Pack
	db.First(&pack,ps.ByName("id"))

	if len(uphone) != 0 {
		pack.Userphone =uphone
	}

	if len(uname) != 0 {
		pack.Username =uname
	}

	if len(state) != 0 {
		pack.State,_ = strconv.Atoi(state)
		pack.Outtime = time.Now()
	}

	if len(havedial) != 0 {
		pack.Havedial,_ = strconv.Atoi(havedial)
	}

	if len(havemess) != 0 {
		pack.Havedial,_ = strconv.Atoi(havemess)
	}

	err := db.Save(&pack)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if err != nil {
		result["success"] = true
	}else{
		result["success"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 主程序入口
func main(){
	// 连接数据库
	db,_ = gorm.Open("mysql", "root:flame@tcp(211.159.218.41:3306)/packeep?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()

	// 关闭自动添加s到表名后面（模型名称）
	db.SingularTable(true)
	// 设置所有表的默认前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "tb_" + defaultTableName
	}

	// 启用gorm内置日志功能
	db.LogMode(true)

	// 定义路由
	router := httprouter.New()
	router.GET("/api/pack",getList)
	router.GET("/api/pack/:id",getPack)
	router.POST("/api/pack",addPack)
	router.PUT("/api/pack/:id",modPack)

	err2 := http.ListenAndServe(":8080",router)
	if err2 != nil {
		log.Fatalln("服务器监听错误：",err2)
	}else{
		log.Println("服务器监听启动，端口：8080")
	}
}

// 传入map字典，返回json格式
func getJson(data interface{})(output []byte){
	var content []byte
	var err error

	content,err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		content = []byte("convert to json fail")
	}
	return content
}