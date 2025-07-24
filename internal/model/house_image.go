package model

// 房源图片
type HouseImage struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                // 图片ID
	HouseID   int64  `gorm:"column:house_id;not null;index" json:"house_id"`              // 房源ID
	ImageURL  string `gorm:"column:image_url;size:255;not null" json:"image_url"`         // 图片URL
	SortOrder int32  `gorm:"column:sort_order;not null;default:0" json:"sort_order"`      // 排序顺序
}

// 表名
func (HouseImage) TableName() string {
	return "house_image"
}