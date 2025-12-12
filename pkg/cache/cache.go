package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"dashgo/internal/config"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	rdb *redis.Client
	ctx context.Context
}

func New(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb, ctx: ctx}, nil
}

func (c *Client) Get(key string) (string, error) {
	return c.rdb.Get(c.ctx, key).Result()
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(c.ctx, key, value, expiration).Err()
}

func (c *Client) Del(key string) error {
	return c.rdb.Del(c.ctx, key).Err()
}

func (c *Client) Exists(key string) (bool, error) {
	n, err := c.rdb.Exists(c.ctx, key).Result()
	return n > 0, err
}

func (c *Client) Incr(key string) (int64, error) {
	return c.rdb.Incr(c.ctx, key).Result()
}

func (c *Client) IncrBy(key string, value int64) (int64, error) {
	return c.rdb.IncrBy(c.ctx, key, value).Result()
}

func (c *Client) HGet(key, field string) (string, error) {
	return c.rdb.HGet(c.ctx, key, field).Result()
}

func (c *Client) HSet(key, field string, value interface{}) error {
	return c.rdb.HSet(c.ctx, key, field, value).Err()
}

func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.rdb.HGetAll(c.ctx, key).Result()
}

func (c *Client) HDel(key string, fields ...string) error {
	return c.rdb.HDel(c.ctx, key, fields...).Err()
}

func (c *Client) Expire(key string, expiration time.Duration) error {
	return c.rdb.Expire(c.ctx, key, expiration).Err()
}

func (c *Client) TTL(key string) (time.Duration, error) {
	return c.rdb.TTL(c.ctx, key).Result()
}

// Cache key prefixes
const (
	KeyServerLastCheckAt = "SERVER_%s_LAST_CHECK_AT_%d"
	KeyServerLastPushAt  = "SERVER_%s_LAST_PUSH_AT_%d"
	KeyServerOnlineUser  = "SERVER_%s_ONLINE_USER_%d"
	KeyServerLoadStatus  = "SERVER_%s_LOAD_STATUS_%d"
	KeyUserOnline        = "USER_ONLINE_%d"

	// Agent 缓存
	KeyAgentConfig    = "AGENT_CONFIG_%d"        // 主机配置缓存
	KeyAgentUsersHash = "AGENT_USERS_HASH_%d"    // 用户列表哈希
	KeyNodeUsers      = "NODE_USERS_%d"          // 节点用户列表
	KeyUserListVersion = "USER_LIST_VERSION"     // 用户列表版本�?

	// 订阅缓存
	KeySubscription     = "SUBSCRIPTION_%d_%s"   // 用户订阅缓存
	KeySubscriptionHash = "SUB_HASH_%d"          // 用户订阅哈希

	// 用户缓存
	KeyUserInfo       = "USER_INFO_%d"           // 用户信息缓存
	KeyUserList       = "USER_LIST_PAGE_%d_%d"   // 用户列表分页缓存
	KeyUserListTotal  = "USER_LIST_TOTAL"        // 用户总数缓存
	KeyUserChanges    = "USER_CHANGES"           // 用户变更列表
	KeyUserChangeVer  = "USER_CHANGE_VERSION"    // 用户变更版本

	// 节点用户缓存
	KeyNodeUserList    = "NODE_USER_LIST_%d"     // 节点用户列表
	KeyNodeUserHash    = "NODE_USER_HASH_%d"     // 节点用户哈希
	KeyNodeUserVersion = "NODE_USER_VERSION_%d"  // 节点用户版本

	// 站点设置缓存
	KeySiteSettings = "SITE_SETTINGS"            // 站点设置
	KeySiteSetting  = "SITE_SETTING_%s"          // 单个设置
)

func ServerLastCheckAtKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerLastCheckAt, serverType, serverID)
}

func ServerLastPushAtKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerLastPushAt, serverType, serverID)
}

func ServerOnlineUserKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerOnlineUser, serverType, serverID)
}

func ServerLoadStatusKey(serverType string, serverID int64) string {
	return fmt.Sprintf(KeyServerLoadStatus, serverType, serverID)
}

func AgentConfigKey(hostID int64) string {
	return fmt.Sprintf(KeyAgentConfig, hostID)
}

func AgentUsersHashKey(hostID int64) string {
	return fmt.Sprintf(KeyAgentUsersHash, hostID)
}

func NodeUsersKey(nodeID int64) string {
	return fmt.Sprintf(KeyNodeUsers, nodeID)
}

func SubscriptionKey(userID int64, format string) string {
	return fmt.Sprintf(KeySubscription, userID, format)
}

func SubscriptionHashKey(userID int64) string {
	return fmt.Sprintf(KeySubscriptionHash, userID)
}

func UserInfoKey(userID int64) string {
	return fmt.Sprintf(KeyUserInfo, userID)
}

func UserListPageKey(page, pageSize int) string {
	return fmt.Sprintf(KeyUserList, page, pageSize)
}

func NodeUserListKey(nodeID int64) string {
	return fmt.Sprintf(KeyNodeUserList, nodeID)
}

func NodeUserHashKey(nodeID int64) string {
	return fmt.Sprintf(KeyNodeUserHash, nodeID)
}

func NodeUserVersionKey(nodeID int64) string {
	return fmt.Sprintf(KeyNodeUserVersion, nodeID)
}

func SiteSettingKey(key string) string {
	return fmt.Sprintf(KeySiteSetting, key)
}

// SetJSON 设置 JSON �?
func (c *Client) SetJSON(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(key, string(data), expiration)
}

// GetJSON 获取 JSON �?
func (c *Client) GetJSON(key string, dest interface{}) error {
	val, err := c.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// DelPattern 删除匹配模式的所有键
func (c *Client) DelPattern(pattern string) error {
	keys, err := c.rdb.Keys(c.ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.rdb.Del(c.ctx, keys...).Err()
	}
	return nil
}

// GetUserListVersion 获取用户列表版本
func (c *Client) GetUserListVersion() (int64, error) {
	val, err := c.Get(KeyUserListVersion)
	if err != nil {
		return 0, err
	}
	var version int64
	fmt.Sscanf(val, "%d", &version)
	return version, nil
}

// IncrUserListVersion 增加用户列表版本
func (c *Client) IncrUserListVersion() (int64, error) {
	return c.Incr(KeyUserListVersion)
}

// GetNodeUserVersion 获取节点用户版本
func (c *Client) GetNodeUserVersion(nodeID int64) (int64, error) {
	val, err := c.Get(NodeUserVersionKey(nodeID))
	if err != nil {
		return 0, err
	}
	var version int64
	fmt.Sscanf(val, "%d", &version)
	return version, nil
}

// IncrNodeUserVersion 增加节点用户版本
func (c *Client) IncrNodeUserVersion(nodeID int64) (int64, error) {
	return c.Incr(NodeUserVersionKey(nodeID))
}

// RecordUserChange 记录用户变更
func (c *Client) RecordUserChange(userID int64, changeType string) error {
	change := map[string]interface{}{
		"user_id": userID,
		"type":    changeType,
		"time":    time.Now().Unix(),
	}
	data, _ := json.Marshal(change)
	return c.rdb.RPush(c.ctx, KeyUserChanges, string(data)).Err()
}

// GetUserChanges 获取用户变更列表
func (c *Client) GetUserChanges(limit int64) ([]map[string]interface{}, error) {
	vals, err := c.rdb.LRange(c.ctx, KeyUserChanges, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	changes := make([]map[string]interface{}, 0, len(vals))
	for _, val := range vals {
		var change map[string]interface{}
		if err := json.Unmarshal([]byte(val), &change); err == nil {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

// ClearUserChanges 清除已处理的用户变更
func (c *Client) ClearUserChanges(count int64) error {
	return c.rdb.LTrim(c.ctx, KeyUserChanges, count, -1).Err()
}

// ComputeHash 计算数据哈希
func ComputeHash(data interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:])
}

// SetWithLock 带锁设置（防止并发）
func (c *Client) SetWithLock(key string, value interface{}, expiration time.Duration) (bool, error) {
	lockKey := "lock:" + key
	ok, err := c.rdb.SetNX(c.ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil || !ok {
		return false, err
	}
	defer c.rdb.Del(c.ctx, lockKey)

	return true, c.Set(key, value, expiration)
}

// SAdd 添加集合成员
func (c *Client) SAdd(key string, members ...interface{}) error {
	return c.rdb.SAdd(c.ctx, key, members...).Err()
}

// SMembers 获取集合所有成�?
func (c *Client) SMembers(key string) ([]string, error) {
	return c.rdb.SMembers(c.ctx, key).Result()
}

// SRem 移除集合成员
func (c *Client) SRem(key string, members ...interface{}) error {
	return c.rdb.SRem(c.ctx, key, members...).Err()
}

// SIsMember 检查是否是集合成员
func (c *Client) SIsMember(key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(c.ctx, key, member).Result()
}
