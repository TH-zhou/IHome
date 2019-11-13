package controllers

import (
	"encoding/json"
	"loveHome/models"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type SessionsController struct {
	beego.Controller
}

func (this *SessionsController) RetData(resp map[string]interface{}, respType string) {
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


func (this *SessionsController) Login() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	userModel := models.User{}

	jsonErr := json.Unmarshal(this.Ctx.Input.RequestBody, &userModel)
	if jsonErr != nil {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		//resp["data"] = userModel

		return
	}

	queryErr := orm.NewOrm().QueryTable("user").Filter("mobile", userModel.Mobile).Filter("password_hash", userModel.Password_hash).One(&userModel)
	if queryErr == orm.ErrNoRows {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	// 记录Session
	this.SetSession("name", userModel.Name)
	this.SetSession("user_id", userModel.Id)
	this.SetSession("mobile", userModel.Mobile)

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}