/**
 * @File: exp
 * @Author: Shaw
 * @Date: 2020/9/22 12:20 AM
 * @Desc

 */

package routers

import (
	"github.com/astaxie/beego/logs"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Exp(w http.ResponseWriter, r *http.Request) {
	trueServer := "http://127.0.0.1:5000"

	u, err := url.Parse(trueServer)
	if err != nil {
		logs.Error(err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}
