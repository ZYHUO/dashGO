package service

import (
	"errors"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
	"dashgo/pkg/cache"
	"dashgo/pkg/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
	cache    *cache.Client
}

func NewUserService(userRepo *repository.UserRepository, cache *cache.Client) *UserService {
	return &UserService{
		userRepo: userRepo,
		cache:    cache,
	}
}

// GetUsersCached 获取用户列表（带缓存�?
func (s *UserService) GetUsersCached(page, pageSize int) ([]model.User, int64, error) {
	cacheKey := cache.UserListPageKey(page, pageSize)

	// 尝试从缓存获�?
	var cachedResult struct {
		Users []model.User `json:"users"`
		Total int64        `json:"total"`
	}
	if err := s.cache.GetJSON(cacheKey, &cachedResult); err == nil {
		return cachedResult.Users, cachedResult.Total, nil
	}

	// 从数据库获取
	users, total, err := s.userRepo.GetUsersPaginated(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存�?分钟�?
	cachedResult.Users = users
	cachedResult.Total = total
	s.cache.SetJSON(cacheKey, cachedResult, 5*time.Minute)

	return users, total, nil
}

// InvalidateUserCache 使用户缓存失�?
func (s *UserService) InvalidateUserCache(userID int64) {
	// 删除用户信息缓存
	s.cache.Del(cache.UserInfoKey(userID))
	// 删除用户列表缓存
	s.cache.DelPattern("USER_LIST_PAGE_*")
	// 记录用户变更
	s.cache.RecordUserChange(userID, "update")
	// 增加版本�?
	s.cache.IncrUserListVersion()
}

// InvalidateUserListCache 使用户列表缓存失�?
func (s *UserService) InvalidateUserListCache() {
	s.cache.DelPattern("USER_LIST_PAGE_*")
	s.cache.Del(cache.KeyUserListTotal)
}

// GetUserByIDCached 获取用户（带缓存�?
func (s *UserService) GetUserByIDCached(userID int64) (*model.User, error) {
	cacheKey := cache.UserInfoKey(userID)

	var user model.User
	if err := s.cache.GetJSON(cacheKey, &user); err == nil {
		return &user, nil
	}

	// 从数据库获取
	dbUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存�?0分钟�?
	s.cache.SetJSON(cacheKey, dbUser, 10*time.Minute)
	return dbUser, nil
}

// Register 用户注册
func (s *UserService) Register(email, password string, inviteUserID *int64) (*model.User, error) {
	// 检查邮箱是否已存在
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		Password:     hashedPassword,
		UUID:         utils.GenerateUUID(),
		Token:        utils.GenerateToken(32),
		InviteUserID: inviteUserID,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(email, password string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	if user.Banned {
		return nil, errors.New("account is banned")
	}

	// 更新最后登录时�?
	now := time.Now().Unix()
	user.LastLoginAt = &now
	s.userRepo.Update(user)

	return user, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id int64) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// GetByToken 根据 Token 获取用户
func (s *UserService) GetByToken(token string) (*model.User, error) {
	return s.userRepo.FindByToken(token)
}

// GetByUUID 根据 UUID 获取用户
func (s *UserService) GetByUUID(uuid string) (*model.User, error) {
	return s.userRepo.FindByUUID(uuid)
}

// GetByUUIDPrefix 根据 UUID 前缀获取用户
func (s *UserService) GetByUUIDPrefix(prefix string) (*model.User, error) {
	return s.userRepo.FindByUUIDPrefix(prefix)
}

// UpdateTraffic 更新用户流量
func (s *UserService) UpdateTraffic(userID int64, u, d int64) error {
	return s.userRepo.UpdateTraffic(userID, u, d)
}

// TrafficFetch 批量处理流量上报（带流控检查）
func (s *UserService) TrafficFetch(server *model.Server, trafficData map[int64][2]int64) error {
	// 计算倍率
	rate := server.Rate
	if rate <= 0 {
		rate = 1
	}

	// 应用倍率并检查流量限�?
	for userID, traffic := range trafficData {
		u := int64(float64(traffic[0]) * rate)
		d := int64(float64(traffic[1]) * rate)
		
		// 更新流量
		if err := s.userRepo.UpdateTraffic(userID, u, d); err != nil {
			continue
		}
		
		// 检查用户是否超流量
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			continue
		}
		
		// 流控检查：如果超流量，标记用户（可选：自动封禁或只是记录）
		if user.TransferEnable > 0 && (user.U + user.D) >= user.TransferEnable {
			// 超流量处理：这里可以选择自动封禁或发送通知
			// 方案1：自动封禁（激进）
			// user.Banned = true
			// s.userRepo.Update(user)
			
			// 方案2：只记录日志（温和）
			// 下次 GetAvailableUsers 时会自动过滤�?
			
			// 使缓存失效，确保下次拉取用户列表时不包含该用�?
			s.InvalidateUserCache(userID)
		}
	}

	return nil
}

// ResetToken 重置用户 Token
func (s *UserService) ResetToken(userID int64) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}

	user.Token = utils.GenerateToken(32)
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return user.Token, nil
}

// ResetUUID 重置用户 UUID
func (s *UserService) ResetUUID(userID int64) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}

	user.UUID = utils.GenerateUUID()
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return user.UUID, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return s.userRepo.Update(user)
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(user *model.User) map[string]interface{} {
	info := map[string]interface{}{
		"id":              user.ID,
		"email":           user.Email,
		"uuid":            user.UUID,
		"token":           user.Token,
		"balance":         user.Balance,
		"plan_id":         user.PlanID,
		"group_id":        user.GroupID,
		"transfer_enable": user.TransferEnable,
		"u":               user.U,
		"d":               user.D,
		"expired_at":      user.ExpiredAt,
		"is_admin":        user.IsAdmin,
		"is_staff":        user.IsStaff,
		"created_at":      user.CreatedAt,
	}

	// 添加套餐信息
	if user.Plan != nil {
		info["plan"] = map[string]interface{}{
			"id":   user.Plan.ID,
			"name": user.Plan.Name,
		}
	}

	return info
}


// GetNodeUsersWithCache 获取节点用户（带缓存和增量同步支持）
func (s *UserService) GetNodeUsersWithCache(nodeID int64, groupIDs []int64, lastVersion int64) (*NodeUsersResult, error) {
	cacheKey := cache.NodeUserListKey(nodeID)
	hashKey := cache.NodeUserHashKey(nodeID)

	// 获取当前版本
	currentVersion, _ := s.cache.GetNodeUserVersion(nodeID)

	// 如果客户端版本与当前版本相同，返回空（无变化�?
	if lastVersion > 0 && lastVersion == currentVersion {
		return &NodeUsersResult{
			Version:  currentVersion,
			HasChange: false,
		}, nil
	}

	// 尝试从缓存获�?
	var users []NodeUserInfo
	if err := s.cache.GetJSON(cacheKey, &users); err == nil {
		// 检查哈希是否变�?
		cachedHash, _ := s.cache.Get(hashKey)
		currentHash := cache.ComputeHash(users)
		if cachedHash == currentHash {
			return &NodeUsersResult{
				Version:   currentVersion,
				Users:     users,
				Hash:      currentHash,
				HasChange: lastVersion != currentVersion,
			}, nil
		}
	}

	// 从数据库获取
	dbUsers, err := s.userRepo.GetAvailableUsers(groupIDs)
	if err != nil {
		return nil, err
	}

	users = make([]NodeUserInfo, 0, len(dbUsers))
	for _, u := range dbUsers {
		users = append(users, NodeUserInfo{
			ID:          u.ID,
			UUID:        u.UUID,
			SpeedLimit:  u.SpeedLimit,
			DeviceLimit: u.DeviceLimit,
		})
	}

	// 计算哈希
	newHash := cache.ComputeHash(users)

	// 更新缓存
	s.cache.SetJSON(cacheKey, users, 5*time.Minute)
	s.cache.Set(hashKey, newHash, 5*time.Minute)

	// 如果哈希变化，增加版本号
	oldHash, _ := s.cache.Get(hashKey)
	if oldHash != newHash {
		currentVersion, _ = s.cache.IncrNodeUserVersion(nodeID)
	}

	return &NodeUsersResult{
		Version:   currentVersion,
		Users:     users,
		Hash:      newHash,
		HasChange: true,
	}, nil
}

// NodeUserInfo 节点用户信息
type NodeUserInfo struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  *int   `json:"speed_limit,omitempty"`
	DeviceLimit *int   `json:"device_limit,omitempty"`
}

// NodeUsersResult 节点用户结果
type NodeUsersResult struct {
	Version   int64          `json:"version"`
	Users     []NodeUserInfo `json:"users,omitempty"`
	Hash      string         `json:"hash"`
	HasChange bool           `json:"has_change"`
}

// GetChangedUsers 获取变更的用户（增量同步�?
func (s *UserService) GetChangedUsers(sinceVersion int64) ([]int64, int64, error) {
	currentVersion, _ := s.cache.GetUserListVersion()
	if sinceVersion >= currentVersion {
		return nil, currentVersion, nil
	}

	// 获取变更记录
	changes, err := s.cache.GetUserChanges(1000)
	if err != nil {
		return nil, currentVersion, err
	}

	userIDs := make([]int64, 0)
	seen := make(map[int64]bool)
	for _, change := range changes {
		if uid, ok := change["user_id"].(float64); ok {
			id := int64(uid)
			if !seen[id] {
				userIDs = append(userIDs, id)
				seen[id] = true
			}
		}
	}

	return userIDs, currentVersion, nil
}


// GetByEmail 根据邮箱获取用户
func (s *UserService) GetByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

// RegisterWithIP �?IP 记录的用户注�?
func (s *UserService) RegisterWithIP(email, password string, inviteUserID *int64, ip string) (*model.User, error) {
	// 检查邮箱是否已存在
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		Password:     hashedPassword,
		UUID:         utils.GenerateUUID(),
		Token:        utils.GenerateToken(32),
		InviteUserID: inviteUserID,
		RegisterIP:   &ip,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// CountByRegisterIP 统计 IP 注册数量
func (s *UserService) CountByRegisterIP(ip string) (int64, error) {
	return s.userRepo.CountByRegisterIP(ip)
}

// SendEmailCode 发送邮箱验证码
func (s *UserService) SendEmailCode(email string) error {
	// 生成 6 位验证码
	code := utils.GenerateNumericCode(6)

	// 存储到缓存（10分钟有效�?
	cacheKey := "email_code:" + email
	if err := s.cache.Set(cacheKey, code, 10*time.Minute); err != nil {
		return err
	}

	// 这里需要调用邮件服务发送验证码
	// 由于 UserService 没有直接引用 MailService，需要通过其他方式
	// 实际实现中可以通过事件或者在 handler 层处�?
	return nil
}

// VerifyEmailCode 验证邮箱验证�?
func (s *UserService) VerifyEmailCode(email, code string) bool {
	cacheKey := "email_code:" + email
	storedCode, err := s.cache.Get(cacheKey)
	if err != nil {
		return false
	}
	if storedCode == code {
		// 验证成功后删除验证码
		s.cache.Del(cacheKey)
		return true
	}
	return false
}

// SetEmailCode 设置邮箱验证码（供外部调用）
func (s *UserService) SetEmailCode(email, code string) error {
	cacheKey := "email_code:" + email
	return s.cache.Set(cacheKey, code, 10*time.Minute)
}

// GetEmailCodeCooldown 获取验证码冷却时�?
func (s *UserService) GetEmailCodeCooldown(email string) int64 {
	cacheKey := "email_code_cooldown:" + email
	ttl, err := s.cache.TTL(cacheKey)
	if err != nil || ttl < 0 {
		return 0
	}
	return int64(ttl.Seconds())
}

// SetEmailCodeCooldown 设置验证码冷�?
func (s *UserService) SetEmailCodeCooldown(email string) error {
	cacheKey := "email_code_cooldown:" + email
	return s.cache.Set(cacheKey, "1", 60*time.Second)
}
