package model

import "time"

type House struct {
	HouseId                 int64     `gorm:"column:house_id;type:bigint;comment:房源ID;primaryKey;not null;" json:"house_id"`                                               // 房源ID
	Title                   string    `gorm:"column:title;type:varchar(100);comment:房源标题;not null;" json:"title"`                                                          // 房源标题
	Description             string    `gorm:"column:description;type:text;comment:房源描述;" json:"description"`                                                               // 房源描述
	LandlordId              int64     `gorm:"column:landlord_id;type:bigint;comment:发布人ID;not null;" json:"landlord_id"`                                                   // 发布人ID
	Address                 string    `gorm:"column:address;type:varchar(255);comment:详细地址;not null;" json:"address"`                                                      // 详细地址
	RegionId                int64     `gorm:"column:region_id;type:bigint;comment:区域/小区ID;default:NULL;" json:"region_id"`                                                 // 区域/小区ID
	CommunityId             int64     `gorm:"column:community_id;type:bigint;comment:小区ID;default:NULL;" json:"community_id"`                                              // 小区ID
	Price                   float64   `gorm:"column:price;type:decimal(10, 2);comment:价格;not null;" json:"price"`                                                          // 价格
	Area                    float32   `gorm:"column:area;type:float;comment:面积;default:NULL;" json:"area"`                                                                 // 面积
	Layout                  string    `gorm:"column:layout;type:varchar(50);comment:户型;default:NULL;" json:"layout"`                                                       // 户型
	Floor                   string    `gorm:"column:floor;type:varchar(20);comment:楼层;default:NULL;" json:"floor"`                                                         // 楼层
	OwnershipCertificateUrl string    `gorm:"column:ownership_certificate_url;type:varchar(255);comment:产权证明图片;not null;" json:"ownership_certificate_url"`                // 产权证明图片
	Orientation             string    `gorm:"column:orientation;type:varchar(20);comment:朝向;default:NULL;" json:"orientation"`                                             // 朝向
	Decoration              string    `gorm:"column:decoration;type:varchar(50);comment:装修;default:NULL;" json:"decoration"`                                               // 装修
	Facilities              string    `gorm:"column:facilities;type:varchar(255);comment:配套设施（逗号分隔）;default:NULL;" json:"facilities"`                                      // 配套设施（逗号分隔）
	Status                  string    `gorm:"column:status;type:enum('active', 'inactive', 'rented');comment:状态（“活跃”，“不活跃”，“已租用”）;not null;default:active;" json:"status"` // 状态（“活跃”，“不活跃”，“已租用”）
	CreatedAt               time.Time `gorm:"column:created_at;type:datetime;comment:发布时间;not null;default:CURRENT_TIMESTAMP;" json:"created_at"`                          // 发布时间
	UpdatedAt               time.Time `gorm:"column:updated_at;type:datetime;comment:更新时间;not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`                          // 更新时间
	DeletedAt               time.Time `gorm:"column:deleted_at;type:datetime;comment:删除时间;default:NULL;" json:"deleted_at"`                                                // 删除时间
}

func (*House) TableName() string {
	return "house"
}

type HouseReservation struct {
	ID          int64 `gorm:"primaryKey"`
	LandlordID  int64
	UserID      int64
	UserName    string
	HouseID     int64
	HouseTitle  string
	ReserveTime string
	CreatedAt   int64
}
