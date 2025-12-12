package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"dashgo/internal/config"
	"dashgo/internal/model"
)

// MailService 邮件服务
type MailService struct {
	config MailConfig
}

func NewMailService(cfg config.MailConfig) *MailService {
	return &MailService{
		config: MailConfig{
			Host:       cfg.Host,
			Port:       cfg.Port,
			Username:   cfg.Username,
			Password:   cfg.Password,
			FromName:   cfg.FromName,
			FromEmail:  cfg.FromAddr,
			Encryption: cfg.Encryption,
		},
	}
}

// MailConfig 邮件配置
type MailConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromName   string
	FromEmail  string
	Encryption string // tls, ssl, none
}

// GetConfig 获取邮件配置
func (s *MailService) GetConfig() *MailConfig {
	return &s.config
}

// SendMail 发送邮�?
func (s *MailService) SendMail(to, subject, body string) error {
	cfg := s.GetConfig()

	if cfg.Host == "" {
		return fmt.Errorf("mail not configured")
	}

	from := cfg.FromEmail
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail)
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	if cfg.Encryption == "tls" || cfg.Encryption == "ssl" {
		return s.sendMailTLS(addr, auth, cfg.FromEmail, to, msg)
	}

	return smtp.SendMail(addr, auth, cfg.FromEmail, []string{to}, msg)
}

func (s *MailService) sendMailTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}
	defer conn.Close()

	host := strings.Split(addr, ":")[0]
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// SendVerifyCode 发送验证码
func (s *MailService) SendVerifyCode(to, code string) error {
	subject := "验证�?
	body := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="color: #1a1a2e; margin-bottom: 20px;">验证�?/h2>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">您的验证码是�?/p>
			<div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; font-size: 32px; font-weight: bold; padding: 20px 40px; border-radius: 12px; display: inline-block; margin: 20px 0;">
				%s
			</div>
			<p style="color: #999; font-size: 14px; margin-top: 20px;">验证码有效期�?10 分钟，请勿泄露给他人�?/p>
		</div>
	`, code)
	return s.SendMail(to, subject, body)
}

// SendWelcome 发送欢迎邮�?
func (s *MailService) SendWelcome(user *model.User) error {
	subject := "欢迎注册"
	body := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="color: #1a1a2e; margin-bottom: 20px;">🎉 欢迎加入</h2>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">您好，感谢您的注册！</p>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">您的账号�?strong>%s</strong></p>
			<div style="margin-top: 30px;">
				<a href="#" style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 12px 30px; border-radius: 8px; text-decoration: none; font-weight: 500;">开始使�?/a>
			</div>
		</div>
	`, user.Email)
	return s.SendMail(user.Email, subject, body)
}

// SendExpireReminder 发送到期提�?
func (s *MailService) SendExpireReminder(user *model.User, daysLeft int) error {
	subject := "订阅即将到期提醒"
	body := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="color: #1a1a2e; margin-bottom: 20px;">�?订阅即将到期</h2>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">您好，您的订阅将�?<strong style="color: #e74c3c;">%d �?/strong>后到期�?/p>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">为避免服务中断，请及时续费�?/p>
			<div style="margin-top: 30px;">
				<a href="#" style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 12px 30px; border-radius: 8px; text-decoration: none; font-weight: 500;">立即续费</a>
			</div>
		</div>
	`, daysLeft)
	return s.SendMail(user.Email, subject, body)
}

// SendTrafficWarning 发送流量预�?
func (s *MailService) SendTrafficWarning(user *model.User, usedPercent int) error {
	subject := "流量使用提醒"
	body := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="color: #1a1a2e; margin-bottom: 20px;">📊 流量使用提醒</h2>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">您好，您的流量已使用 <strong style="color: #e74c3c;">%d%%</strong>�?/p>
			<div style="background: #f5f5f5; border-radius: 10px; padding: 4px; margin: 20px 0;">
				<div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); height: 20px; border-radius: 8px; width: %d%%;"></div>
			</div>
			<p style="color: #999; font-size: 14px;">请合理使用流量，或考虑升级套餐�?/p>
		</div>
	`, usedPercent, usedPercent)
	return s.SendMail(user.Email, subject, body)
}

// SendOrderPaid 发送订单支付成功通知
func (s *MailService) SendOrderPaid(user *model.User, order *model.Order) error {
	subject := "订单支付成功"
	body := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="color: #1a1a2e; margin-bottom: 20px;">�?支付成功</h2>
			<p style="color: #666; font-size: 16px; line-height: 1.6;">您的订单已支付成功！</p>
			<div style="background: #f8f9fa; border-radius: 12px; padding: 20px; margin: 20px 0;">
				<p style="margin: 8px 0; color: #666;"><span style="color: #999;">订单号：</span>%s</p>
				<p style="margin: 8px 0; color: #666;"><span style="color: #999;">金额�?/span>¥%.2f</p>
			</div>
			<p style="color: #999; font-size: 14px;">感谢您的支持�?/p>
		</div>
	`, order.TradeNo, float64(order.TotalAmount)/100)
	return s.SendMail(user.Email, subject, body)
}

// MailTemplate 邮件模板
type MailTemplate struct {
	Name    string
	Subject string
	Body    string
}

// RenderTemplate 渲染模板
func (s *MailService) RenderTemplate(tpl *MailTemplate, data interface{}) (string, string, error) {
	subjectTpl, err := template.New("subject").Parse(tpl.Subject)
	if err != nil {
		return "", "", err
	}

	bodyTpl, err := template.New("body").Parse(tpl.Body)
	if err != nil {
		return "", "", err
	}

	var subjectBuf, bodyBuf bytes.Buffer
	if err := subjectTpl.Execute(&subjectBuf, data); err != nil {
		return "", "", err
	}
	if err := bodyTpl.Execute(&bodyBuf, data); err != nil {
		return "", "", err
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}


// IsConfigured 检查邮件是否已配置
func (s *MailService) IsConfigured() bool {
	return s.config.Host != "" && s.config.Username != ""
}
