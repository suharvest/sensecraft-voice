package model

func init() {
	register(&User{})
}

type User struct {
	Id        int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	Username  string `gorm:"column:username;type:varchar(64);uniqueIndex:uk_users_username;not null" json:"username"`
	Password  string `gorm:"column:password;type:varchar(128);not null" json:"-"` // 密码不返回给前端
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null;autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null;autoUpdateTime:milli" json:"updated_at"`
}

func (u *User) TableName() string { return "users" }
