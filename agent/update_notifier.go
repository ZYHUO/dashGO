package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UpdateStatus 更新状�?
type UpdateStatus string

const (
	// UpdateStatusSuccess 更新成功
	UpdateStatusSuccess UpdateStatus = "success"
	// UpdateStatusFailed 更新失败
	UpdateStatusFailed UpdateStatus = "failed"
	// UpdateStatusRollback 已回�?
	UpdateStatusRollback UpdateStatus = "rollback"
)

// UpdateNotification 更新通知
type UpdateNotification struct {
	Status       UpdateStatus `json:"status"`
	FromVersion  string       `json:"from_version"`
	ToVersion    string       `json:"to_version"`
	ErrorMessage string       `json:"error_message,omitempty"`
	Timestamp    time.Time    `json:"timestamp"`
}

// UpdateNotifier 更新通知�?
type UpdateNotifier struct {
	panelURL string
	token    string
	client   *http.Client
}

// NewUpdateNotifier 创建更新通知�?
func NewUpdateNotifier(panelURL, token string) *UpdateNotifier {
	return &UpdateNotifier{
		panelURL: panelURL,
		token:    token,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// NotifySuccess 发送更新成功通知
func (un *UpdateNotifier) NotifySuccess(fromVersion, toVersion string) error {
	notification := UpdateNotification{
		Status:      UpdateStatusSuccess,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Timestamp:   time.Now(),
	}

	fmt.Printf("📤 发送更新成功通知: %s -> %s\n", fromVersion, toVersion)
	
	if err := un.sendNotification(notification); err != nil {
		fmt.Printf("�?发送成功通知失败: %v\n", err)
		return err
	}

	fmt.Println("�?成功通知已发�?)
	return nil
}

// NotifyFailure 发送更新失败告�?
func (un *UpdateNotifier) NotifyFailure(fromVersion, toVersion string, err error) error {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	notification := UpdateNotification{
		Status:       UpdateStatusFailed,
		FromVersion:  fromVersion,
		ToVersion:    toVersion,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now(),
	}

	fmt.Printf("📤 发送更新失败告�? %s -> %s\n", fromVersion, toVersion)
	fmt.Printf("   错误: %s\n", errorMessage)
	
	if err := un.sendNotification(notification); err != nil {
		fmt.Printf("�?发送失败告警失�? %v\n", err)
		return err
	}

	fmt.Println("�?失败告警已发�?)
	return nil
}

// NotifyRollback 发送回滚通知
func (un *UpdateNotifier) NotifyRollback(fromVersion, toVersion string, err error) error {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	notification := UpdateNotification{
		Status:       UpdateStatusRollback,
		FromVersion:  fromVersion,
		ToVersion:    toVersion,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now(),
	}

	fmt.Printf("📤 发送回滚通知: %s -> %s\n", fromVersion, toVersion)
	fmt.Printf("   原因: %s\n", errorMessage)
	
	if err := un.sendNotification(notification); err != nil {
		fmt.Printf("�?发送回滚通知失败: %v\n", err)
		return err
	}

	fmt.Println("�?回滚通知已发�?)
	return nil
}

// sendNotification 发送通知�?Panel
func (un *UpdateNotifier) sendNotification(notification UpdateNotification) error {
	url := un.panelURL + "/api/v1/agent/update-status"

	// 序列化通知数据
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", un.token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请�?
	resp, err := un.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状�?
	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if errMsg, ok := result["error"].(string); ok {
				return fmt.Errorf("server error: %s", errMsg)
			}
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
