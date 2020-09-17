/**
 * @File: param
 * @Author: Shaw
 * @Date: 2020/5/5 5:26 PM
 * @Desc

 */

package routers

import (
	"GZHU-Pi/env"
	"GZHU-Pi/services/minapp"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
	"net/http"
	"time"
)

func Param(w http.ResponseWriter, r *http.Request) {
	paramType := r.URL.Query().Get("type")
	if paramType == "app" {
		getAppParam(w, r)
		return
	}
	if paramType == "modal" {
		modalParam(w, r)
		return
	}
	modalParam(w, r)
}

func modalParam(w http.ResponseWriter, r *http.Request) {

	data, err := getModalParam()
	if err != nil || data == nil {
		Response(w, r, nil, http.StatusOK, fmt.Sprint(err))
		return
	}

	var user env.TUser
	if len(r.Cookies()) == 0 {
		Response(w, r, nil, http.StatusOK, "no cookie")
		return
	}
	user.ID, err = GetUserID(r)
	if err != nil || user.ID <= 0 {
		Response(w, r, nil, http.StatusUnauthorized, fmt.Sprintf("%v", err))
		return
	}

	open, ok := data["open"]
	o, ok := open.(bool)
	if !ok || !o {
		Response(w, r, nil, http.StatusOK, "close")
		return
	}

	key := fmt.Sprintf("gzhupi:param:action:user:%d", user.ID)
	_, err = env.RedisCli.Get(key).Result()
	if err != nil && err != redis.Nil {
		logs.Error(err)
		Response(w, r, nil, http.StatusUnauthorized, err.Error())
		return
	}

	if err == redis.Nil {
		var store = map[string]interface{}{
			"timestamp": time.Now().Unix() * 1000,
			"data":      data,
			"times":     1,
		}
		js, err := json.Marshal(store)
		if err != nil {
			logs.Error(err)
			Response(w, r, nil, http.StatusInternalServerError, err.Error())
			return
		}
		err = env.RedisCli.Set(key, string(js), 0).Err()
		if err != nil {
			Response(w, r, nil, http.StatusInternalServerError, err.Error())
			logs.Error(err)
			return
		}
	}

	if err == redis.Nil {
		//第一次才响应数据
		Response(w, r, data, http.StatusOK, "request ok")
	} else {
		//已经有记录了，响应空数据
		Response(w, r, nil, http.StatusOK, "request ok")
	}
}

func getModalParam() (data map[string]interface{}, err error) {
	key := fmt.Sprintf("gzhupi:param:action")
	val, err := env.RedisCli.Get(key).Result()
	if err != nil && err != redis.Nil {
		logs.Error(err)
		return
	}
	if err == redis.Nil {
		//example := map[string]interface{}{
		//	"open":   false,
		//	"modal": map[string]interface{}{
		//		"valid":   false,
		//		"cancel":  false,
		//		"confirm": false,
		//
		//		"img":       "",
		//		"title":     "活动推荐",
		//		"sub_title": "义工",
		//		"content":   []string{"经历了一个漫长的假期,你一定想出去走走，外面的世界那么大"},
		//		"btn_text":  "我想去看看",
		//		"nav_to":    "/pages/Setting/webview/webview?src=https://mp.weixin.qq.com/s/NbVEBpvlPgpfbAf_tRh09w",
		//	},
		//}
		//js, _ := json.Marshal(example)
		//err = env.RedisCli.Set(key, string(js), 0).Err()
		//if err != nil {
		//	logs.Error(err)
		//	return
		//}
		//logs.Warn("action param is not set")
		return
	}
	data = make(map[string]interface{})
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		logs.Error(err)
		return
	}

	return
}

func getAppParam(w http.ResponseWriter, r *http.Request) {

	minApp, err := minapp.GetMinApp()
	if err != nil {
		Response(w, r, nil, http.StatusUnauthorized, err.Error())
		return
	}

	var gs = &env.CacheOptions{
		Key:      fmt.Sprintf("gzhupi:minapp:param"),
		Duration: 1 * time.Minute,
		Receiver: make(map[string]interface{}),
		Fun: func() (interface{}, error) {
			return minApp.GetTableData("78887", "5d4daa727b9e3c65e7983f54")
		},
	}
	_, err = env.GetSetCache(gs)
	if err != nil {
		logs.Error(err)
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}

	var data = gs.Receiver.(map[string]interface{})

	Response(w, r, data, http.StatusOK, "request ok")
}
