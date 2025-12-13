package sms

import (
	"fmt"
	"log"
)

// SMSClient 短信客户端接口
type SMSClient interface {
	SendCode(phone, code string) error
}

// MockSMSClient 模拟短信客户端（学习用）
type MockSMSClient struct{}

// NewMockSMSClient 创建Mock客户端
func NewMockSMSClient() *MockSMSClient {
	return &MockSMSClient{}
}

// SendCode 发送验证码（打印到控制台）
func (m *MockSMSClient) SendCode(phone, code string) error {
	// 直接打印到控制台，方便学习测试
	fmt.Println("========================================")
	fmt.Printf("📱 短信发送成功！\n")
	fmt.Printf("手机号: %s\n", phone)
	fmt.Printf("验证码: %s\n", code)
	fmt.Printf("提示: 这是模拟短信，验证码已显示在上方\n")
	fmt.Println("========================================")

	// 同时记录日志
	log.Printf("[SMS] 发送验证码到 %s: %s", phone, code)

	return nil
}
