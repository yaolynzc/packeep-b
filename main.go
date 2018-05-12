package main

import (
	"bytes"
	"net/http"
	"log"
	"time"
	"github.com/jinzhu/gorm"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	_"github.com/go-sql-driver/mysql"
	"strconv"
	"io/ioutil"
	"strings"
	"encoding/base64"
	"image"
	"os"
	"fmt"
	"image/png"
)

// 定义数据库连接实例db对象
var db *gorm.DB
// 存入数据库的图片相对路径前缀
const picPrefix = "upload/pic/"

type Pack struct {
	Id int `gorm:"auto_increment"`
	Poscode string `gorm:"not null"`
	Userphone string `gorm:"not null"`
	Username string `gorm:"not null"`
	State  int
	Havedial int
	Havemess int
	Havepic string
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
		result["success"] = true
		result["pack"] = pack
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
	pack.Havepic = ""
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
	// 每次电话通知后，数值加1
	if len(havedial) != 0 {
		pack.Havedial += 1
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

// 删除记录
func delPack(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	id := ps.ByName("id")
	var pack Pack
	db.First(&pack, id) // 查询id为1的product

	// 删除
	res := db.Delete(&pack)

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

// 拍照上传并签收
func uploadPic(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()
	// 获取传值
	havepic := r.FormValue("havepic")
	state := r.FormValue("state")
	id,_ := strconv.Atoi(ps.ByName("id"))

	var pack Pack
	db.First(&pack,id)

	// 更新state状态值
	if len(state) != 0 {
		pack.State,_ = strconv.Atoi(state)
		pack.Outtime = time.Now()
	}

	// 如果图片不为空，保存图片文件到当前upload/pic目录，数据库存入相对路径
	if len(havepic) != 0 {
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(havepic))
		// 转换成png格式的图像，需要导入：_“image/png”
		pngPic, _, _ := image.Decode(reader)

		// 图片存储路径检测
		havePath,_ := PathExists(picPrefix)
		if(havePath){
			// 图片保存到磁盘目录
			nowstr := time.Now().Unix()
			picPath := picPrefix + strconv.Itoa(id) + "_" + strconv.FormatInt(nowstr,10) + ".png"
			wt, err := os.Create(picPath)
			if err != nil {
				fmt.Println("图片保存失败!")
			}
			defer wt.Close()
			// 转换为jpeg格式的图像，这里质量为30（质量取值是1-100）
			//jpeg.Encode(wt, m, &jpeg.Options{30})
			png.Encode(wt,pngPic)
			// 图片路径值写入数据库
			pack.Havepic = picPath
		}
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

// 根据手机号部分数字获取符合要求的手机号和用户姓名distinct
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
	rows, _ := db.Model(&Pack{}).Where("userphone like ? and username is not null","%"+uphone+"%").Select("distinct userphone,username ").Limit(size).Rows()
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

// 根据手机号获取用户姓名
func getUnameByPhone(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	uphone := ps.ByName("tel")

	var pack Pack
	res := db.Where("userphone = ? and username is not null", uphone).First(&pack)

	//定义返回的数据结构
	result := make(map[string]interface{})
	if res.RowsAffected > 0 {
		result["success"] = true
		result["pack"] = pack
	}else{
		result["success"] = false
	}

	// 返回json
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJson(result))
}

// 接收前端post请求，通过后台post实现给指定号码拨打电话，并将结果返回前端
func sendDialByPhone(w http.ResponseWriter,r *http.Request,ps httprouter.Params){
	r.ParseForm()

	uphone := r.FormValue("uphone")
	uname := r.FormValue("uname")

	//定义返回的数据结构
	result := make(map[string]interface{})

	// 配置“赛邮-云通信-mysubmail”语音通知API基本参数
	config := make(map[string]interface{})
	vars := make(map[string]string)
	url := "https://api.mysubmail.com/voice/xsend"
	config["appid"] = "20694"
	config["signature"] = "c0428970b976d160b77d59b3e28d7137"
	config["to"] = uphone
	config["project"] = "ux0bf2"
	vars["name"] = uname
	config["vars"] = vars

	postRes := HttpPost(url,config)
	if(strings.Contains(postRes,"success")){
		result["success"] = true
	}else{
		result["success"] = false
	}
	// 返回json
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
	router.DELETE("/api/pack/:id", delPack)
	router.GET("/api/phone", getUphoneList)
	router.GET("/api/phone/:tel", getUnameByPhone)
	router.POST("/api/phone", sendDialByPhone)
	router.POST("/api/pack/:id", uploadPic)

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

// 通过go发送post请求，通过mysubmail api接口拨打电话
func HttpPost(queryurl string, postdata map[string]interface{}) string {
	data, err := json.Marshal(postdata)
	if err != nil {
		return err.Error()
	}

	body := bytes.NewBuffer([]byte(data))

	retstr, err := http.Post(queryurl, "application/json;charset=utf-8", body)

	if err != nil {
		return err.Error()
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return err.Error()
	}
	return string(result)
}

// 判断目录是否存在，不存在则创建
func PathExists(path string) (bool, error){
	_, err := os.Stat(path)
	// 存在直接返回true
	if err == nil {
		return true, nil
	}
	// 不存在则创建，然后返回true
	if os.IsNotExist(err) {
		// 递归创建目录
		err := os.MkdirAll(path, os.ModePerm)
		if(err == nil) {
			return true, nil
		}
	}
	return false, err
}