package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewUpdateNotifier(t *testing.T) {
	panelURL := "https://panel.example.com"
	token := "test-token"

	notifier := NewUpdateNotifier(panelURL, token)

	if notifier.panelURL != panelURL {
		t.Errorf("panelURL = %v, want %v", notifier.panelURL, panelURL)
	}
	if notifier.token != token {
		t.Errorf("token = %v, want %v", notifier.token, token)
	}
	if notifier.client == nil {
		t.Error("client should not be nil")
	}
}

func TestUpdateNotifier_NotifySuccess(t *testing.T) {
	// 创建测试服务告
	var receivedNotification UpdateNotification
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}

		// 验证请求路径
		if r.URL.Path != "/api/v1/agent/update-status" {
			t.Errorf("Path = %v, want /api/v1/agent/update-status", r.URL.Path)
		}

		// 验证请求告
		if r.Header.Get("Authorization") != "test-token" {
			t.Errorf("Authorization = %v, want test-token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", r.Header.Get("Content-Type"))
		}

		// 解析请求告
		if err := json.NewDecoder(r.Body).Decode(&receivedNotification); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	// 创建通知告
	notifier := NewUpdateNotifier(server.URL, "test-token")

	// 发送成功通知
	err := notifier.NotifySuccess("v1.0.0", "v1.1.0")
	if err != nil {
		t.Errorf("NotifySuccess() error = %v", err)
	}

	// 验证接收到的通知
	if receivedNotification.Status != UpdateStatusSuccess {
		t.Errorf("Status = %v, want %v", receivedNotification.Status, UpdateStatusSuccess)
	}
	if receivedNotification.FromVersion != "v1.0.0" {
		t.Errorf("FromVersion = %v, want v1.0.0", receivedNotification.FromVersion)
	}
	if receivedNotification.ToVersion != "v1.1.0" {
		t.Errorf("ToVersion = %v, want v1.1.0", receivedNotification.ToVersion)
	}
	if receivedNotification.ErrorMessage != "" {
		t.Errorf("ErrorMessage should be empty for success notification")
	}
}

func TestUpdateNotifier_NotifyFailure(t *testing.T) {
	// 创建测试服务告
	var receivedNotification UpdateNotification
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedNotification)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	// 创建通知告
	notifier := NewUpdateNotifier(server.URL, "test-token")

	// 发送失败通知
	testErr := errors.New("download failed")
	err := notifier.NotifyFailure("v1.0.0", "v1.1.0", testErr)
	if err != nil {
		t.Errorf("NotifyFailure() error = %v", err)
	}

	// 验证接收到的通知
	if receivedNotification.Status != UpdateStatusFailed {
		t.Errorf("Status = %v, want %v", receivedNotification.Status, UpdateStatusFailed)
	}
	if receivedNotification.FromVersion != "v1.0.0" {
		t.Errorf("FromVersion = %v, want v1.0.0", receivedNotification.FromVersion)
	}
	if receivedNotification.ToVersion != "v1.1.0" {
		t.Errorf("ToVersion = %v, want v1.1.0", receivedNotification.ToVersion)
	}
	if receivedNotification.ErrorMessage != "download failed" {
		t.Errorf("ErrorMessage = %v, want 'download failed'", receivedNotification.ErrorMessage)
	}
}

func TestUpdateNotifier_NotifyRollback(t *testing.T) {
	// 创建测试服务告
	var receivedNotification UpdateNotification
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedNotification)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	// 创建通知告
	notifier := NewUpdateNotifier(server.URL, "test-token")

	// 发送回滚通知
	testErr := errors.New("replace failed")
	err := notifier.NotifyRollback("v1.0.0", "v1.1.0", testErr)
	if err != nil {
		t.Errorf("NotifyRollback() error = %v", err)
	}

	// 验证接收到的通知
	if receivedNotification.Status != UpdateStatusRollback {
		t.Errorf("Status = %v, want %v", receivedNotification.Status, UpdateStatusRollback)
	}
	if receivedNotification.ErrorMessage != "replace failed" {
		t.Errorf("ErrorMessage = %v, want 'replace failed'", receivedNotification.ErrorMessage)
	}
}

func TestUpdateNotifier_ServerError(t *testing.T) {
	// 创建返回错误的测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "internal server error",
		})
	}))
	defer server.Close()

	// 创建通知告
	notifier := NewUpdateNotifier(server.URL, "test-token")

	// 发送通知应该返回错误
	err := notifier.NotifySuccess("v1.0.0", "v1.1.0")
	if err == nil {
		t.Error("NotifySuccess() should return error when server returns error")
	}
}

func TestUpdateNotifier_NetworkError(t *testing.T) {
	// 使用无效告URL
	notifier := NewUpdateNotifier("http://invalid-url-that-does-not-exist.local", "test-token")

	// 发送通知应该返回网络错误
	err := notifier.NotifySuccess("v1.0.0", "v1.1.0")
	if err == nil {
		t.Error("NotifySuccess() should return error when network fails")
	}
}

func TestUpdateNotification_Timestamp(t *testing.T) {
	// 创建测试服务告
	var receivedNotification UpdateNotification
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedNotification)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	// 创建通知告
	notifier := NewUpdateNotifier(server.URL, "test-token")

	// 记录发送前的时告
	beforeSend := time.Now()

	// 发送通知
	notifier.NotifySuccess("v1.0.0", "v1.1.0")

	// 记录发送后的时告
	afterSend := time.Now()

	// 验证时间戳在合理范围告
	if receivedNotification.Timestamp.Before(beforeSend) || receivedNotification.Timestamp.After(afterSend) {
		t.Errorf("Timestamp %v is not between %v and %v", 
			receivedNotification.Timestamp, beforeSend, afterSend)
	}
}

func TestUpdateStatus_Constants(t *testing.T) {
	// Test that all status constants are defined correctly
	statuses := []UpdateStatus{
		UpdateStatusSuccess,
		UpdateStatusFailed,
		UpdateStatusRollback,
	}

	expectedValues := []string{
		"success",
		"failed",
		"rollback",
	}

	for i, status := range statuses {
		if string(status) != expectedValues[i] {
			t.Errorf("Status %d = %v, want %v", i, status, expectedValues[i])
		}
	}
}

func TestUpdateNotifier_NotifyWithNilError(t *testing.T) {
	// 创建测试服务告
	var receivedNotification UpdateNotification
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedNotification)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	// 创建通知告
	notifier := NewUpdateNotifier(server.URL, "test-token")

	// 发送失败通知，但错误告nil
	err := notifier.NotifyFailure("v1.0.0", "v1.1.0", nil)
	if err != nil {
		t.Errorf("NotifyFailure() error = %v", err)
	}

	// 验证错误消息为空
	if receivedNotification.ErrorMessage != "" {
		t.Errorf("ErrorMessage should be empty when error is nil, got %v", receivedNotification.ErrorMessage)
	}
}
