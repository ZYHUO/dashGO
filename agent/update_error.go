package main

import (
	"fmt"
)

// UpdateErrorCategory 更新错误类别
type UpdateErrorCategory string

const (
	// ErrorCategoryNetwork 网络错误
	ErrorCategoryNetwork UpdateErrorCategory = "network"
	// ErrorCategoryFile 文件错误
	ErrorCategoryFile UpdateErrorCategory = "file"
	// ErrorCategoryVerification 验证错误
	ErrorCategoryVerification UpdateErrorCategory = "verification"
	// ErrorCategoryUpdate 更新错误
	ErrorCategoryUpdate UpdateErrorCategory = "update"
	// ErrorCategoryUnknown 未知错误
	ErrorCategoryUnknown UpdateErrorCategory = "unknown"
)

// UpdateError 更新错误类型
type UpdateError struct {
	Category  UpdateErrorCategory
	Message   string
	Retryable bool
	Err       error // 原始错误
}

// Error 实现 error 接口
func (e *UpdateError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Category, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// Unwrap 支持 errors.Unwrap
func (e *UpdateError) Unwrap() error {
	return e.Err
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, err error) *UpdateError {
	return &UpdateError{
		Category:  ErrorCategoryNetwork,
		Message:   message,
		Retryable: true,
		Err:       err,
	}
}

// NewFileError 创建文件错误
func NewFileError(message string, err error) *UpdateError {
	return &UpdateError{
		Category:  ErrorCategoryFile,
		Message:   message,
		Retryable: false,
		Err:       err,
	}
}

// NewVerificationError 创建验证错误
func NewVerificationError(message string, err error) *UpdateError {
	return &UpdateError{
		Category:  ErrorCategoryVerification,
		Message:   message,
		Retryable: false,
		Err:       err,
	}
}

// NewUpdateError 创建更新错误
func NewUpdateError(message string, err error) *UpdateError {
	return &UpdateError{
		Category:  ErrorCategoryUpdate,
		Message:   message,
		Retryable: false,
		Err:       err,
	}
}

// HandleError 统一错误处理函数
// 根据错误类型决定是否重试，并记录详细日志
func HandleError(err error) bool {
	if err == nil {
		return false
	}

	// 尝试转换�?UpdateError
	if updateErr, ok := err.(*UpdateError); ok {
		// 记录详细错误信息
		fmt.Printf("�?更新错误 [%s]: %s\n", updateErr.Category, updateErr.Message)
		if updateErr.Err != nil {
			fmt.Printf("   详细信息: %v\n", updateErr.Err)
		}

		// 根据错误类别提供建议
		switch updateErr.Category {
		case ErrorCategoryNetwork:
			fmt.Println("   建议: 检查网络连接，稍后会自动重�?)
		case ErrorCategoryFile:
			fmt.Println("   建议: 检查磁盘空间和文件权限")
		case ErrorCategoryVerification:
			fmt.Println("   建议: 文件可能已损坏，请联系管理员")
		case ErrorCategoryUpdate:
			fmt.Println("   建议: 更新过程失败，已自动回滚到原版本")
		}

		return updateErr.Retryable
	}

	// 未知错误类型
	fmt.Printf("�?未知错误: %v\n", err)
	return false
}
