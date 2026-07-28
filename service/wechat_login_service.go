package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"wxcloudrun-golang/db/dao"
)

// WechatLoginResponse 微信云托管登录接口返回结构。
type WechatLoginResponse struct {
	Success    bool   `json:"success"`
	OpenID     string `json:"openId,omitempty"`
	IsNewUser  bool   `json:"isNewUser"`
	ServerTime int64  `json:"serverTime,omitempty"`
	Error      string `json:"error,omitempty"`
}

// WechatLoginHandler 读取微信云托管注入的可信 OpenID，并判断是否首次登录。
func WechatLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	if r.Method != http.MethodPost {
		writeWechatLoginJSON(w, http.StatusMethodNotAllowed, WechatLoginResponse{
			Success: false,
			Error:   "METHOD_NOT_ALLOWED",
		})
		return
	}

	// 这些请求头由微信云托管在 wx.cloud.callContainer 调用时自动注入。
	openID := r.Header.Get("X-WX-OPENID")
	appID := r.Header.Get("X-WX-APPID")
	unionID := r.Header.Get("X-WX-UNIONID")

	if openID == "" {
		log.Println("[WechatLogin] X-WX-OPENID is empty")
		writeWechatLoginJSON(w, http.StatusUnauthorized, WechatLoginResponse{
			Success: false,
			Error:   "OPENID_EMPTY",
		})
		return
	}

	now := time.Now()
	isNewUser, err := dao.RegisterOrLoginDataNexusUser(
		openID,
		appID,
		unionID,
		now,
	)
	if err != nil {
		log.Printf("[WechatLogin] database error, user=%s, err=%v", maskOpenID(openID), err)
		writeWechatLoginJSON(w, http.StatusInternalServerError, WechatLoginResponse{
			Success: false,
			Error:   "DATABASE_ERROR",
		})
		return
	}

	log.Printf(
		"[WechatLogin] success, user=%s, isNewUser=%t",
		maskOpenID(openID),
		isNewUser,
	)

	writeWechatLoginJSON(w, http.StatusOK, WechatLoginResponse{
		Success:    true,
		OpenID:     openID,
		IsNewUser:  isNewUser,
		ServerTime: now.Unix(),
	})
}

func writeWechatLoginJSON(w http.ResponseWriter, status int, value WechatLoginResponse) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("[WechatLogin] encode response failed: %v", err)
	}
}

func maskOpenID(openID string) string {
	sum := sha256.Sum256([]byte(openID))
	return hex.EncodeToString(sum[:4])
}
