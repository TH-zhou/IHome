package controllers

import (
	"encoding/json"
	"loveHome/models"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	_ "github.com/astaxie/beego/cache/redis"
)

type AreaController struct {
	beego.Controller
}


func (this *AreaController) RetData(resp map[string]interface{}, respType string) {
	switch strings.ToLower(respType) {
	case "json":
		this.Data["json"] = &resp
		this.ServeJSON()
	case "xml":
		this.Data["xml"] = &resp
		this.ServeXML()
	case "josnp":
		this.Data["jsonp"] = &resp
		this.ServeJSONP()
	default:
		this.Data["json"] = &resp
		this.ServeJSON()
	}
}

func (this *AreaController) GetArea() {
	resp := make(map[string]interface{})

	// 打包成json给前端
	defer this.RetData(resp, "json")

	// 从Redis中读取数据
	// 初始化一个全局变量对象
	cache_conn, err := cache.NewCache("redis", `{"key":"lovehome","conn":":6379","dbNum":"0"}`)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}

	areasModel := []models.Area{}

	// 从redis中取数据
	if cache_conn.IsExist("area") {
		if redisArea := cache_conn.Get("area"); redisArea != nil {
			resp["errno"] = models.RECODE_OK
			resp["errmsg"] = models.RecodeText(models.RECODE_OK)
			//area := json.Unmarshal(redisArea, &models.Area{})
			json.Unmarshal(redisArea.([]byte), &areasModel)
			resp["data"] = areasModel

			beego.Info("取到数据了")

			return
		}
	}

	// 从数据库中读取数据

	_, allErr := orm.NewOrm().QueryTable("area").All(&areasModel)
	if allErr != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = &areasModel


	// 将area数据存入Redis
	jsonStr, jsonErr := json.Marshal(areasModel)
	if jsonErr != nil {
		resp["error"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)

		return
	}

	redisPutErr := cache_conn.Put("area", jsonStr, time.Second * 3600)
	if redisPutErr != nil {
		resp["errno"] = 1234
		resp["errmsg"] = "缓存操作失败了"

		return
	}
}
