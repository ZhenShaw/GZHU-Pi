package routers

import (
	"GZHU-Pi/pkg/cet"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func GetPastCetCaptcha(w http.ResponseWriter, r *http.Request) {
	baseUrl := "http://appquery.neea.edu.cn/api/verify/get"
	params := url.Values{
		"t": []string{fmt.Sprintf("%.16f", rand.Float64())},
	}
	fullUrl := baseUrl + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullUrl, nil)
	req.Header = http.Header{
		"Referer":    []string{"http://cjcx.neea.edu.cn/html1/folder/20051/1156-1.htm"},
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36"},
	}
	var curCookies []*http.Cookie = nil
	// var err error;
	curCookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		// Transport:nil,
		// CheckRedirect: nil,
		Jar: curCookieJar,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		Response(w, r, nil, http.StatusBadRequest, err.Error())
		return
	}
	curCookies = httpClient.Jar.Cookies(req.URL)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	for _, v := range curCookies {
		if v.Name == "verify" {
			http.SetCookie(w, v)
		}
	}
	baseImg := fmt.Sprintf("data:%s;base64,%s", resp.Header["Content-Type"][0], base64.StdEncoding.EncodeToString(body))
	Response(w, r, baseImg, http.StatusOK, "request ok")
}

func GetPastCet(w http.ResponseWriter, r *http.Request) {
	captchaCookie, _ := r.Cookie("verify")
	captcha := r.URL.Query().Get("captcha")

	if captcha == "" || captchaCookie == nil {
		err := fmt.Errorf("验证码相关信息缺失")
		Response(w, r, nil, http.StatusUnauthorized, err.Error())
		return
	}

	subject := r.URL.Query().Get("subject")
	xm := r.URL.Query().Get("xm")
	sfz := r.URL.Query().Get("sfz")

	if subject == "" || sfz == "" || xm == "" {
		err := fmt.Errorf("考试类别/身份证号/姓名不能为空")
		Response(w, r, nil, http.StatusUnauthorized, err.Error())
		return
	}

	client := cet.NewPastCetClient()
	client.Captcha = captcha
	client.Subject = subject
	client.PersonID = sfz
	client.Name = xm

	// 查询结果
	err := client.GetPastCetInfo(captchaCookie)
	if err != nil {
		Response(w, r, nil, http.StatusBadRequest, err.Error())
		return
	}

	Response(w, r, client, http.StatusOK, "request ok")
}
