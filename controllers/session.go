package controllers

import (
	"loveHome/models"
	"strings"

	"github.com/astaxie/beego"
)

type SessionController struct {
	beego.Controller
}

func (this *SessionController) RetData(resp map[string]interface{}, respType string) {
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

//  获取Session数据
func (this *SessionController) SessionData() {
	resp := make(map[string]interface{})
	// 将json数据返回给前端
	defer this.RetData(resp, "json")

	userModel := models.User{}

	// 获取Session
	name, ok := this.GetSession("name").(string)
	if !ok {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		resp["data"] = userModel

		return
	}

	userModel.Name = name
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = userModel

}

// 清空Session
func (this *SessionController) DelSessionData() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 清除Session
	this.DelSession("name")

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}
