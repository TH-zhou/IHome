package controllers

import (
	"encoding/json"
	"loveHome/models"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type OrdersController struct {
	beego.Controller
}

func (this *OrdersController) RetData(resp map[string]interface{}, respType string) {
	switch respType {
	case "json":
		this.Data["json"] = resp
		this.ServeJSON()
	case "xml":
		this.Data["xml"] = resp
		this.ServeXML()
	case "jsonp":
		this.Data["jsonp"] = resp
		this.ServeJSONP()
	default:
		this.Data["json"] = resp
		this.ServeJSON()
	}
}

func (this *OrdersController) PostOrders() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	requestBodyMap := make(map[string]string)
	UnmarshalErr := json.Unmarshal(this.Ctx.Input.RequestBody, &requestBodyMap)
	if UnmarshalErr != nil {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	// 从session中获取用乎userid
	userid := this.GetSession("user_id").(int)

	// 判断时间end_date在start_date之后
	end := this.timeUnix(requestBodyMap["end_date"])
	start := this.timeUnix(requestBodyMap["start_date"])
	if end < start {
		resp["errno"]= 11
		resp["errmsg"] = "结束时间早于开始时间"

		return
	}

	diffDay := (end - start) / (3600 * 24) + 1

	// 获取房子id
	house_id := requestBodyMap["house_id"]
	if house_id == "undefined" {
		house_id = "3"
	}
	house_id_int, _ := strconv.Atoi(house_id)

	houseModel := models.House{Id: house_id_int}
	ormer := orm.NewOrm()
	ReadErr := ormer.Read(&houseModel)
	if ReadErr != nil {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	//房主不能预定自己的房间
	if userid == houseModel.User.Id {
		resp["errno"] = 12345
		resp["errmsg"] = "不能预定自己的房子"

		return
	}

	// 判断该房间在当前时间有没有被预定
	num, numErr := ormer.QueryTable("order_house").Filter("House", house_id_int).Filter("ctime__gte", start).Filter("ctime__lt", end).Count()
	if numErr != nil {
		resp["errno"] = 22
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}
	if num > 0 {
		resp["errno"] = 123456
		resp["errmsg"] = "房子已被预约"

		return
	}

	// 封装完整的order信息
	orderModel := models.OrderHouse{}
	userModel := models.User{Id: userid}
	orderModel.User = &userModel
	orderModel.House = &houseModel
	orderModel.Begin_date, _ = time.Parse("2006-01-02 15:04:05", time.Unix(start, 0).Format("2006-01-02 15:04:05"))
	orderModel.End_date, _ = time.Parse("2006-01-02 15:04:05", time.Unix(end, 0).Format("2006-01-02 15:04:05"))
	orderModel.Days = int(diffDay)
	orderModel.House_price = houseModel.Price
	orderModel.Amount = houseModel.Price * int(diffDay)
	orderModel.Status = models.ORDER_STATUS_PAID
	orderModel.Comment = "Good"
	orderModel.Ctime, _ = time.Parse("2006-01-02 15:04:05", time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05"))

	beego.Info(orderModel)
	orderid, insertErr := ormer.Insert(&orderModel)
	if insertErr != nil {
		resp["errno"] = 11
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = orderid
	return
}

func (this *OrdersController) timeUnix(timeString string) int64 {
	timeLayout := "2006-01-02"
	local, _ := time.LoadLocation("Local")
	tmp, _ := time.ParseInLocation(timeLayout, timeString, local)
	timestamp := tmp.Unix()

	return timestamp
}

func (this *OrdersController) GetOrdersData() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	//role := this.GetString("role")
	userid := this.GetSession("user_id").(int)
	ordersModel := []models.OrderHouse{}
	ormer := orm.NewOrm()
	_, err := ormer.QueryTable("order_house").Filter("User", userid).RelatedSel().All(&ordersModel)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}

	orderMap := make(map[string]interface{})
	orderMap["orders"] = ordersModel

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = orderMap

	return
}
