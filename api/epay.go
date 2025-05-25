package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/smartwalle/alipay/v3"
	"github.com/spf13/viper"
	"github.com/yiffyi/epay-fwd/epay"
	"github.com/yiffyi/epay-fwd/sec"
)

func SetupEpayEndpoints(g *echo.Group) {
	log.Info().Msg("Setting up Epay endpoints")
	g.POST("/:env/submit.php", HandleEpaySubmit)
}

func buildAlipayNotifyUrl() (string, error) {
	log.Debug().Str("site_url", viper.GetString("site_url")).Msg("Building Alipay notify URL")
	base, err := url.Parse(viper.GetString("site_url"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse site URL")
		return "", err
	}

	base.Path = "/alipay/notify"
	notifyUrl := base.String()
	log.Debug().Str("notify_url", notifyUrl).Msg("Built Alipay notify URL")
	return notifyUrl, nil
}

func HandleEpaySubmit(c echo.Context) error {
	env := c.Param("env")
	log.Info().Str("env", env).Msg("Handling Epay submit request")

	isProd := strings.HasPrefix(env, "prod")
	log.Debug().Bool("is_prod", isProd).Msg("Environment check")

	if isProd && !viper.GetBool("alipay.enable_production") {
		log.Warn().Msg("Production environment is disabled but received production request")
		return echo.NewHTTPError(http.StatusForbidden, "production environment is disabled")
	}

	var client, err = alipay.New(viper.GetString("alipay.app_id"), viper.GetString("alipay.app_private_key"), isProd)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Alipay client")
		c.Logger().Error(err)
		return err
	}
	client.LoadAliPayPublicKey(viper.GetString("alipay.server_public_key"))

	var epayParam epay.EpaySubmitRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &epayParam); err != nil {
		log.Error().Err(err).Msg("Failed to bind request body to EpaySubmitRequest")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	log.Debug().
		Int("pid", epayParam.Pid).
		Str("out_trade_no", epayParam.OutTradeNo).
		Str("name", epayParam.Name).
		Str("money", epayParam.Money).
		Msg("Received Epay submit parameters")

	val := epay.NewEpaySignValidator(sec.DeriveMyEpayKey(epayParam.Pid, viper.GetString("epay.fwd_secret")))

	if err := val.Validate(&epayParam); err != nil {
		log.Error().Err(err).Int("pid", epayParam.Pid).Msg("Failed to validate Epay signature")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	log.Debug().Int("pid", epayParam.Pid).Msg("Epay signature validated successfully")

	epayParamCarrier := epay.ParamCarrier{
		Pid:       epayParam.Pid,
		NotifyUrl: epayParam.NotifyUrl,
		Param:     epayParam.Param,
	}

	passbackParams, err := epayParamCarrier.Encode()
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode param carrier")
		return err
	}

	notifyUrl, err := buildAlipayNotifyUrl()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build Alipay notify URL")
		return err
	}

	alipayParam := alipay.TradePagePay{
		Trade: alipay.Trade{
			NotifyURL: notifyUrl,
			ReturnURL: epayParam.ReturnUrl,

			Subject:     epayParam.Name,
			OutTradeNo:  epayParam.OutTradeNo,
			TotalAmount: epayParam.Money,
			ProductCode: "FAST_INSTANT_TRADE_PAY",

			PassbackParams: passbackParams,
		},
	}

	log.Debug().
		Str("notify_url", notifyUrl).
		Str("return_url", epayParam.ReturnUrl).
		Str("out_trade_no", epayParam.OutTradeNo).
		Str("subject", epayParam.Name).
		Str("total_amount", epayParam.Money).
		Msg("Creating Alipay trade page pay request")

	result, err := client.TradePagePay(alipayParam)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Alipay trade page pay")
		return err
	}

	log.Info().
		Str("out_trade_no", epayParam.OutTradeNo).
		Str("redirect_url", result.String()).
		Msg("Redirecting to Alipay payment page")

	// return c.Redirect(http.StatusTemporaryRedirect, result.String())
	// 必须使用 302，否则 epay 的 POST body 会被保留
	return c.Redirect(http.StatusFound, result.String())
}
