// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     gaowenbin(gaowenbinmarr@gmail.com)

package card

import (
	"github.com/chanxuehong/wechat/mp"
	"net/http"
)

type Client struct {
	mp.WechatClient
}

// 创建一个新的 Client.
//  如果 HttpClient == nil 则默认用 http.DefaultClient
func NewClient(TokenServer mp.TokenServer, HttpClient *http.Client) *Client {
	if TokenServer == nil {
		panic("TokenServer == nil")
	}
	if HttpClient == nil {
		HttpClient = http.DefaultClient
	}

	return &Client{
		WechatClient: mp.WechatClient{
			TokenServer: TokenServer,
			HttpClient:  HttpClient,
		},
	}
}