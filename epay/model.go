package epay

import (
	"fmt"
	"net/url"
)

type EpaySignedRequest interface {
	GetSign() string
	GetSignType() string
}

type EpaySubmitRequest struct {
	Pid        int    `json:"pid" form:"pid"`
	Type       string `json:"type" form:"type"`
	OutTradeNo string `json:"out_trade_no" form:"out_trade_no"`
	NotifyUrl  string `json:"notify_url" form:"notify_url"`
	ReturnUrl  string `json:"return_url" form:"return_url"`
	Name       string `json:"name" form:"name"`
	Money      string `json:"money" form:"money"`
	Param      string `json:"param" form:"param"`
	Device     string `json:"device" form:"device"`
	Sign       string `json:"sign" form:"sign"`
	SignType   string `json:"sign_type" form:"sign_type"`
}

// GetSign returns the sign value
func (r *EpaySubmitRequest) GetSign() string {
	return r.Sign
}

// GetSignType returns the sign type
func (r *EpaySubmitRequest) GetSignType() string {
	return r.SignType
}

// ToURLValues converts the EpaySubmitRequest to url.Values
func (r *EpaySubmitRequest) ToURLValues() url.Values {
	values := url.Values{}

	values.Add("pid", fmt.Sprintf("%d", r.Pid))
	values.Add("type", r.Type)
	values.Add("out_trade_no", r.OutTradeNo)
	values.Add("notify_url", r.NotifyUrl)
	values.Add("return_url", r.ReturnUrl)
	values.Add("name", r.Name)
	values.Add("money", r.Money)

	// Only add param if it's not empty
	if r.Param != "" {
		values.Add("param", r.Param)
	}

	values.Add("sign", r.Sign)
	values.Add("sign_type", r.SignType)

	return values
}

// EpayNotifyRequest represents the notification data sent by the payment system
type EpayNotifyRequest struct {
	Pid         int    `json:"pid" query:"pid"`                   // 商户ID
	TradeNo     string `json:"trade_no" query:"trade_no"`         // 易支付订单号
	OutTradeNo  string `json:"out_trade_no" query:"out_trade_no"` // 商户订单号
	Type        string `json:"type" query:"type"`                 // 支付方式
	Name        string `json:"name" query:"name"`                 // 商品名称
	Money       string `json:"money" query:"money"`               // 商品金额
	TradeStatus string `json:"trade_status" query:"trade_status"` // 支付状态
	Param       string `json:"param" query:"param"`               // 业务扩展参数
	Sign        string `json:"sign" query:"sign"`                 // 签名字符串
	SignType    string `json:"sign_type" query:"sign_type"`       // 签名类型
}

// GetSign returns the sign value
func (r *EpayNotifyRequest) GetSign() string {
	return r.Sign
}

// GetSignType returns the sign type
func (r *EpayNotifyRequest) GetSignType() string {
	return r.SignType
}

// ToURLValues converts the EpayNotifyRequest to url.Values
func (r *EpayNotifyRequest) ToURLValues() url.Values {
	values := url.Values{}

	values.Add("pid", fmt.Sprintf("%d", r.Pid))
	values.Add("trade_no", r.TradeNo)
	values.Add("out_trade_no", r.OutTradeNo)
	values.Add("type", r.Type)
	values.Add("name", r.Name)
	values.Add("money", r.Money)
	values.Add("trade_status", r.TradeStatus)

	// Only add param if it's not empty
	if r.Param != "" {
		values.Add("param", r.Param)
	}

	values.Add("sign", r.Sign)
	values.Add("sign_type", r.SignType)

	return values
}
