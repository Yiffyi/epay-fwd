package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/yiffyi/epay-fwd/epay"
	"github.com/yiffyi/epay-fwd/misc"
	"github.com/yiffyi/epay-fwd/sec"
)

func main() {
	misc.SetupConfig()

	param := epay.EpaySubmitRequest{
		Pid:        123456,
		Type:       "alipay",
		OutTradeNo: time.Now().Format("SANDBOX20060102150405"),
		NotifyUrl:  "https://localhost/a",
		ReturnUrl:  "https://localhost/a",
		Name:       "Test Product",
		Money:      "1.00",
		Param:      "test",
		SignType:   "MD5",
	}

	epayKey := sec.DeriveMyEpayKey(param.Pid, viper.GetString("epay.fwd_secret"))
	fmt.Println(epayKey)

	var err error
	param.Sign, err = epay.CalculateSign(&param, epayKey)
	if err != nil {
		panic(err)
	}

	fmt.Println(param.Sign)

	// Send the GET request
	resp, err := http.Post("http://127.0.0.1:1323/epay/test/submit.php", "application/x-www-form-urlencoded", strings.NewReader(param.ToURLValues().Encode()))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// bytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(bytes))
	fmt.Println(resp.Request.URL)
}
