package model

func init() {
	register(&Store{})
}

type Store struct {
	Id        int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	Name      string `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Code      string `gorm:"column:code;type:varchar(32);uniqueIndex:uk_stores_code;not null" json:"code"`
	Address   string `gorm:"column:address;type:varchar(256);default:'';not null" json:"address"`
	Contact   string `gorm:"column:contact;type:varchar(64);default:'';not null" json:"contact"`
	Phone     string `gorm:"column:phone;type:varchar(20);default:'';not null" json:"phone"`
	Status    int8   `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (s *Store) TableName() string { return "stores" }
