/**
 * @File: taoke
 * @Author: Shaw
 * @Date: 2020/10/6 2:47 PM
 * @Desc

 */

package taoke

import (
	"GZHU-Pi/env"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type TaoKe struct {
	ApKey  string
	TbName string
	PID    string
	Client *http.Client
}

var tk *TaoKe
var one sync.Once

func NewTaoKe() *TaoKe {
	if tk != nil {
		return tk
	}
	one.Do(func() {

		tk = &TaoKe{
			ApKey:  env.Conf.TaoKe.ApKey,
			TbName: env.Conf.TaoKe.TbName,
			PID:    env.Conf.TaoKe.PID,
			Client: &http.Client{},
		}
	})
	return tk
}

type TaoResp struct {
	HasCoupon  bool   `json:"has_coupon"`
	Tpwd       string `json:"tpwd"`        //口令
	TpwdStr    string `json:"tpwd_str"`    //口令描述
	QuanLimit  string `json:"quanlimit"`   //启用价
	YouHuiQuan string `json:"youhuiquan"`  //券面额
	CouponInfo string `json:"coupon_info"` //券描述
}

func (t *TaoKe) ConvertToken(content string) (text string, err error) {
	resp, err := t.Convert(content)
	if err != nil {
		logs.Error(err)
		return "没有找到信息", nil
	}
	fmt.Println(resp)

	tpl1 := `为您找到优惠券 【%v】
一一一一一一一一一一一一
【满减价】：￥%v元
【券面额】：￥%v元
一一一一一一一一一一一一
长按复制此整条信息
%v
一一一一一一一一一一一一`
	tpl2 := `没有找到优惠信息~
一一一一一一一一一一一一
【满减价】：——
【券面额】：——
一一一一一一一一一一一一
长按复制此整条信息
%v
一一一一一一一一一一一一`
	if resp.HasCoupon {
		text = fmt.Sprintf(tpl1, resp.CouponInfo, resp.QuanLimit, resp.YouHuiQuan, resp.TpwdStr)
	} else if resp.TpwdStr != "" {
		text = fmt.Sprintf(tpl2, resp.TpwdStr)
	} else {
		text = fmt.Sprintf("该商品没有找到优惠信息 \n\n%s", content)
	}
	return
}

func (t *TaoKe) Convert(content string) (data TaoResp, err error) {
	form := map[string]interface{}{
		"apkey":    t.ApKey,
		"pid":      t.PID,
		"tbname":   t.TbName,
		"tpwd":     1, //返回口令
		"shorturl": 1,
		"content":  content,
	}
	urlData := url.Values{
		"apkey":     []string{t.ApKey},
		"pid":       []string{t.PID},
		"tbname":    []string{t.TbName},
		"tpwd":      []string{"1"}, //返回口令
		"shorturl":  []string{"1"},
		"extsearch": []string{"1"},
		"content":   []string{content},
	}

	apiUrl := "http://api.web.21ds.cn/taoke/doItemHighCommissionPromotionLinkByAll?" + urlData.Encode()

	body, _ := json.Marshal(&form)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		logs.Error(err)
		return
	}
	resp, err := t.Client.Do(req)
	if err != nil {
		logs.Error(err)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	var ApiData struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data TaoResp `json:"data"`
	}

	if err = json.Unmarshal(body, &ApiData); err != nil {
		logs.Error(err)
		return
	}
	data = ApiData.Data
	return
}
