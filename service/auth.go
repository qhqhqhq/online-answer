package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"online-answer/db"
	"online-answer/db/model"
	"online-answer/utils"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	var loginResp LoginResponse
	cli := db.Get()

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&loginReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wxresp, err := fetchWXCode2Session(loginReq.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	loginResp.Token = wxresp.Openid

	err = cli.Transaction(func(tx *gorm.DB) error {
		var group model.Group
		// 使用行级锁来查询组
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Members").Where(&model.Group{Secret: loginReq.Secret}).First(&group).Error != nil {
			return errors.New("group not found")
		}
		loginResp.GroupNumber = group.Number

		if len(group.Members) >= 3 {
			for _, member := range group.Members {
				if member.OpenID == wxresp.Openid {
					return nil
				}
			}
			return errors.New("full group")
		}

		// 添加新成员
		if err := tx.Save(&model.User{OpenID: wxresp.Openid, GroupNumber: group.Number}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg, _ := json.Marshal(&loginResp)
	w.Header().Set("content-type", "application/json")
	w.Write(msg)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	openId, groupNumber, err := utils.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	cli := db.Get()
	if cli.Delete(&model.User{OpenID: openId, GroupNumber: groupNumber}).Error != nil {
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}

}

func fetchWXCode2Session(code string) (*WXCode2SessionResponse, error) {
	var APP_ID = os.Getenv("APP_ID")
	var APP_SECRET = os.Getenv("APP_SECRET")

	var code2sessionResp WXCode2SessionResponse
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", APP_ID, APP_SECRET, code)
	wxresp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("wx server not available")
	}
	defer wxresp.Body.Close()

	body, err := io.ReadAll(wxresp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read wx response")
	}

	err = json.Unmarshal(body, &code2sessionResp)
	if err != nil {
		return nil, fmt.Errorf("cannot parse wx response")
	}

	return &code2sessionResp, nil
}
