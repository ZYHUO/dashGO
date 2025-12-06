package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"xboard/internal/model"
	"xboard/internal/repository"
)

// PaymentService 支付服务
type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	orderRepo   *repository.OrderRepository
	orderSvc    *OrderService
}

func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	orderRepo *repository.OrderRepository,
	orderSvc *OrderService,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		orderSvc:    orderSvc,
	}
}

// PaymentResult 支付结果
type PaymentResult struct {
	Type      string `json:"type"`       // redirect, qrcode
	Data      string `json:"data"`       // URL or QR code content
	PaymentID int64  `json:"payment_id"`
}

// GetEnabledPayments 获取启用的支付方式
func (s *PaymentService) GetEnabledPayments() ([]model.Payment, error) {
	return s.paymentRepo.GetEnabled()
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(tradeNo string, paymentID int64) (*PaymentResult, error) {
	order, err := s.orderRepo.FindByTradeNo(tradeNo)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.Status != model.OrderStatusPending {
		return nil, errors.New("order is not pending")
	}

	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return nil, errors.New("payment method not found")
	}

	if !payment.Enable {
		return nil, errors.New("payment method is disabled")
	}

	// 更新订单支付方式
	order.PaymentID = &paymentID
	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}

	// 根据支付方式创建支付
	var config map[string]string
	json.Unmarshal([]byte(payment.Config), &config)

	switch payment.Payment {
	case "epay":
		return s.createEpayPayment(order, payment, config)
	case "stripe":
		return s.createStripePayment(order, payment, config)
	case "alipay":
		return s.createAlipayPayment(order, payment, config)
	default:
		return nil, errors.New("unsupported payment method")
	}
}

// HandleCallback 处理支付回调
func (s *PaymentService) HandleCallback(paymentUUID string, params map[string]string) error {
	payment, err := s.paymentRepo.FindByUUID(paymentUUID)
	if err != nil {
		return errors.New("payment not found")
	}

	var config map[string]string
	json.Unmarshal([]byte(payment.Config), &config)

	// 验证签名
	switch payment.Payment {
	case "epay":
		if !s.verifyEpaySign(params, config["key"]) {
			return errors.New("invalid signature")
		}
	}

	// 获取订单号
	tradeNo := params["out_trade_no"]
	if tradeNo == "" {
		return errors.New("trade_no not found")
	}

	// 完成订单
	callbackNo := params["trade_no"]
	return s.orderSvc.CompleteOrder(tradeNo, callbackNo)
}

// createEpayPayment 创建易支付
func (s *PaymentService) createEpayPayment(order *model.Order, payment *model.Payment, config map[string]string) (*PaymentResult, error) {
	apiURL := config["url"]
	pid := config["pid"]
	key := config["key"]

	notifyURL := config["notify_url"]
	if payment.NotifyDomain != nil && *payment.NotifyDomain != "" {
		notifyURL = *payment.NotifyDomain + "/api/v1/payment/notify/" + payment.UUID
	}

	params := map[string]string{
		"pid":          pid,
		"type":         "alipay",
		"out_trade_no": order.TradeNo,
		"notify_url":   notifyURL,
		"return_url":   config["return_url"],
		"name":         "订阅服务",
		"money":        fmt.Sprintf("%.2f", float64(order.TotalAmount)/100),
	}

	// 生成签名
	params["sign"] = s.generateEpaySign(params, key)
	params["sign_type"] = "MD5"

	// 构建支付 URL
	payURL := apiURL + "/submit.php?" + s.buildQuery(params)

	return &PaymentResult{
		Type:      "redirect",
		Data:      payURL,
		PaymentID: payment.ID,
	}, nil
}

// createStripePayment 创建 Stripe 支付
func (s *PaymentService) createStripePayment(order *model.Order, payment *model.Payment, config map[string]string) (*PaymentResult, error) {
	// TODO: 实现 Stripe 支付
	return nil, errors.New("stripe payment not implemented")
}

// createAlipayPayment 创建支付宝支付
func (s *PaymentService) createAlipayPayment(order *model.Order, payment *model.Payment, config map[string]string) (*PaymentResult, error) {
	// TODO: 实现支付宝支付
	return nil, errors.New("alipay payment not implemented")
}

// generateEpaySign 生成易支付签名
func (s *PaymentService) generateEpaySign(params map[string]string, key string) string {
	// 按 key 排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接字符串
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}
	buf.WriteString(key)

	// MD5
	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}

// verifyEpaySign 验证易支付签名
func (s *PaymentService) verifyEpaySign(params map[string]string, key string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	return s.generateEpaySign(params, key) == sign
}

// buildQuery 构建查询字符串
func (s *PaymentService) buildQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// CheckPaymentStatus 检查支付状态（主动查询）
func (s *PaymentService) CheckPaymentStatus(tradeNo string) (bool, error) {
	order, err := s.orderRepo.FindByTradeNo(tradeNo)
	if err != nil {
		return false, err
	}

	if order.Status != model.OrderStatusPending {
		return order.Status == model.OrderStatusCompleted, nil
	}

	if order.PaymentID == nil {
		return false, nil
	}

	payment, err := s.paymentRepo.FindByID(*order.PaymentID)
	if err != nil {
		return false, err
	}

	var config map[string]string
	json.Unmarshal([]byte(payment.Config), &config)

	switch payment.Payment {
	case "epay":
		return s.queryEpayStatus(order, config)
	}

	return false, nil
}

// queryEpayStatus 查询易支付状态
func (s *PaymentService) queryEpayStatus(order *model.Order, config map[string]string) (bool, error) {
	apiURL := config["url"] + "/api.php"
	params := map[string]string{
		"act":          "order",
		"pid":          config["pid"],
		"key":          config["key"],
		"out_trade_no": order.TradeNo,
	}

	resp, err := http.Get(apiURL + "?" + s.buildQuery(params))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if status, ok := result["status"].(float64); ok && status == 1 {
		// 支付成功，完成订单
		callbackNo := ""
		if tn, ok := result["trade_no"].(string); ok {
			callbackNo = tn
		}
		s.orderSvc.CompleteOrder(order.TradeNo, callbackNo)
		return true, nil
	}

	return false, nil
}
