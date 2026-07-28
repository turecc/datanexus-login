package model

import "time"

// DataNexusUserModel 记录微信用户首次登录状态，用于 DataNexus REGISTER 去重。
// OpenID 上的唯一索引保证并发登录时只会创建一条用户记录。
type DataNexusUserModel struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OpenID      string    `gorm:"column:openId;type:varchar(128);not null;uniqueIndex:uk_datanexus_users_openid" json:"openId"`
	AppID       string    `gorm:"column:appId;type:varchar(64);not null;default:''" json:"appId"`
	UnionID     string    `gorm:"column:unionId;type:varchar(128);not null;default:''" json:"unionId"`
	CreatedAt   time.Time `gorm:"column:createdAt;not null;autoCreateTime" json:"createdAt"`
	LastLoginAt time.Time `gorm:"column:lastLoginAt;not null" json:"lastLoginAt"`
}

// TableName 固定表名，避免不同系统下大小写不一致。
func (DataNexusUserModel) TableName() string {
	return "DatanexusUsers"
}
