package api

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/smartwalle/alipay/v3"
	"github.com/spf13/viper"
	"github.com/yiffyi/epay-fwd/epay"
	"github.com/yiffyi/epay-fwd/sec"
)

func SetupAlipayEndpoints(g *echo.Group) {
	log.Info().Msg("Setting up Alipay endpoints")
	g.POST("/notify", HandleAlipayNotify)
	// g.POST("/return", HandleAlipayReturn)
}

func HandleAlipayReturn(c echo.Context) error {
	log.Info().Msg("Handling Alipay return request")
	values := c.QueryParams()

	client, err := alipay.New(viper.GetString("alipay.app_id"), viper.GetString("alipay.app_private_key"), true)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Alipay client")
		return err
	}

	if err := client.VerifySign(values); err != nil {
		log.Error().Err(err).Msg("Failed to verify Alipay signature")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	log.Debug().Msg("Alipay signature verified successfully")

	urlBytes, err := base64.RawURLEncoding.DecodeString(c.Param("realUrl"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode URL from parameter")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	redirectUrl := string(urlBytes)
	log.Info().Str("redirect_url", redirectUrl).Msg("Redirecting user after Alipay return")

	// TODO: no arbitrary redirect
	// TODO: check and include order status

	return c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func HandleAlipayNotify(c echo.Context) error {
	log.Info().Msg("Handling Alipay notification")

	values, err := c.FormParams()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get form parameters")
		return err
	}

	client, err := alipay.New(viper.GetString("alipay.app_id"), viper.GetString("alipay.app_private_key"), true)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Alipay client")
		return err
	}
	client.LoadAliPayPublicKey(viper.GetString("alipay.server_public_key"))

	// DecodeNotification 内部已调用 VerifySign 方法验证签名
	notify, err := client.DecodeNotification(values)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode Alipay notification")
		return err
	}

	log.Debug().
		Str("trade_no", notify.TradeNo).
		Str("out_trade_no", notify.OutTradeNo).
		Str("trade_status", string(notify.TradeStatus)).
		Str("total_amount", notify.TotalAmount).
		Msg("Received Alipay notification")

	tradeStatus := "UNKNOWN"
	if notify.TradeStatus == "TRADE_SUCCESS" || notify.TradeStatus == "TRADE_FINISHED" {
		tradeStatus = "TRADE_SUCCESS"
	} else {
		tradeStatus = string(notify.TradeStatus)
	}
	log.Debug().Str("normalized_trade_status", tradeStatus).Msg("Normalized trade status")

	epayParamCarrier, err := epay.DecodeParamCarrier(notify.PassbackParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode param carrier from passback params")
		return err
	}

	log.Debug().
		Int("pid", epayParamCarrier.Pid).
		Str("notify_url", epayParamCarrier.NotifyUrl).
		Str("param", epayParamCarrier.Param).
		Msg("Decoded param carrier")

	// Create the EpayNotifyRequest
	epayNotify := epay.EpayNotifyRequest{
		Pid:         epayParamCarrier.Pid,
		TradeNo:     notify.TradeNo,
		OutTradeNo:  notify.OutTradeNo,
		Type:        "alipay",
		Name:        notify.Subject,
		Money:       notify.TotalAmount,
		TradeStatus: tradeStatus,
		Param:       epayParamCarrier.Param,
		SignType:    "MD5",
	}

	log.Debug().
		Int("pid", epayNotify.Pid).
		Str("trade_no", epayNotify.TradeNo).
		Str("out_trade_no", epayNotify.OutTradeNo).
		Str("type", epayNotify.Type).
		Str("name", epayNotify.Name).
		Str("money", epayNotify.Money).
		Str("trade_status", epayNotify.TradeStatus).
		Msg("Created Epay notify request")

	// Calculate the sign
	merchantKey := sec.DeriveMyEpayKey(epayParamCarrier.Pid, viper.GetString("epay.fwd_secret"))
	epayNotify.Sign, err = epay.CalculateSign(&epayNotify, merchantKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate sign for Epay notify request")
		return err
	}

	// Parse the base URL
	notifyURL, err := url.Parse(epayParamCarrier.NotifyUrl)
	if err != nil {
		log.Error().Err(err).Str("notify_url", epayParamCarrier.NotifyUrl).Msg("Failed to parse notify URL")
		return err
	}

	// Set the query parameters
	notifyURL.RawQuery = epayNotify.ToURLValues().Encode()

	fullNotifyURL := notifyURL.String()
	log.Info().Str("notify_url", fullNotifyURL).Msg("Sending notification to merchant")

	// Send the GET request
	resp, err := http.Get(fullNotifyURL)
	if err != nil {
		log.Error().Err(err).Str("notify_url", fullNotifyURL).Msg("Failed to send GET request to notify URL")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("notify_url", fullNotifyURL).
			Msg("Received non-OK response from merchant notify URL")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send notification")
	}

	log.Info().
		Int("status_code", resp.StatusCode).
		Str("out_trade_no", notify.OutTradeNo).
		Msg("Successfully forwarded notification to merchant")

	// alipay.ACKNotification(c.Response().Writer)
	return c.Stream(resp.StatusCode, resp.Header.Get("Content-Type"), resp.Body)
}
