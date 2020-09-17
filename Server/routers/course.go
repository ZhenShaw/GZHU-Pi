/**
 * @File: course
 * @Author: Shaw
 * @Date: 2020/8/27 1:04 AM
 * @Desc

 */

package routers

import (
	"GZHU-Pi/env"
	"GZHU-Pi/pkg"
	"GZHU-Pi/pkg/gzhu_jw"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

type Req struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Year     string `json:"year"`
	Sem      string `json:"sem"`
}

func Course2(w http.ResponseWriter, r *http.Request) {

	u, err := ReadRequestArg(r, "username")
	p, err0 := ReadRequestArg(r, "password")
	if err != nil || err0 != nil {
		logs.Error(err, err0)
		Response(w, r, nil, http.StatusBadRequest, err.Error())
		return
	}
	username, _ := u.(string)
	password, _ := p.(string)

	year, sem := gzhu_jw.Year, gzhu_jw.SemCode[1]
	s, _ := ReadRequestArg(r, "year_sem")
	ys, _ := s.(string)
	yearSem := strings.Split(ys, "-")
	if len(yearSem) == 3 {
		year = yearSem[0]
		sem = yearSem[2]
		if sem == "1" {
			sem = "3"
		}
		if sem == "2" {
			sem = "12"
		}
	}

	req := Req{
		Username: username,
		Password: password,
		Year:     year,
		Sem:      sem,
	}

	var gs = &env.CacheOptions{
		Key:      fmt.Sprintf("gzhupi:course:%s:%s", username, s),
		Duration: 30 * time.Minute,
		Receiver: new(gzhu_jw.CourseData),
		Fun: func() (interface{}, error) {
			return GetCourse2(req)
		},
	}
	_, err = env.GetSetCache(gs)
	if err != nil {
		logs.Error(err)
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}

	var data = gs.Receiver.(*gzhu_jw.CourseData)

	go pkg.SetDemoCache("course", username, data)
	Response(w, r, data, http.StatusOK, "request ok")

}

func GetCourse2(req Req) (courseData *gzhu_jw.CourseData, err error) {

	var seleniumUrl = os.Getenv("SELENIUM_URL")
	if seleniumUrl == "" {
		seleniumUrl = "http://cst.gzhu.edu.cn:9010/jwxt/course"
	}

	postData, _ := json.Marshal(req)

	logs.Info("start post to %s", seleniumUrl)

	var resp *http.Response
	header := http.Header{"Content-Type": []string{"application/json"}}
	resp, err = doRequest("POST", seleniumUrl, header, bytes.NewReader(postData))
	if err != nil {
		logs.Error(err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logs.Info(string(body))
		e := make(map[string]interface{})
		_ = json.Unmarshal(body, &e)
		err = fmt.Errorf(fmt.Sprintf("%s:%s", e["msg"], e["detail"]))
		return
	}

	creditMatcher := make(map[string]float64)
	courseData = &gzhu_jw.CourseData{
		CourseList:    gzhu_jw.ParseCourse(body, creditMatcher),
		SjkCourseList: gzhu_jw.ParseSjk(body),
	}
	return
}

var transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

func doRequest(method, url string, header http.Header, body io.Reader) (*http.Response, error) {
	t1 := time.Now()

	cookieJar, _ := cookiejar.New(nil)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	client := &http.Client{
		Jar:       cookieJar,
		Timeout:   time.Minute,
		Transport: transport,
	}
	resp, err := client.Do(req)

	logs.Debug("请求耗时：", time.Since(t1), url)
	return resp, err
}
