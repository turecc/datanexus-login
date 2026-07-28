package model

import "time"

// DataNexusUserModel 记录微信用户首次登录状态，用于 DataNexus REGISTER 去重。
type DataNexusUserModel struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OpenID      string    `gorm:"column:openId;type:varchar(128);not null;uniqueIndex:uk_datanexus_users_openid"`
	AppID       string    `gorm:"column:appId;type:varchar(64);not null;default:''"`
	UnionID     string    `gorm:"column:unionId;type:varchar(128);not null;default:''"`
	CreatedAt   time.Time `gorm:"column:createdAt;not null;autoCreateTime"`
	LastLoginAt time.Time `gorm:"column:lastLoginAt;not null"`
}

// TableName 固定表名，避免不同系统下大小写不一致。
func (DataNexusUserModel) TableName() string {
	return "DatanexusUsers"
}
