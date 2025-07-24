package data

import (
	"context"
	"fmt"
	"github.com/wwmmzz-900/anjuke/internal/model"

	"github.com/wwmmzz-900/anjuke/internal/biz"
)

type houseRepo struct {
	data *Data // Data 结构体包含 *sql.DB
}

func NewHouseRepo(data *Data) biz.HouseRepo {
	return &houseRepo{data: data}
}

// 查询用户最近浏览的房源价格区间
func (r *houseRepo) GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error) {
	type result struct {
		MinPrice float64
		MaxPrice float64
	}

	var res result

	// 使用user_behavior表查询用户浏览过的房源价格区间
	err := r.data.db.
		Table("user_behavior AS ub").
		Select("MIN(h.price) AS min_price, MAX(h.price) AS max_price").
		Joins("JOIN house h ON ub.house_id = h.house_id"). // 使用正确的表名house
		Where("ub.user_id = ? AND ub.behavior = 'view'", userID).
		Order("ub.created_at DESC").
		Limit(20).
		Scan(&res).Error

	if err != nil {
		// 如果查询失败，返回默认价格区间
		return 800, 5000, nil
	}

	// 如果没有浏览记录或价格为0，返回默认区间
	if res.MinPrice == 0 && res.MaxPrice == 0 {
		return 800, 5000, nil
	}

	return res.MinPrice, res.MaxPrice, nil
}

// 查询个性化推荐房源
func (r *houseRepo) GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*biz.House, int, error) {
	var houses []*biz.House
	var total int64

	// 添加详细日志
	fmt.Printf("开始查询个性化推荐: minPrice=%.2f, maxPrice=%.2f, page=%d, pageSize=%d\n",
		minPrice, maxPrice, page, pageSize)

	// 尝试从数据库查询总数
	err := r.data.db.
		Table("house").
		Where("price BETWEEN ? AND ?", minPrice, maxPrice).
		Where("status = 'active'"). // 使用active状态
		Count(&total).Error

	if err != nil {
		fmt.Printf("查询个性化推荐总数失败: %v\n", err)
		// 如果查询失败，返回错误
		return nil, 0, fmt.Errorf("查询个性化推荐总数失败: %v", err)
	}

	fmt.Printf("查询到符合条件的房源总数: %d\n", total)

	// 使用已有的模型定义
	var results []model.House

	// 查询房源列表
	err = r.data.db.Debug(). // 添加Debug()以输出SQL语句
					Table("house").
					Select("house_id, title, description, price, area, layout").
					Where("price BETWEEN ? AND ?", minPrice, maxPrice).
					Where("status = 'active'").
					Order("house_id DESC").
					Limit(pageSize).
					Offset((page - 1) * pageSize).
					Scan(&results).Error

	if err != nil {
		fmt.Printf("查询个性化推荐列表失败: %v\n", err)
		// 如果查询失败，返回错误
		return nil, 0, fmt.Errorf("查询个性化推荐列表失败: %v", err)
	}

	// 如果没有数据，返回默认数据而不是错误
	if len(results) == 0 {
		fmt.Printf("未找到符合条件的房源，返回默认数据\n")
		// 返回默认数据
		return getDefaultHouses(), 3, nil
	}

	// 获取房源ID列表
	houseIDs := make([]int64, len(results))
	for i, result := range results {
		houseIDs[i] = result.HouseId
	}

	// 批量获取房源图片
	imageMap := r.getHouseImages(houseIDs)

	// 转换为业务层结构体
	houses = make([]*biz.House, 0, len(results))
	for _, result := range results {
		imageURL := imageMap[result.HouseId]

		houses = append(houses, &biz.House{
			HouseID:     result.HouseId,
			Title:       result.Title,
			Description: result.Description,
			Price:       result.Price,
			Area:        float64(result.Area),
			Layout:      result.Layout,
			ImageURL:    imageURL,
		})
	}

	fmt.Printf("成功查询到 %d 条个性化推荐房源\n", len(houses))
	for i, house := range houses {
		fmt.Printf("房源 %d: ID=%d, 标题=%s, 价格=%.2f\n", i+1, house.HouseID, house.Title, house.Price)
	}

	return houses, int(total), nil
}

// 批量获取房源图片
func (r *houseRepo) getHouseImages(houseIDs []int64) map[int64]string {
	if len(houseIDs) == 0 {
		return make(map[int64]string)
	}

	// 使用模型定义
	var images []model.HouseImage
	err := r.data.db.Debug(). // 添加Debug()以输出SQL语句
					Table("house_image").
					Select("house_id, image_url").
					Where("house_id IN ?", houseIDs).
					Where("sort_order = 0"). // 获取第一张图片
					Scan(&images).Error

	imageMap := make(map[int64]string)
	if err != nil {
		fmt.Printf("获取房源图片失败: %v\n", err)
		// 如果获取失败，返回默认图片
		for _, houseID := range houseIDs {
			imageMap[houseID] = "https://example.com/default-house.jpg"
		}
		return imageMap
	}

	fmt.Printf("成功获取到 %d 张房源图片\n", len(images))

	// 构建图片映射
	for _, img := range images {
		imageMap[img.HouseID] = img.ImageURL
		fmt.Printf("房源 %d 的图片: %s\n", img.HouseID, img.ImageURL)
	}

	// 为没有图片的房源设置默认图片
	for _, houseID := range houseIDs {
		if _, exists := imageMap[houseID]; !exists {
			imageMap[houseID] = "https://example.com/default-house.jpg"
			fmt.Printf("房源 %d 没有图片，使用默认图片\n", houseID)
		}
	}

	return imageMap
}

// 获取默认房源数据
func getDefaultHouses() []*biz.House {
	return []*biz.House{
		{
			HouseID:     101,
			Title:       "精装修两室一厅",
			Description: "地铁口附近，交通便利，精装修",
			Price:       3500.0,
			Area:        85.5,
			Layout:      "2室1厅1卫",
			ImageURL:    "https://example.com/house1.jpg",
		},
		{
			HouseID:     102,
			Title:       "温馨三室两厅",
			Description: "小区环境优美，配套设施完善",
			Price:       4200.0,
			Area:        120.0,
			Layout:      "3室2厅2卫",
			ImageURL:    "https://example.com/house2.jpg",
		},
		{
			HouseID:     103,
			Title:       "豪华公寓",
			Description: "高端小区，装修豪华，设施齐全",
			Price:       5800.0,
			Area:        150.0,
			Layout:      "3室2厅2卫",
			ImageURL:    "https://example.com/house3.jpg",
		},
	}
}

// 查询推荐房源
func (r *houseRepo) GetRecommendList(ctx context.Context, page, pageSize int) ([]*biz.House, int, error) {
	var houses []*biz.House
	var total int64

	// 添加详细日志
	fmt.Printf("开始查询推荐房源: page=%d, pageSize=%d\n", page, pageSize)

	// 查询总数
	err := r.data.db.
		Table("house").
		Where("status = 'active'").
		Count(&total).Error

	if err != nil {
		fmt.Printf("查询推荐房源总数失败: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("查询到符合条件的房源总数: %d\n", total)

	// 使用已有的模型定义
	var results []model.House

	// 查询房源列表
	err = r.data.db.Debug(). // 添加Debug()以输出SQL语句
					Table("house").
					Select("house_id, title, description, price, area, layout").
					Where("status = 'active'").
					Order("house_id DESC").
					Limit(pageSize).
					Offset((page - 1) * pageSize).
					Scan(&results).Error

	if err != nil {
		fmt.Printf("查询推荐房源列表失败: %v\n", err)
		return nil, 0, err
	}

	// 如果没有数据，返回默认数据
	if len(results) == 0 {
		fmt.Printf("未找到符合条件的房源，返回默认数据\n")
		return getDefaultHouses(), 3, nil
	}

	// 获取房源ID列表
	houseIDs := make([]int64, len(results))
	for i, result := range results {
		houseIDs[i] = result.HouseId
	}

	// 批量获取房源图片
	imageMap := r.getHouseImages(houseIDs)

	// 转换为业务层结构体
	houses = make([]*biz.House, 0, len(results))
	for _, result := range results {
		imageURL := imageMap[result.HouseId]

		houses = append(houses, &biz.House{
			HouseID:     result.HouseId,
			Title:       result.Title,
			Description: result.Description,
			Price:       result.Price,
			Area:        float64(result.Area),
			Layout:      result.Layout,
			ImageURL:    imageURL,
		})
	}

	fmt.Printf("成功查询到 %d 条推荐房源\n", len(houses))
	for i, house := range houses {
		fmt.Printf("房源 %d: ID=%d, 标题=%s, 价格=%.2f\n", i+1, house.HouseID, house.Title, house.Price)
	}

	return houses, int(total), nil
}

// 获取房源的第一张图片
func (r *houseRepo) getHouseFirstImage(houseID int64) string {
	var imageURL string
	err := r.data.db.
		Table("house_image").
		Select("image_url").
		Where("house_id = ?", houseID).
		Order("sort_order ASC").
		Limit(1).
		Scan(&imageURL).Error

	if err != nil || imageURL == "" {
		// 如果没有找到图片，返回默认图片URL
		return "https://example.com/default-house.jpg"
	}

	return imageURL
}

// 批量获取房源图片
// 预约记录模型

// 保存预约
func (r *houseRepo) CreateReservation(ctx context.Context, reservation *model.HouseReservation) error {
	return r.data.db.WithContext(ctx).Create(reservation).Error
}

// 查询是否已预约
func (r *houseRepo) HasReservation(ctx context.Context, userID, houseID int64) (bool, error) {
	var count int64
	err := r.data.db.WithContext(ctx).
		Model(&model.HouseReservation{}).
		Where("user_id = ? AND house_id = ?", userID, houseID).
		Count(&count).Error
	return count > 0, err
}
