package routers

import (
	"GZHU-Pi/env"
	"GZHU-Pi/services/kafka"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/silenceper/wechat/message"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//是否关闭匿名视图
func fake(u *env.VUser) bool {

	if u.ID <= 2 && u.ID > 0 {
		return true
	}
	if isTestUser(u.StuID.String) {
		return false
	}
	//if len(u.StuID.String) > 2 {
	//	return false
	//}
	return false
}

func TableAccessHandle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	//====== 无需校验token的接口 =======

	if strings.Contains(r.URL.Path, "/auth") ||
		strings.Contains(r.URL.Path, "/wx") ||
		strings.Contains(r.URL.Path, "/jwxt") ||
		strings.Contains(r.URL.Path, "/library") ||
		strings.Contains(r.URL.Path, "/second") {
		next(w, r)
		return
	}
	if !strings.Contains(r.URL.Path, env.Conf.Db.Dbname) &&
		!strings.Contains(strings.ToUpper(r.URL.Path), "QUERIES") {
		next(w, r)
		return
	}
	if strings.ToUpper(r.Method) == "GET" && strings.Contains(r.URL.Path, "_topic") {
		topicViewCounter(r.URL)
	}

	if strings.ToUpper(r.Method) == "GET" {

		if !strings.Contains(strings.ToUpper(r.URL.Path), "QUERIES") && strings.Contains(r.URL.Path, "v_topic") || strings.Contains(r.URL.Path, "v_discuss") {
			//ctx, err := InitCtx(w, r)
			//if err != nil {
			//	return
			//}
			//p := getCtxValue(ctx)
			//if fake(p.user) && strings.Contains(r.URL.Path, "v_topic") {
			//	r.URL.Path = strings.ReplaceAll(r.URL.Path, "v_topic", "v_topic2")
			//}
			//if fake(p.user) && strings.Contains(r.URL.Path, "v_discuss") {
			//	r.URL.Path = strings.ReplaceAll(r.URL.Path, "v_discuss", "v_discuss2")
			//}
		}

		next(w, r)
		return
	}

	//======= 数据库可以找到对应用户、需要检查token =========

	ctx, err := InitCtx(w, r)
	if err != nil {
		return
	}
	switch strings.ToUpper(r.Method) {

	case "GET":
	case "POST":

		switch {
		case strings.Contains(r.URL.Path, "t_topic"):
			err = topicCheck(ctx)
		case strings.Contains(r.URL.Path, "t_discuss"):
			err = discussCheck(ctx)
		case strings.Contains(r.URL.Path, "t_relation"):
			err = relationCheck(ctx)
		case strings.Contains(r.URL.Path, "t_notify"):
			err = courseNotifyCheck(ctx)
			if err != nil {
				Response(w, r, nil, http.StatusBadRequest, err.Error())
			} else {
				Response(w, r, nil, http.StatusOK, "")
			}
			return
		case strings.Contains(r.URL.Path, "t_teach_evaluation"):
			err = teachEvaluationCheck(ctx)
		default:
			err = fmt.Errorf("illegal request")
		}
		if err != nil {
			logs.Error(err)
			Response(w, r, nil, http.StatusBadRequest, err.Error())
			return
		}
	case "PUT", "PATCH":
		p := getCtxValue(ctx)
		if strings.Contains(r.URL.Path, "t_user") {
			logs.Info(r.URL.RawQuery)
			r.URL.RawQuery = fmt.Sprintf("id=%d", p.user.ID)
			if err := userCheck(ctx); err != nil {
				Response(w, r, nil, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			err = fmt.Errorf("illegal request")
			logs.Error(err)
			Response(w, r, nil, http.StatusBadRequest, err.Error())
			return
		}
	case "DELETE":
		p := getCtxValue(ctx)
		qry := strings.ReplaceAll(p.r.URL.Query().Get("id"), "$eq.", "")
		id, err := strconv.ParseInt(qry, 10, 64)
		if err != nil {
			logs.Error(err)
			Response(w, r, nil, http.StatusBadRequest, err.Error())
			return
		}

		switch {
		case strings.Contains(r.URL.Path, "t_topic"):
			var t env.TTopic
			p.gormDB.First(&t, id)
			if t.CreatedBy.Int64 != p.user.ID {
				err = fmt.Errorf("permission denied")
			}
		case strings.Contains(r.URL.Path, "t_discuss"):
			var t env.TDiscuss
			p.gormDB.First(&t, id)
			if t.CreatedBy.Int64 != p.user.ID {
				err = fmt.Errorf("permission denied")
			}
		case strings.Contains(r.URL.Path, "t_relation"):
			var t env.TRelation
			p.gormDB.First(&t, id)
			if t.CreatedBy.Int64 != p.user.ID {
				err = fmt.Errorf("permission denied")
			}
		case strings.Contains(r.URL.Path, "t_notify"):
			var t env.TNotify
			p.gormDB.First(&t, id)
			if t.CreatedBy.Int64 != p.user.ID {
				err = fmt.Errorf("permission denied")
			} else {
				_ = env.RedisCli.ZRem(env.KeyCourseNotifyZSet, t.Digest.String).Err()
			}

		default:
			err = fmt.Errorf("illegal request table")
		}
		if err != nil {
			logs.Error(err)
			Response(w, r, nil, http.StatusBadRequest, err.Error())
			return
		}
	default:
		_, _ = w.Write([]byte("unsupported method: " + r.Method))
		return
	}
	next(w, r)
}

func topicCheck(ctx context.Context) (err error) {
	p := getCtxValue(ctx)

	body, err := ioutil.ReadAll(p.r.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	defer p.r.Body.Close()
	if len(body) == 0 {
		err = fmt.Errorf("Call api by post with empty body ")
		logs.Error(err)
		return
	}
	var t env.TTopic
	err = json.Unmarshal(body, &t)
	if err != nil {
		logs.Error(err)
		return
	}
	if t.Type.String == "" || t.Title.String == "" || t.Content.String == "" {
		err = fmt.Errorf("必要字段咋能为空")
		logs.Error(err)
		return
	}
	if t.Anonymous.Bool == true && t.Anonymity.String == "" {
		err = fmt.Errorf("请指定 Anonymity 的值")
		logs.Error(err)
		return
	}
	if t.CreatedBy.Valid {
		err = fmt.Errorf("不能手动指定created_by")
		logs.Error(err)
		return
	}

	newBodyStr := fmt.Sprintf(`%s,"created_by":%d}`, strings.TrimSuffix(string(body), "}"), p.user.ID)

	body = []byte(newBodyStr)
	p.r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return
}

func discussCheck(ctx context.Context) (err error) {
	p := getCtxValue(ctx)

	body, err := ioutil.ReadAll(p.r.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	defer p.r.Body.Close()
	if len(body) == 0 {
		err = fmt.Errorf("Call api by post with empty body ")
		logs.Error(err)
		return
	}
	var t env.TDiscuss
	err = json.Unmarshal(body, &t)
	if err != nil {
		logs.Error(err)
		return
	}
	//if t.Object == "" {
	//	err = fmt.Errorf("必要字段咋能为空")
	//	logs.Error(err)
	//	return
	//}
	if t.ObjectID <= 0 || t.Content.String == "" {
		err = fmt.Errorf("必要字段咋能为空")
		logs.Error(err)
		return
	}
	if t.Anonymous.Bool == true && t.Anonymity.String == "" {
		err = fmt.Errorf("请指定 Anonymity 的值")
		logs.Error(err)
		return
	}
	if t.CreatedBy.Valid {
		err = fmt.Errorf("不能手动指定created_by")
		logs.Error(err)
		return
	}

	newBodyStr := fmt.Sprintf(`%s,"created_by":%d}`, strings.TrimSuffix(string(body), "}"), p.user.ID)

	body = []byte(newBodyStr)
	p.r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return
}

func relationCheck(ctx context.Context) (err error) {
	p := getCtxValue(ctx)

	body, err := ioutil.ReadAll(p.r.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	defer p.r.Body.Close()
	if len(body) == 0 {
		err = fmt.Errorf("Call api by post with empty body ")
		logs.Error(err)
		return
	}
	var t env.TRelation
	err = json.Unmarshal(body, &t)
	if err != nil {
		logs.Error(err)
		return
	}
	if t.ObjectID <= 0 {
		err = fmt.Errorf("Are you kidding me ? ")
		logs.Error(err)
		return
	}
	if t.Object != "t_topic" && t.Object != "t_discuss" {
		err = fmt.Errorf("unsupported object name: %s", t.Object)
		logs.Error(err)
		return
	}
	if t.Type != "star" && t.Type != "claim" && t.Type != "favourite" {
		err = fmt.Errorf("Are you kidding me ? ")
		logs.Error(err)
		return
	}
	if t.CreatedBy.Valid {
		err = fmt.Errorf("不能手动指定created_by")
		logs.Error(err)
		return
	}
	//根据唯一主键删除，防止写入冲突
	p.gormDB.Where("object_id=? and object=? and type=? and created_by=?",
		t.ObjectID, t.Object, t.Type, p.user.ID).Delete(env.TRelation{})

	newBodyStr := fmt.Sprintf(`%s,"created_by":%d}`, strings.TrimSuffix(string(body), "}"), p.user.ID)

	body = []byte(newBodyStr)
	p.r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return
}

func userCheck(ctx context.Context) (err error) {

	p := getCtxValue(ctx)

	body, err := ioutil.ReadAll(p.r.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	defer p.r.Body.Close()
	if len(body) == 0 {
		err = fmt.Errorf("Call api by post with empty body ")
		logs.Error(err)
		return
	}
	var u env.TUser
	err = json.Unmarshal(body, &u)
	if err != nil {
		logs.Error(err)
		return
	}
	if u.ID != 0 || u.RoleID.Int64 != 0 || u.OpenID.String != "" {
		err = fmt.Errorf("could not update id/role_id/open_id")
		logs.Error(err)
		return
	}
	if u.Phone.String != "" && !verifyPhone(u.Phone.String) {
		err = fmt.Errorf("%s not a valid phone number", u.Phone.String)
		return
	}
	return
}

//浏览人数+1
func topicViewCounter(u *url.URL) {

	if u == nil {
		return
	}
	q := u.Query().Get("id")
	idStr := strings.Trim(q, "$eq.")

	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		return
	}

	db := env.GetGorm()
	db.Model(&env.TTopic{ID: int64(id)}).
		UpdateColumn("viewed", gorm.Expr("viewed + ?", 1))

}

func courseNotifyCheck(ctx context.Context) (err error) {
	p := getCtxValue(ctx)

	body, err := ioutil.ReadAll(p.r.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	defer p.r.Body.Close()
	p.r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if len(body) == 0 {
		err = fmt.Errorf("Call api by post with empty body ")
		logs.Error(err)
		return
	}
	var t env.TNotify
	err = json.Unmarshal(body, &t)
	if err != nil {
		logs.Error(err)
		return
	}
	//err = validZeroNullValue(&t)
	//if err != nil {
	//	logs.Error(err)
	//	return
	//}

	t.Type = null.StringFrom("上课提醒")
	if t.SentTime.Unix() < time.Now().Unix() {
		err = fmt.Errorf("数据非法 %s %d", t.Type.String, t.SentTime.Unix())
		logs.Error(err)
		return
	}

	if p.user.MpOpenID.String == "" {
		err = fmt.Errorf("用户未绑定公众号，无法创建提醒")
		logs.Error(err)
		return
	}

	var tplData = make(map[string]*message.DataItem)
	err = json.Unmarshal(t.Data, &tplData)
	if err != nil {
		logs.Error(err)
		return
	}
	keyword1 := tplData["keyword1"].Value
	keyword2 := tplData["keyword2"].Value
	keyword3 := tplData["keyword3"].Value
	if len([]rune(keyword1)) > 50 || len([]rune(keyword2)) > 50 || len([]rune(keyword3)) > 50 {
		err = fmt.Errorf("模板字段长度超出")
		logs.Error(err)
		return
	}
	if keyword1 == "" || keyword2 == "" || keyword3 == "" {
		err = fmt.Errorf("参数非法")
		logs.Error(err)
		return
	}
	tplData["first"] = &message.DataItem{Value: "您有一门课程即将开始！"}
	tplData["remark"] = &message.DataItem{Value: "点击进入课程提醒管理"}

	t.Data, err = json.Marshal(tplData)
	if err != nil {
		logs.Error(err)
		return
	}

	//小程序转跳信息
	m := message.Message{}
	m.MiniProgram.AppID = env.Conf.WeiXin.MinAppID
	m.MiniProgram.PagePath = classNotifyMgrPath
	miniProgram, err := json.Marshal(&m.MiniProgram)
	if err != nil {
		logs.Error(err)
		return
	}
	t.ToUser = null.StringFrom(p.user.MpOpenID.String)
	t.TemplateID = null.StringFrom(classNotifyTpl)
	t.Digest = null.StringFrom(env.StringMD5(fmt.Sprintf("%d_%s_%s_%s", p.user.ID, p.user.MpOpenID.String, keyword1, t.SentTime.String())))
	t.MiniProgram = miniProgram
	t.CreatedBy = null.IntFrom(p.user.ID)

	//查重
	_, err = env.RedisCli.ZRank(env.KeyCourseNotifyZSet, t.Digest.String).Result()
	if err != nil && err != redis.Nil {
		logs.Error(err)
		return
	}
	if err == nil {
		err = fmt.Errorf("同一名称时间的提醒任务 %s %s 已经存在", keyword1, t.SentTime.String())
		logs.Error(err)
		return
	}

	if !env.Conf.Kafka.Enabled {
		err = fmt.Errorf("服务不可用")
		logs.Error(err)
		return
	}
	notifies := []env.TNotify{t}
	//写入消息队列
	info, err := json.Marshal(notifies)
	if err != nil {
		logs.Error(err)
		return
	}
	err = env.Kafka.SendData(&kafka.ProduceData{
		Topic: env.QueueTopicCourse,
		Data:  info,
	})
	if err != nil {
		logs.Error(err)
		return
	}

	time.Sleep(200 * time.Millisecond)

	return
}

func teachEvaluationCheck(ctx context.Context) (err error) {
	p := getCtxValue(ctx)

	body, err := ioutil.ReadAll(p.r.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	defer p.r.Body.Close()
	if len(body) == 0 {
		err = fmt.Errorf("Call api by post with empty body ")
		logs.Error(err)
		return
	}
	var t env.TTeachEvaluation
	err = json.Unmarshal(body, &t)
	if err != nil {
		logs.Error(err)
		return
	}
	if t.CourseID == "" || t.TeacherID == "" {
		err = fmt.Errorf("Are you kidding me ? ")
		logs.Error(err)
		return
	}
	if t.CreatedBy.Valid {
		err = fmt.Errorf("不能手动指定created_by")
		logs.Error(err)
		return
	}

	newBodyStr := fmt.Sprintf(`%s,"created_by":%d}`, strings.TrimSuffix(string(body), "}"), p.user.ID)

	body = []byte(newBodyStr)
	p.r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return
}
