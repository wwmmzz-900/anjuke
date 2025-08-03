package model

import (
	"time"
)

// House 房源信息表结构体
type House struct {
	HouseId                 int64      `gorm:"column:house_id;primaryKey;autoIncrement" json:"house_id"`                                                // 房源ID
	Title                   string     `gorm:"column:title;type:varchar(100);not null" json:"title"`                                                    // 房源标题
	Description             string     `gorm:"column:description;type:text" json:"description"`                                                         // 房源描述
	LandlordId              int64      `gorm:"column:landlord_id;not null" json:"landlord_id"`                                                          // 发布人ID
	Address                 string     `gorm:"column:address;type:varchar(255);not null" json:"address"`                                                // 详细地址
	RegionId                *int64     `gorm:"column:region_id" json:"region_id"`                                                                       // 区域/小区ID
	CommunityId             *int64     `gorm:"column:community_id" json:"community_id"`                                                                 // 小区ID
	Price                   float64    `gorm:"column:price;type:decimal(10,2);not null" json:"price"`                                                   // 价格
	Area                    *float64   `gorm:"column:area;type:float" json:"area"`                                                                      // 面积
	Layout                  *string    `gorm:"column:layout;type:varchar(50)" json:"layout"`                                                            // 户型
	Floor                   *string    `gorm:"column:floor;type:varchar(20)" json:"floor"`                                                              // 楼层
	OwnershipCertificateUrl string     `gorm:"column:ownership_certificate_url;type:varchar(255);not null" json:"ownership_certificate_url"`            // 产权证明图片
	Orientation             *string    `gorm:"column:orientation;type:varchar(20)" json:"orientation"`                                                  // 朝向
	Decoration              *string    `gorm:"column:decoration;type:varchar(50)" json:"decoration"`                                                    // 装修
	Facilities              *string    `gorm:"column:facilities;type:varchar(255)" json:"facilities"`                                                   // 配套设施（逗号分隔）
	Status                  *string    `gorm:"column:status;type:enum('active','inactive','rented')" json:"status"`                                     // 状态
	HouseType               string     `gorm:"column:house_type;type:enum('rent','second_hand','new_house');not null" json:"house_type"`                // 租房状态（租房/二手房/新房）
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:current_timestamp" json:"created_at"`                                  // 发布时间
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp" json:"updated_at"`      // 更新时间
	DeletedAt               *time.Time `gorm:"column:deleted_at" json:"deleted_at"`                                                                     // 删除时间
	OpeningTime             *time.Time `gorm:"column:opening_time" json:"opening_time"`                                                                 // 开盘时间
	OpeningStatus           *string    `gorm:"column:opening_status;type:enum('pending','opened','cancelled');default:'pending'" json:"opening_status"` // 开盘状态
	ViewCount               *int64     `gorm:"column:view_count" json:"view_count"`                                                                     // 浏览次数
	FavoriteCount           *int64     `gorm:"column:favorite_count" json:"favorite_count"`                                                             // 收藏次数
	LikesCount              *int64     `gorm:"column:likes_count" json:"likes_count"`                                                                   // 点赞次数
}

// TableName 指定结构体对应的数据库表名
func (House) TableName() string {
	return "house"
}
