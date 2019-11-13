package controllers

import (
	"encoding/json"
	"loveHome/models"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type HouseIndexController struct {
	beego.Controller
}

func (this *HouseIndexController) RetData(resp map[string]interface{}, respType string) {
	switch strings.ToLower(respType) {
	case "json":
		this.Data["json"] = &resp
		this.ServeJSON()
	case "xml":
		this.Data["xml"] = &resp
		this.ServeXML()
	case "jsonp":
		this.Data["jsonp"] = &resp
		this.ServeXML()
	}
}

func (this *HouseIndexController) GetHouseIndex() {
	resp := make(map[string]interface{})

	// 将数据转为json发送给前端
	defer this.RetData(resp, "json")

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}

func (this *HouseIndexController) GetHouseData() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 获取userid
	userid := this.GetSession("user_id")

	houseModel:= []models.House{}
	ormer := orm.NewOrm()

	qs := ormer.QueryTable("house")
	num, allErr := qs.Filter("user__id", userid).All(&houseModel)
	if allErr != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}
	if num == 0 {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	respData := make(map[string]interface{})

	respData["houses"] = houseModel
	resp["data"] = respData

	return
}


func (this *HouseIndexController) PostHouseData() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 从前端拿到数据
	respData := make(map[string]interface{})
	json.Unmarshal(this.Ctx.Input.RequestBody, &respData)

	// 将数据插入到数据库
	house := models.House{}
	house.Title = respData["title"].(string)
	price, _ := strconv.Atoi(respData["price"].(string))
	house.Price = price
	house.Address = respData["address"].(string)
	room_count, _ := strconv.Atoi(respData["room_count"].(string))
	house.Room_count = room_count
	acreage, _ := strconv.Atoi(respData["acreage"].(string))
	house.Acreage = acreage
	house.Unit = respData["unit"].(string)
	capacity, _ := strconv.Atoi(respData["capacity"].(string))
	house.Capacity = capacity
	house.Beds = respData["beds"].(string)
	deposit, _ := strconv.Atoi(respData["deposit"].(string))
	house.Deposit = deposit
	min_days, _ := strconv.Atoi(respData["min_days"].(string))
	house.Min_days = min_days
	max_days, _ := strconv.Atoi(respData["max_days"].(string))
	house.Max_days = max_days

	// 设施处理
	facilitys := []models.Facility{}
	for _, fid := range respData["facility"].([]interface{}) {
		f_id, _ := strconv.Atoi(fid.(string))
		fac := models.Facility{Id: f_id}
		facilitys = append(facilitys, fac)
	}

	// 关联地区
	area_id, _ := strconv.Atoi(respData["area_id"].(string))
	area := models.Area{Id: area_id}
	house.Area = &area

	// 关联用户
	user := models.User{Id: this.GetSession("user_id").(int)}
	house.User = &user

	// 多对多添加数据
	ormer := orm.NewOrm()

	// 开启事务
	ormer.Begin()

	house_id, house_id_err := ormer.Insert(&house) // 先添加主，也就是房子信息
	if house_id_err != nil {
		resp["errno"] = 1010
		resp["errmsg"] = "房子信息添加失败"

		// 事务回滚
		ormer.Rollback()

		return
	}
	house.Id = int(house_id)
	// 得到房子信息之后，和设施绑定添加多个设施
	m2m := ormer.QueryM2M(&house, "Facilities")
	m2mNum, m2mErr := m2m.Add(facilitys)
	if m2mErr != nil || m2mNum == 0 {
		resp["errno"] = 1011
		resp["errmsg"] = "设置信息添加失败"

		// 事务回滚
		ormer.Rollback()

		return
	}

	returnData := make(map[string]interface{})
	returnData["house_id"] = house_id

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	respData["data"] = returnData

	// 事务提交
	ormer.Commit()

	return
}


func (this *HouseIndexController) HouseData() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	houseMap := make(map[string]interface{})

	// 获取id参数
	id := this.GetString(":id")
	idInt, _ := strconv.Atoi(id)

	// 获取当前用户userid
	userid := this.GetSession("user_id")

	ormer := orm.NewOrm()
	house := models.House{Id: idInt}
	if reanErr := ormer.Read(&house); reanErr != nil {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	ormer.LoadRelated(&house, "Area")
	ormer.LoadRelated(&house, "User")
	ormer.LoadRelated(&house, "Images")
	ormer.LoadRelated(&house, "Facilities")

	users := models.User{Id: userid.(int)}
	house.User = &users

	houseMap["house"] = house

	resp["data"] = houseMap
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)

	return


	//houseMap["acreage"] = house.Acreage
	//houseMap["address"] = house.Address
	//houseMap["beds"] = house.Beds
	//houseMap["capacity"] = house.Capacity
	//houseMap["deposit"] = house.Deposit
	//houseMap["facilities"] = house.Facilities
	//houseMap["img_urls"] = house.Images
	//houseMap["mix_days"] = house.Min_days
	//houseMap["max_days"] = house.Max_days
	//houseMap["price"] = house.Price


}
