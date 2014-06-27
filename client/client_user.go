// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat for the canonical source repository
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     chanxuehong@gmail.com

package client

import (
	"errors"
	"fmt"
	"github.com/chanxuehong/wechat/user"
)

// 创建分组
func (c *Client) UserGroupCreate(name string) (*user.Group, error) {
	if len(name) == 0 {
		return nil, errors.New(`name == ""`)
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := userGroupCreateURL(token)

	var request struct {
		Group struct {
			Name string `json:"name"`
		} `json:"group"`
	}

	request.Group.Name = name

	var result struct {
		Group struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"group"`
		Error
	}
	if err = c.postJSON(_url, &request, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}

	var group user.Group
	group.Id = result.Group.Id
	group.Name = result.Group.Name
	return &group, nil
}

// 查询所有分组
func (c *Client) UserGroupGet() ([]user.Group, error) {
	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := userGroupGetURL(token)

	var result = struct {
		Groups []user.Group `json:"groups"`
		Error
	}{
		Groups: make([]user.Group, 0, 64), // GroupCountLimit
	}
	if err = c.getJSON(_url, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}
	return result.Groups, nil
}

// 修改分组名
func (c *Client) UserGroupRename(groupid int, name string) (err error) {
	if len(name) == 0 {
		return errors.New(`name == ""`)
	}

	token, err := c.Token()
	if err != nil {
		return
	}
	_url := userGroupRenameURL(token)

	var request struct {
		Group struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"group"`
	}
	request.Group.Id = groupid
	request.Group.Name = name

	var result Error
	if err = c.postJSON(_url, request, &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		return &result
	}

	return
}

// 查询用户所在分组
func (c *Client) UserInWhichGroup(openid string) (groupid int, err error) {
	if len(openid) == 0 {
		err = errors.New(`openid == ""`)
		return
	}

	token, err := c.Token()
	if err != nil {
		return
	}
	_url := userInWhichGroupURL(token)

	var request = struct {
		OpenId string `json:"openid"`
	}{OpenId: openid}

	var result struct {
		GroupId int `json:"groupid"`
		Error
	}
	if err = c.postJSON(_url, &request, &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = &result.Error
		return
	}

	groupid = result.GroupId
	return
}

// 移动用户分组
func (c *Client) UserMoveToGroup(openid string, toGroupId int) (err error) {
	if len(openid) == 0 {
		return errors.New(`openid == ""`)
	}

	token, err := c.Token()
	if err != nil {
		return
	}
	_url := userMoveToGroupURL(token)

	var request = struct {
		OpenId    string `json:"openid"`
		ToGroupId int    `json:"to_groupid"`
	}{
		OpenId:    openid,
		ToGroupId: toGroupId,
	}

	var result Error
	if err = c.postJSON(_url, &request, &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		return &result
	}

	return
}

// 获取用户基本信息.
//  lang 可能的取值是 zh_CN, zh_TW, en; 如果留空 "" 则默认为 zh_CN.
func (c *Client) UserInfo(openid string, lang string) (*user.UserInfo, error) {
	if len(openid) == 0 {
		return nil, errors.New(`openid == ""`)
	}

	switch lang {
	case "":
		lang = user.Language_zh_CN
	case user.Language_zh_CN, user.Language_zh_TW, user.Language_en:
	default:
		return nil, errors.New(`lang 必须是 "", zh_CN, zh_TW, en 之一`)
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := userInfoURL(token, openid, lang)

	var result struct {
		Subscribe int `json:"subscribe"` // 用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号，拉取不到其余信息。
		user.UserInfo
		Error
	}
	if err = c.getJSON(_url, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}
	if result.Subscribe == 0 {
		return nil, fmt.Errorf("该用户 %s 没有订阅这个公众号", openid)
	}
	return &result.UserInfo, nil
}

// 获取关注者返回的数据结构
type userGetResponse struct {
	TotalCount int `json:"total"` // 关注该公众账号的总用户数
	GetCount   int `json:"count"` // 拉取的OPENID个数，最大值为10000
	Data       struct {
		OpenId []string `json:"openid"`
	} `json:"data"` // 列表数据，OPENID的列表
	// 拉取列表的后一个用户的OPENID, 如果 next_openid == "" 则表示没有了用户数据
	NextOpenId string `json:"next_openid"`
}

// 获取关注者列表, 如果 beginOpenId == "" 则表示从头遍历
func (c *Client) userGet(beginOpenId string) (*userGetResponse, error) {
	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := userGetURL(token, beginOpenId)

	var result struct {
		userGetResponse
		Error
	}
	result.userGetResponse.Data.OpenId = make([]string, 0, user.UserPageCountLimit)
	if err = c.getJSON(_url, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}
	return &result.userGetResponse, nil
}

// 该结构实现了 user.UserIterator 接口
type userGetIterator struct {
	userGetResponse *userGetResponse // 对于 HasNext() 表示上次返回的数据

	wechatClient   *Client // 关联的微信 Client
	nextPageCalled bool    // NextPage() 是否调用过
}

func (iter *userGetIterator) Total() int {
	return iter.userGetResponse.TotalCount
}
func (iter *userGetIterator) HasNext() bool {
	// 第一批数据不需要通过 NextPage() 来获取, 因为在创建这个对象的时候就获取了;
	// 后续的数据都要通过 NextPage() 来获取, 所以要通过上一次的 NextOpenId 来判断了.
	if !iter.nextPageCalled {
		return iter.userGetResponse.GetCount > 0
	}
	return iter.userGetResponse.NextOpenId != ""
}
func (iter *userGetIterator) NextPage() ([]string, error) {
	// 第一次调用 NextPage(), 因为在创建这个对象的时候已经获取了数据, 所以直接返回.
	if !iter.nextPageCalled {
		iter.nextPageCalled = true
		return iter.userGetResponse.Data.OpenId, nil
	}

	// 不是第一次调用的都要从服务器拉取数据
	resp, err := iter.wechatClient.userGet(iter.userGetResponse.NextOpenId)
	if err != nil {
		return nil, err
	}

	iter.userGetResponse = resp // 覆盖老数据
	return resp.Data.OpenId, nil
}

// 关注用户遍历器, 如果 beginOpenId == "" 则表示从头遍历
func (c *Client) UserIterator(beginOpenId string) (user.UserIterator, error) {
	resp, err := c.userGet(beginOpenId)
	if err != nil {
		return nil, err
	}
	var iter userGetIterator
	iter.userGetResponse = resp
	iter.wechatClient = c
	return &iter, nil
}