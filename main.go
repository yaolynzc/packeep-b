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
	Id int `gorm:"auto_increment"`
	Poscode string `gorm:"not null"`
	Userphone string `gorm:"not null"`
	Username string `gorm:"not null"`
	State  int
	Havedial int
	Havemess int
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
	state := r.FormValue("state")
	page := r.FormValue("page")
	size := r.FormValue("size")

	// 定义返回的数据结构
	result := make(map[string]interface{})

	// 符合查询条件的总数
	if len(page) == 0 && len(size) == 0 {
		var count int

		// 查询全部
		rs := db.Model(&Pack{}).Where("1=1").Count(&count)
		// 根据uphone筛选
		if len(uphone) != 0 {
			rs = rs.Where("userphone like ?", "%"+uphone+"%").Count(&count)
		}
		// 根据state筛选
		if len(state) != 0 {
			state_int,_ := strconv.Atoi(state)
			rs = rs.Where("state = ?",state_int).Count(&count)
		}

		result["count"] = count		// 符合条件的记录总数
	}else {
		// 页码和每页个数
		page_int, _ := strconv.Atoi(page)
		size_int, _ := strconv.Atoi(size)

		// 定义pack切片，存储数据集
		var packs []*Pack

		// 查询全部
		rs := db.Find(&packs)
		// 根据uphone筛选
		if len(uphone) != 0 {
			//rs = rs.Where("userphone like ?", "%"+uphone+"%").Order("intime desc").Offset((page_int - 1) * size_int).Limit(size_int).Find(&packs)
			rs = rs.Where("userphone like ?", "%"+uphone+"%").Find(&packs)
		}
		// 根据state筛选
		if len(state) != 0 {
			state_int,_ := strconv.Atoi(state)
			rs = rs.Where("state = ?",state_int).Find(&packs)
		}
		// 排序，分页
		if state == "1" {
			rs = rs.Order("outtime desc").Offset((page_int - 1) * size_int).Limit(size_int).Find(&packs)
		}else {
			rs = rs.Order("intime desc").Offset((page_int - 1) * size_int).Limit(size_int).Find(&packs)
		}

		result["dt"] = packs
	}

	result["success"] = true

	// 返回json结果
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 根据id查询
func getPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	id,_ := strconv.Atoi(ps.ByName("id"))
	var pack Pack
	res := db.First(&pack,id)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if res.RowsAffected  > 0 {
		result["pack"] = pack
		result["success"] = true
	}else{
		result["success"] = false
	}

	// 返回json
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 新增记录
func addPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()

	pcode := r.FormValue("pcode")
	uphone := r.FormValue("uphone")
	uname := r.FormValue("uname")

	var pack Pack
	pack.Poscode = pcode
	pack.Userphone = uphone
	pack.Username = uname
	pack.State = 0
	pack.Havedial = 0
	pack.Havemess = 0
	pack.Intime = time.Now()
	pack.Outtime = time.Now()

	// 执行新增
	res := db.Create(&pack)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if res.RowsAffected > 0 {
		result["success"] = true
	}else{
		result["success"] = false
	}

	// 返回json
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 修改记录
func modPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()
	// 获取传值
	pcode := r.FormValue("pcode")
	uphone := r.FormValue("uphone")
	uname := r.FormValue("uname")
	state := r.FormValue("state")
	havedial := r.FormValue("havedial")
	havemess := r.FormValue("havemess")

	id,_ := strconv.Atoi(ps.ByName("id"))
	var pack Pack
	db.First(&pack,id)

	if len(pcode) != 0 {
		pack.Poscode =pcode
	}

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

	// 执行更新
	res := db.Save(&pack)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if res.RowsAffected > 0 {
		result["success"] = true
	}else{
		result["success"] = false
	}

	// 返回json
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 根据手机号分组获取packs
func getUphoneList(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()
	uphone := r.FormValue("uphone")
	size := r.FormValue("size")

	// 定义返回的数据结构
	result := make(map[string]interface{})

	// 定义只有用户电话和姓名的结构体
	type Uphonename struct {
		Userphone string
		Username string
	}

	// 定义uphonenames切片，存储数据集
	var uphonenames []Uphonename
	rows, _ := db.Model(&Pack{}).Where("userphone like ?","%"+uphone+"%").Select("distinct userphone,username ").Limit(size).Rows()
	defer rows.Close()
	for rows.Next() {
		var uphonename Uphonename
		db.ScanRows(rows, &uphonename)
		//fmt.Printf(uphonename.Userphone)
		uphonenames = append(uphonenames,uphonename)
	}
	result["dt"] = uphonenames
	result["success"] = true

	// 返回json结果
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 主程序入口
func main(){
	// 连接数据库
	db,_ = gorm.Open("mysql", "root:flame@tcp(211.159.218.41:3306)/packeepdev?charset=utf8&parseTime=True&loc=Local")
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
	router.GET("/api/phone",getUphoneList)

	err := http.ListenAndServe(":8080",router)
	if err != nil {
		log.Fatalln("服务器成功启动，端口：8080")
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