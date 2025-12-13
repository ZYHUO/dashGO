package service

import (
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
)

// NoticeService 公告服务
type NoticeService struct {
	noticeRepo *repository.NoticeRepository
}

func NewNoticeService(noticeRepo *repository.NoticeRepository) *NoticeService {
	return &NoticeService{noticeRepo: noticeRepo}
}

// GetAll 获取所有公告
func (s *NoticeService) GetAll() ([]model.Notice, error) {
	return s.noticeRepo.GetAll()
}

// GetVisible 获取可见公告
func (s *NoticeService) GetVisible() ([]model.Notice, error) {
	return s.noticeRepo.GetVisible()
}

// GetPublic 获取公开公告（用于前端展示）
func (s *NoticeService) GetPublic() ([]map[string]interface{}, error) {
	notices, err := s.noticeRepo.GetVisible()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(notices))
	for _, notice := range notices {
		result = append(result, map[string]interface{}{
			"id":         notice.ID,
			"title":      notice.Title,
			"content":    notice.Content,
			"img_url":    notice.ImgURL,
			"created_at": notice.CreatedAt,
		})
	}
	return result, nil
}

// GetByID 根据 ID 获取公告
func (s *NoticeService) GetByID(id int64) (*model.Notice, error) {
	return s.noticeRepo.FindByID(id)
}

// Create 创建公告
func (s *NoticeService) Create(notice *model.Notice) error {
	notice.CreatedAt = time.Now().Unix()
	notice.UpdatedAt = time.Now().Unix()
	return s.noticeRepo.Create(notice)
}

// Update 更新公告
func (s *NoticeService) Update(notice *model.Notice) error {
	notice.UpdatedAt = time.Now().Unix()
	return s.noticeRepo.Update(notice)
}

// Delete 删除公告
func (s *NoticeService) Delete(id int64) error {
	return s.noticeRepo.Delete(id)
}

// KnowledgeService 知识库服务
type KnowledgeService struct {
	knowledgeRepo *repository.KnowledgeRepository
}

func NewKnowledgeService(knowledgeRepo *repository.KnowledgeRepository) *KnowledgeService {
	return &KnowledgeService{knowledgeRepo: knowledgeRepo}
}

// GetAll 获取所有知识库文章
func (s *KnowledgeService) GetAll() ([]model.Knowledge, error) {
	return s.knowledgeRepo.GetAll()
}

// GetVisible 获取可见文章
func (s *KnowledgeService) GetVisible(language string) ([]model.Knowledge, error) {
	return s.knowledgeRepo.GetVisible(language)
}

// GetByCategory 按分类获取文告
func (s *KnowledgeService) GetByCategory(category, language string) ([]model.Knowledge, error) {
	return s.knowledgeRepo.GetByCategory(category, language)
}

// GetByID 根据 ID 获取文章
func (s *KnowledgeService) GetByID(id int64) (*model.Knowledge, error) {
	return s.knowledgeRepo.FindByID(id)
}

// Create 创建文章
func (s *KnowledgeService) Create(knowledge *model.Knowledge) error {
	knowledge.CreatedAt = time.Now().Unix()
	knowledge.UpdatedAt = time.Now().Unix()
	return s.knowledgeRepo.Create(knowledge)
}

// Update 更新文章
func (s *KnowledgeService) Update(knowledge *model.Knowledge) error {
	knowledge.UpdatedAt = time.Now().Unix()
	return s.knowledgeRepo.Update(knowledge)
}

// Delete 删除文章
func (s *KnowledgeService) Delete(id int64) error {
	return s.knowledgeRepo.Delete(id)
}

// GetCategories 获取所有分告
func (s *KnowledgeService) GetCategories() ([]string, error) {
	return s.knowledgeRepo.GetCategories("")
}

// GetPublic 获取公开知识库文告
func (s *KnowledgeService) GetPublic(category string) ([]map[string]interface{}, error) {
	var items []model.Knowledge
	var err error

	if category != "" {
		items, err = s.knowledgeRepo.GetByCategory(category, "")
	} else {
		items, err = s.knowledgeRepo.GetVisible("")
	}

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]interface{}{
			"id":       item.ID,
			"title":    item.Title,
			"body":     item.Body,
			"category": item.Category,
			"language": item.Language,
		})
	}
	return result, nil
}
