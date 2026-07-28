package dao

import (
	"time"

	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"

	"gorm.io/gorm/clause"
)

// RegisterOrLoginDataNexusUser 写入首次登录记录，已存在时只更新最后登录时间。
// 返回 true 表示首次成功插入，false 表示该 OpenID 已存在。
func RegisterOrLoginDataNexusUser(
	openID string,
	appID string,
	unionID string,
	now time.Time,
) (bool, error) {
	cli := db.Get()

	user := &model.DataNexusUserModel{
		OpenID:      openID,
		AppID:       appID,
		UnionID:     unionID,
		CreatedAt:   now,
		LastLoginAt: now,
	}

	// 依赖 openId 唯一索引保证并发安全：只有第一条请求可以成功插入。
	createResult := cli.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(user)
	if createResult.Error != nil {
		return false, createResult.Error
	}
	if createResult.RowsAffected == 1 {
		return true, nil
	}

	updates := map[string]interface{}{
		"lastLoginAt": now,
	}
	if appID != "" {
		updates["appId"] = appID
	}
	if unionID != "" {
		updates["unionId"] = unionID
	}

	updateResult := cli.
		Model(&model.DataNexusUserModel{}).
		Where("openId = ?", openID).
		Updates(updates)
	if updateResult.Error != nil {
		return false, updateResult.Error
	}

	return false, nil
}
