package controllers

import (
	"encoding/json"
	"loveHome/models"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) RetData(resp map[string]interface{}, respType string) {
	switch strings.ToLower(respType) {
	case "json":
		this.Data["json"] = &resp
		this.ServeJSON()
	case "xml":
		this.Data["xml"] = &resp
		this.ServeXML()
	case "jsonp":
		this.Data["jsonp"] = &resp
		this.ServeJSONP()
	}
}

func (this *UserController) Reg() {
	resp := make(map[string]interface{})

	defer this.RetData(resp, "json")

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)

	UnmarshalErr := json.Unmarshal(this.Ctx.Input.RequestBody, &resp)
	if UnmarshalErr != nil {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	userModel := models.User{}
	mobile := resp["mobile"].(string)
	userModel.Name = mobile
	userModel.Mobile = mobile
	userModel.Password_hash = resp["password"].(string)

	userid, err := orm.NewOrm().Insert(&userModel)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		this.RetData(resp, "json")

		return
	}

	// 存Session
	this.SetSession("name", mobile)
	this.SetSession("user_id", userid)
	this.SetSession("mobile", mobile)
}

// 上传图片
func (this *UserController) PostAvatar() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 获取上传数据
	f, h, err := this.GetFile("avatar")
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)

		return
	}
	defer f.Close()

	// 获取文件后缀
	ext := path.Ext(h.Filename)

	// 当前秒级时间戳
	filetime := time.Now().Unix()
	filetimeString := strconv.FormatInt(filetime, 10)

	// 获取session中的userid
	userid := this.GetSession("user_id")
	useridString := strconv.Itoa(userid.(int))

	dir := "static/upload/"+useridString+"/"

	// 创建目录 os.ModePerm: 0777权限
	os.MkdirAll(dir, os.ModePerm)

	// 图片名称
	filename := dir+filetimeString+ext

	// 保存图片
	SaveToFileErr := this.SaveToFile("avatar", filename)
	if SaveToFileErr != nil {
		resp["errno"] = 111
		resp["errmsg"] = "图片保存失败"

		return
	}

	upNum, upErr := orm.NewOrm().QueryTable("user").Filter("id", userid).Update(orm.Params{"avatar_url": filename})
	if upErr != nil || upNum == 0 {
		resp["errno"] = models.RECODE_USERERR
		resp["errmsg"] = models.RecodeText(models.RECODE_USERERR)

		return
	}

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = filename

	return

}


// 获取用户信息
func (this *UserController) GetUserData() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 从Session中获取用户ID
	userid := this.GetSession("user_id")

	userModel := models.User{}
	// 查找用户信息
	oneErr := orm.NewOrm().QueryTable("user").Filter("id", userid).One(&userModel, "id", "name", "mobile", "real_name", "id_card", "avatar_url")
	if oneErr != nil {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)

		return
	}

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = userModel

	return
}

// 更改用户名
func (this *UserController) UpUserName() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 获取当前用户id
	userid := this.GetSession("user_id")

	// 获取提交过来的用户名
	userNameMap := make(map[string]string)
	UnmarshalErr := json.Unmarshal(this.Ctx.Input.RequestBody, &userNameMap)
	if UnmarshalErr != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)

		return
	}

	// 更改用户名
	_, upErr := orm.NewOrm().QueryTable("user").Filter("id", userid).Update(orm.Params{"name": userNameMap["name"]})
	if upErr != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

		return
	}

	// 重新设置Session
	this.SetSession("name", userNameMap["name"])

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = userNameMap["name"]

	return
}

// 实名认证
func (this *UserController) UserAuth() {
	resp := make(map[string]interface{})
	defer this.RetData(resp, "json")

	// 获取当前提交的方法
	if this.Ctx.Request.Method == "POST" {

		// 获取提交过来的数据
		userModel := models.User{}
		UnmarshalErr := json.Unmarshal(this.Ctx.Input.RequestBody, &userModel)
		if UnmarshalErr != nil {
			resp["errno"] = models.RECODE_REQERR
			resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)

			return
		}

		// 获取当前Session用户id
		userid := this.GetSession("user_id")

		upNum, uperr := orm.NewOrm().QueryTable("user").Filter("id", userid).Update(orm.Params{"real_name": userModel.Real_name, "id_card": userModel.Id_card})
		if uperr != nil || upNum == 0 {
			resp["errno"] = models.RECODE_DBERR
			resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

			return
		}

		resp["errno"] = models.RECODE_OK
		resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	}


	return
}
