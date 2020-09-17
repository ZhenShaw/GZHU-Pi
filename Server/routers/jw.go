package routers

import (
	"GZHU-Pi/env"
	"GZHU-Pi/pkg"
	"GZHU-Pi/pkg/gzhu_jw"
	"context"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/mo7zayed/reqip"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TODO Note
//var Jwxt = make(map[string]pkg.Jwxt)
var Jwxt sync.Map

func isTestUser(user string) bool {
	if user == "20180831" || user == "20200504" {
		return true
	}
	return false
}

func newJWClient(r *http.Request, username, password string) (client pkg.Jwxt, err error) {

	school := r.URL.Query().Get("school")

	//测试用户
	u, err := ReadRequestArg(r, "username")
	user, _ := u.(string)
	if isTestUser(user) {
		school = "demo"
	}

	if school == "" {
		school = "gzhu"
	}

	if school == "gzhu" {
		client, err = gzhu_jw.BasicAuthClient(username, password)

		if err == nil {
			if uid, err := GetUserID(r); uid > 0 && err == nil {
				s := fmt.Sprintf(`update t_user set stu_id='%s' where id=%d and length(stu_id)<8`,
					username, uid)
				env.GetGorm().Exec(s)
			}
		} else {
			if strings.Contains(err.Error(), "EOF") {
				err = fmt.Errorf("服务忙，请几分钟后再试")
			}
		}

		return
	}

	if school == "demo" {
		client = &pkg.Demo{Username: username, Password: password}
	}
	return
}

//根据时间获取学期字符串，2 <= month < 8 作为第二学期
func getYearSem(t time.Time) (sem string) {
	if t.Month() >= 2 && t.Month() < 8 {
		sem = fmt.Sprintf("%d-%d-2", t.Year()-1, t.Year()) //第二学期
	} else if t.Month() <= 1 {
		sem = fmt.Sprintf("%d-%d-1", t.Year()-1, t.Year()) //第一学期
	} else {
		sem = fmt.Sprintf("%d-%d-1", t.Year(), t.Year()+1) //第一学期
	}
	return
}

func getCacheKey(r *http.Request, username string) string {
	s := r.URL.Query().Get("school")
	if r.URL.Query().Get("school") == "" {
		s = "gzhu"
	}
	return s + username
}

//教务系统统一中间件，做一些准备客户端的公共操作
func JWMiddleWare(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if strings.ToUpper(r.Method) == "GET" {
			username := r.URL.Query().Get("username")
			if username == "" {
				Response(w, r, nil, http.StatusUnauthorized, "Unauthorized")
				return
			}
			//client, ok := Jwxt[getCacheKey(r, username)]
			c, _ := Jwxt.Load(getCacheKey(r, username))
			client, ok := c.(pkg.Jwxt)
			if !ok {
				Response(w, r, nil, http.StatusUnauthorized, "Unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), "client", client)
			// 创建新的请求
			r = r.WithContext(ctx)
			next(w, r)
			return
		}

		u, err := ReadRequestArg(r, "username")
		p, err0 := ReadRequestArg(r, "password")
		if err != nil || err0 != nil {
			logs.Error(err, err0)
			Response(w, r, nil, http.StatusBadRequest, err.Error())
			return
		}
		username, _ := u.(string)
		password, _ := p.(string)
		if username == "" || password == "" {
			Response(w, r, nil, http.StatusUnauthorized, "Unauthorized")
			return
		}
		logs.Info("用户：%s IP: %s 接口：%s ", username, reqip.GetClientIP(r), r.URL.Path)

		//从缓存中获取客户端，不存在或者过期则创建
		//client, ok := Jwxt[getCacheKey(r, username)]
		c, _ := Jwxt.Load(getCacheKey(r, username))
		client, ok := c.(pkg.Jwxt)
		if !ok || client == nil || time.Now().After(client.GetExpiresAt()) {
			if client != nil && time.Now().After(client.GetExpiresAt()) {
				logs.Debug("客户端在 %ds 前过期了", time.Now().Unix()-client.GetExpiresAt().Unix())
			}
			client, err = newJWClient(r, username, password)
			if err != nil {
				logs.Error(err)
				Response(w, r, nil, http.StatusUnauthorized, err.Error())
				return
			}
			//将客户端存入缓存
			//Jwxt[getCacheKey(r, username)] = client
			Jwxt.Store(getCacheKey(r, username), client)
		}
		if client != nil && !time.Now().After(client.GetExpiresAt()) {
			logs.Debug("客户端正常 %ds 后过期", client.GetExpiresAt().Unix()-time.Now().Unix())
		}
		//如果客户端不发生错误而被删除，则更新过期时间
		defer func() {
			c, _ := Jwxt.Load(getCacheKey(r, username))
			client, ok := c.(pkg.Jwxt)
			if ok || client != nil {
				client.SetExpiresAt(time.Now().Add(20 * time.Minute))
			}
		}()
		//把客户端通过context传递给下一级
		ctx := context.WithValue(r.Context(), "client", client)
		// 创建新的请求
		r = r.WithContext(ctx)
		next(w, r)

	}
}

//获取课表信息，参数为 year semester
func Course(w http.ResponseWriter, r *http.Request) {
	//从context提取客户端
	c := r.Context().Value("client")
	if c == nil {
		Response(w, r, nil, http.StatusInternalServerError, "get nil client from context")
		return
	}
	client, ok := c.(pkg.Jwxt)
	if !ok {
		Response(w, r, nil, http.StatusInternalServerError, "get a wrong client from context")
		return
	}

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

	data, err := client.GetCourse(year, sem)
	if err != nil {
		logs.Error(err)
		Jwxt.Delete(getCacheKey(r, client.GetUsername())) //发生错误，从缓存中删除
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}

	if data != nil {
		for _, v := range data.CourseList {
			v.YearSem = ys
			v.StuID = client.GetUsername()
		}
	}

	go pkg.SetDemoCache("course", client.GetUsername(), data)
	Response(w, r, data, http.StatusOK, "request ok")

	//====响应后的处理
	if isTestUser(client.GetUsername()) {
		return
	}
	if data == nil {
		return
	}
	var userID int64
	if len(r.Cookies()) > 0 {
		userID, err = GetUserID(r)
		if err != nil {
			logs.Error(err, r.Cookies())
			return
		}
	} else {
		logs.Warn("miss cookie 跳过加入提醒")
		return
	}
	for _, v := range data.CourseList {
		v.CreatedBy = null.IntFrom(userID)
	}

	if len(data.CourseList) == 0 {
		return
	}

	//非本学期课程不加入提醒
	if ys != getYearSem(time.Now()) {
		logs.Warn("%s != %s 跳过加入提醒", ys, getYearSem(time.Now()))
		return
	}

	//添加自动提醒，每个学生每学期，只自动设置上课提醒一次
	go func() {
		if !env.Conf.Kafka.Enabled {
			return
		}
		fm, _ := ReadRequestArg(r, "first_monday")
		firstMonday, _ := fm.(string)
		if firstMonday == "" {
			firstMonday = gzhu_jw.FirstMonday
			logs.Warn("use default firstMonday %s", firstMonday)
		}
		err = AddCourseNotify(data.CourseList, firstMonday)
		if err != nil {
			logs.Error(err)
			return
		}

		key := fmt.Sprintf("gzhupi:notify:course:stu:%d_%s_%s", userID, ys, client.GetUsername())
		env.RedisCli.Set(key, key, 120*24*time.Hour)
		//_, err := env.RedisCli.Get(key).Result()
		//if err == redis.Nil {
		//	err = AddCourseNotify(data.CourseList, firstMonday)
		//	if err != nil {
		//		return
		//	}
		//	err = env.RedisCli.Set(key, key, 120*24*time.Hour).Err()
		//	if err != nil {
		//		logs.Error(err)
		//		return
		//	}
		//} else if err != nil {
		//	logs.Error(err, key)
		//	return
		//}
		//logs.Debug("key exists, skip add course notify key=%s", key)
	}()

}

func Exam(w http.ResponseWriter, r *http.Request) {
	//从context提取客户端
	c := r.Context().Value("client")
	if c == nil {
		Response(w, r, nil, http.StatusInternalServerError, "get nil client from context")
		return
	}
	client, ok := c.(pkg.Jwxt)
	if !ok {
		Response(w, r, nil, http.StatusInternalServerError, "get a wrong client from context")
		return
	}

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

	data, err := client.GetExam(year, sem)
	if err != nil {
		logs.Error(err)
		Jwxt.Delete(getCacheKey(r, client.GetUsername()))
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}
	go pkg.SetDemoCache("exam", client.GetUsername(), data)
	Response(w, r, data, http.StatusOK, "request ok")
}

func Grade(w http.ResponseWriter, r *http.Request) {

	//从context提取客户端
	c := r.Context().Value("client")
	if c == nil {
		Response(w, r, nil, http.StatusInternalServerError, "get nil client from context")
		return
	}
	client, ok := c.(pkg.Jwxt)
	if !ok {
		Response(w, r, nil, http.StatusInternalServerError, "get a wrong client from context")
		return
	}

	//====缓存处理
	var gs = &env.CacheOptions{
		Key:      fmt.Sprintf("gzhupi:grade:%s", client.GetUsername()),
		Duration: 15 * time.Minute,
		Receiver: new(gzhu_jw.GradeData),
		Fun: func() (interface{}, error) {
			return client.GetAllGrade("", "")
		},
	}
	_, err := env.GetSetCache(gs)
	if err != nil {
		logs.Error(err)
		Jwxt.Delete(getCacheKey(r, client.GetUsername()))
		if err == gzhu_jw.AuthError {
			Response(w, r, nil, http.StatusUnauthorized, err.Error())
			return
		}
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}
	var data = gs.Receiver.(*gzhu_jw.GradeData)

	go pkg.SetDemoCache("grade", client.GetUsername(), data)
	Response(w, r, data, http.StatusOK, "request ok")

	//if data == nil || data.StuInfo == nil || usingCache {
	//	return
	//}
	//
	//stuInfo := data.StuInfo
	//var grades []*env.TGrade
	//for _, v := range data.SemList {
	//	grades = append(grades, v.GradeList...)
	//}
	////加入消息队列
	//go func() {
	//	if !env.Conf.Kafka.Enabled {
	//		return
	//	}
	//	info, err := json.Marshal(stuInfo)
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//	err = env.Kafka.SendData(&kafka.ProduceData{
	//		Topic: env.QueueTopicStuInfo,
	//		Data:  info,
	//	})
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//
	//	//====
	//	g, err := json.Marshal(grades)
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//	err = env.Kafka.SendData(&kafka.ProduceData{
	//		Topic: env.QueueTopicGrade,
	//		Data:  g,
	//	})
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//}()

}

func EmptyRoom(w http.ResponseWriter, r *http.Request) {
	//从context提取客户端
	c := r.Context().Value("client")
	if c == nil {
		Response(w, r, nil, http.StatusInternalServerError, "get nil client from context")
		return
	}
	client, ok := c.(pkg.Jwxt)
	if !ok {
		Response(w, r, nil, http.StatusInternalServerError, "get a wrong client from context")
		return
	}

	data, err := client.GetEmptyRoom(r)
	if err != nil {
		logs.Error(err)
		Jwxt.Delete(getCacheKey(r, client.GetUsername()))
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}
	go pkg.SetDemoCache("empty-room", client.GetUsername(), data)
	Response(w, r, data, http.StatusOK, "request ok")
}

func Achieve(w http.ResponseWriter, r *http.Request) {
	//从context提取客户端
	c := r.Context().Value("client")
	if c == nil {
		Response(w, r, nil, http.StatusInternalServerError, "get nil client from context")
		return
	}
	client, ok := c.(pkg.Jwxt)
	if !ok {
		Response(w, r, nil, http.StatusInternalServerError, "get a wrong client from context")
		return
	}

	data, err := client.GetAchieve()
	if err != nil {
		logs.Error(err)
		Jwxt.Delete(getCacheKey(r, client.GetUsername()))
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}
	go pkg.SetDemoCache("achieve", client.GetUsername(), data)
	Response(w, r, data, http.StatusOK, "request ok")
}

func AllCourse(w http.ResponseWriter, r *http.Request) {
	//从context提取客户端
	c := r.Context().Value("client")
	if c == nil {
		Response(w, r, nil, http.StatusInternalServerError, "get nil client from context")
		return
	}
	client, ok := c.(pkg.Jwxt)
	if !ok {
		Response(w, r, nil, http.StatusInternalServerError, "get a wrong client from context")
		return
	}

	year := r.URL.Query().Get("year")
	sem := r.URL.Query().Get("sem")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	count, _ := strconv.Atoi(r.URL.Query().Get("count"))

	data, csvData, err := client.SearchAllCourse(year, sem, page, count)
	if err != nil {
		logs.Error(err)
		Jwxt.Delete(getCacheKey(r, client.GetUsername()))
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}

	//导出文件
	if r.URL.Query().Get("action") == "export" {
		w.Header().Set("Content-Type", "application/csv")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=%s", "export.csv"))
		_, _ = w.Write(csvData)
	} else {
		Response(w, r, data, http.StatusOK, "request ok")
	}
}

func Rank(w http.ResponseWriter, r *http.Request) {
	tip, _ := env.RedisCli.Get("gzhupi:param:tips").Result()
	if tip == "" {
		tip = "暂无法提供服务"
	}
	Response(w, r, nil, http.StatusInternalServerError, tip)
	return
	username := r.URL.Query().Get("username")
	user, err := VUserByCookies(r)
	if err != nil {
		Response(w, r, nil, http.StatusBadRequest, err.Error())
		return
	}
	if user.StuID.String != "" && user.StuID.String != username && !isTestUser(username) {
		err = fmt.Errorf("Unauthorized ")
		logs.Error(err)
		Response(w, r, nil, http.StatusUnauthorized, err.Error())
		return
	}
	logs.Info("用户：%s IP: %s 接口：%s ", username, reqip.GetClientIP(r), r.URL.Path)

	var client pkg.Jwxt
	client = &gzhu_jw.JWClient{Username: username}
	if isTestUser(username) {
		client = &pkg.Demo{Username: username}
	}

	//====缓存处理
	var gs = &env.CacheOptions{
		Key:      fmt.Sprintf("gzhupi:rank:%s", username),
		Duration: 15 * time.Minute,
		Receiver: make(map[string]interface{}),
		Fun: func() (interface{}, error) {
			return client.GetRank(client.GetUsername())
		},
	}
	_, err = env.GetSetCache(gs)
	if err != nil {
		logs.Error(err)
		Response(w, r, nil, http.StatusInternalServerError, err.Error())
		return
	}

	var data = gs.Receiver.(map[string]interface{})

	go pkg.SetDemoCache("rank", username, data)
	Response(w, r, data, http.StatusOK, "request ok")
}
