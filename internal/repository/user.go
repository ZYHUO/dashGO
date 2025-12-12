package repository

import (
	"time"
	"dashgo/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id int64) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Plan").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByToken(token string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Plan").Where("token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUUID(uuid string) (*model.User, error) {
	var user model.User
	err := r.db.Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAvailableUsers 获取指定权限组的可用用户（带流控检查）
func (r *UserRepository) GetAvailableUsers(groupIDs []int64) ([]model.User, error) {
	var users []model.User
	now := getCurrentTimestamp()
	
	query := r.db.Where("banned = ?", false)
	
	// 如果指定了用户组，则过滤
	if len(groupIDs) > 0 {
		query = query.Where("group_id IN ?", groupIDs)
	}
	
	// 流控检查：
	// 1. transfer_enable = 0 表示无限流量
	// 2. u + d < transfer_enable 表示还有剩余流量
	query = query.Where("(transfer_enable = 0 OR u + d < transfer_enable)")
	
	// 过期检�?
	query = query.Where("(expired_at IS NULL OR expired_at = 0 OR expired_at >= ?)", now)
	
	// 只选择必要的字�?
	err := query.Select("id", "uuid", "speed_limit", "device_limit", "u", "d", "transfer_enable").Find(&users).Error
	return users, err
}

// GetAllAvailableUsers 获取所有可用用户（不限制组，带流控检查）
func (r *UserRepository) GetAllAvailableUsers() ([]model.User, error) {
	var users []model.User
	now := getCurrentTimestamp()
	err := r.db.
		Where("banned = ?", false).
		Where("(transfer_enable = 0 OR u + d < transfer_enable)"). // 流量�?表示无限�?
		Where("(expired_at IS NULL OR expired_at = 0 OR expired_at >= ?)", now).
		Select("id", "uuid", "speed_limit", "device_limit", "u", "d", "transfer_enable").
		Find(&users).Error
	return users, err
}

// UpdateTraffic 更新用户流量
func (r *UserRepository) UpdateTraffic(userID int64, u, d int64) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"u": gorm.Expr("u + ?", u),
			"d": gorm.Expr("d + ?", d),
			"t": getCurrentTimestamp(),
		}).Error
}

// BatchUpdateTraffic 批量更新用户流量
func (r *UserRepository) BatchUpdateTraffic(trafficData map[int64][2]int64) error {
	tx := r.db.Begin()
	for userID, traffic := range trafficData {
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"u": gorm.Expr("u + ?", traffic[0]),
				"d": gorm.Expr("d + ?", traffic[1]),
				"t": getCurrentTimestamp(),
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *UserRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) CountByPlanID(planID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("plan_id = ?", planID).Count(&count).Error
	return count, err
}

// Count 统计用户总数
func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

// CountActive 统计活跃用户�?
func (r *UserRepository) CountActive() (int64, error) {
	var count int64
	now := time.Now().Unix()
	err := r.db.Model(&model.User{}).
		Where("banned = ?", false).
		Where("(expired_at >= ? OR expired_at IS NULL OR expired_at = 0)", now).
		Where("plan_id IS NOT NULL").
		Count(&count).Error
	return count, err
}

// CountOnline 统计在线用户�?
func (r *UserRepository) CountOnline(seconds int64) (int64, error) {
	var count int64
	threshold := time.Now().Unix() - seconds
	err := r.db.Model(&model.User{}).Where("t >= ?", threshold).Count(&count).Error
	return count, err
}

// FindAll 查询所有用户（支持搜索和分页）
func (r *UserRepository) FindAll(search string, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})
	if search != "" {
		query = query.Where("email LIKE ?", "%"+search+"%")
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// CountByInviteUserID 统计被邀请用户数
func (r *UserRepository) CountByInviteUserID(inviteUserID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("invite_user_id = ?", inviteUserID).Count(&count).Error
	return count, err
}

// GetUsersNeedTrafficReset 获取需要重置流量的用户
func (r *UserRepository) GetUsersNeedTrafficReset() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("plan_id IS NOT NULL").
		Where("plan_id > 0").
		Where("(u > 0 OR d > 0)").
		Find(&users).Error
	return users, err
}

// GetUsersExpiringSoon 获取即将过期的用�?
func (r *UserRepository) GetUsersExpiringSoon(days int) ([]model.User, error) {
	var users []model.User
	now := time.Now().Unix()
	threshold := now + int64(days*86400)
	err := r.db.Where("expired_at > ?", now).
		Where("expired_at <= ?", threshold).
		Where("banned = ?", false).
		Find(&users).Error
	return users, err
}

// GetUsersWithHighTrafficUsage 获取流量使用率高的用�?
func (r *UserRepository) GetUsersWithHighTrafficUsage(percentage int) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("transfer_enable > 0").
		Where("(u + d) * 100 / transfer_enable >= ?", percentage).
		Where("banned = ?", false).
		Find(&users).Error
	return users, err
}

// CountByDateRange 统计指定日期范围内的新用户数
func (r *UserRepository) CountByDateRange(startTime, endTime int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Where("created_at >= ?", startTime).
		Where("created_at < ?", endTime).
		Count(&count).Error
	return count, err
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}


// FindByTelegramID 根据 Telegram ID 查找用户
func (r *UserRepository) FindByTelegramID(telegramID int64) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Plan").Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersPaginated 获取用户列表（分页）
func (r *UserRepository) GetUsersPaginated(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}


// CountByRegisterIP 统计 IP 注册数量
func (r *UserRepository) CountByRegisterIP(ip string) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("register_ip = ?", ip).Count(&count).Error
	return count, err
}


// FindByUUIDPrefix 根据 UUID 前缀查找用户
func (r *UserRepository) FindByUUIDPrefix(prefix string) (*model.User, error) {
	var user model.User
	err := r.db.Where("uuid LIKE ?", prefix+"%").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
