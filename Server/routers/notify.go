/**
 * @File: notify
 * @Author: Shaw
 * @Date: 2020/5/28 2:17 AM
 * @Desc

 */

package routers

import (
	"GZHU-Pi/env"
	"GZHU-Pi/services/kafka"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/silenceper/wechat/message"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
)

var beforeMinutes time.Duration = 30                               //通知提前时间
var classNotifyTpl = "aFpe_zN27IOKa3I_WhATW4-CxxcsOhwlFJbLJpz1zuk" //微信公众号上课提醒通知模板
var classNotifyMgrPath = "/pages/Campus/tools/notice"              //通知管理转跳地址
var mpBindPath = "/pages/Setting/login/auth"                       //公众号绑定页面
var syncPath = "/pages/Setting/login/sync"                         //同步中心页面
var wallPost = "/pages/Life/wall/post"                             //广大墙发帖

func init() {
	go func() {
		time.Sleep(5 * time.Second)
		if env.Conf.WeiXin.Notify {
			exp := `0 0/3 * * * *` //每隔3分钟
			logs.Info("添加定时任务: 上课通知提醒 ", exp)
			env.CornTask(exp, SentNotification)
		} else {
			logs.Info("不开启上课通知提醒")
		}
	}()
}

//把还没有过期的课程通知写入kafka消息队列
func AddCourseNotify(courses []*env.TStuCourse, firstMonday string) (err error) {
	if len(courses) == 0 || firstMonday == "" {
		err = fmt.Errorf("skip adding course notification, empty firstMonday or courses")
		logs.Warn(err)
		return
	}
	stuID := courses[0].StuID
	if stuID == "" {
		err = fmt.Errorf("skip adding course notification, empty stu_id")
		logs.Warn(err)
		return
	}
	db := env.GetGorm()
	var user env.TUser
	err = db.Where("stu_id = ?", stuID).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.Error(err)
		return
	}
	if user.MpOpenID.String == "" {
		err = fmt.Errorf("skip adding course notification, empty mp_open_id of stu_id = %s", stuID)
		logs.Warn(err)
		return
	}
	appID := env.Conf.WeiXin.MinAppID
	if appID == "" {
		err = fmt.Errorf("skip adding course notification, empty app_id config")
		logs.Warn(err)
		return
	}

	var notifies []env.TNotify
	for _, c := range courses {
		if len([]rune(c.Teacher)) > 10 {
			c.Teacher = string([]rune(c.Teacher)[0:10]) + "..."
		}
		for _, t := range CalStartTime(firstMonday, c) {
			if time.Now().Unix() > t.Unix() { //已经超过上课时间,跳过
				continue
			}

			//模板数据
			data := map[string]*message.DataItem{
				"first":    {Value: "您有一门课程即将开始！"},
				"keyword1": {Value: fmt.Sprintf("%s-%s", c.CourseName, c.Teacher)}, //"编译原理-汤茂斌"
				"keyword2": {Value: fmt.Sprintf("%s %s %s",
					c.WhichDay, c.CourseTime, t.Format("01-02 15:04"))}, //"星期三 7-8节 15:45"
				"keyword3": {Value: c.ClassPlace},                       //"理科南312"
				"remark":   {Value: "点击进入课程提醒管理"},
			}

			var addi, tplData, miniProgram []byte
			addi, err = json.Marshal(c)
			if err != nil {
				logs.Error(err)
				return
			}
			tplData, err = json.Marshal(data)
			if err != nil {
				logs.Error(err)
				return
			}
			//小程序转跳信息
			m := message.Message{}
			m.MiniProgram.AppID = appID
			m.MiniProgram.PagePath = classNotifyMgrPath
			miniProgram, err = json.Marshal(&m.MiniProgram)
			if err != nil {
				logs.Error(err)
				return
			}
			//Digest唯一哈希： 通知时间、课程、学生、班级、通知对象 都一样的情况下，只能存在一条
			notify := env.TNotify{
				Digest:   null.StringFrom(env.StringMD5(t.String() + c.StuID + c.CourseID + c.TeacherID + user.MpOpenID.String)),
				Type:     null.StringFrom("上课提醒"),
				SentTime: t.Add(-beforeMinutes * time.Minute), //提前指定时间

				ToUser:      null.StringFrom(user.MpOpenID.String),
				TemplateID:  null.StringFrom(classNotifyTpl),
				Data:        tplData,
				MiniProgram: miniProgram,

				Addi:      addi,
				Status:    null.IntFrom(0),
				CreatedBy: c.CreatedBy,
			}
			logs.Info("添加通知：%s %s", notify.SentTime.String(), notify.ToUser.String)
			notifies = append(notifies, notify)
		}
	}

	if !env.Conf.Kafka.Enabled {
		return
	}
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
	return
}

//根据上课的周段、星期、开始节次 计算这门课的每一节课的开始上课时间
func CalStartTime(firstMonday string, c *env.TStuCourse) (times []time.Time) {
	if c == nil {
		return
	}

	var classTime = map[int]int{
		1:  8*60 + 30, //8:30分上第一节课
		2:  9*60 + 20,
		3:  10*60 + 25,
		4:  11*60 + 15,
		5:  13*60 + 50,
		6:  14*60 + 40,
		7:  15*60 + 45,
		8:  16*60 + 35,
		9:  18*60 + 20,
		10: 19*60 + 10,
		11: 20*60 + 00,
	}

	start, err := time.ParseInLocation("2006-01-02", firstMonday, time.Local)
	if err != nil {
		logs.Error(err)
		return
	}
	for k, v := range c.WeekSection {
		if k&1 == 0 && k+1 < len(c.WeekSection) {

			for week := v; week <= c.WeekSection[k+1]; week++ {
				internalDays := (week-1)*7 + int(c.Weekday) - 1
				internal := internalDays*24*60*60 + classTime[int(c.Start)]*60

				addSecond := start.Unix() + int64(internal)
				abs := time.Unix(addSecond, 0)

				times = append(times, abs)

				//logs.Debug("第%d周 星期%d 第%d节，距离开学：%d天，共%ds",
				//	week, c.Weekday, c.Start, internalDays, internal)
				//logs.Debug(abs.String())
			}
		}
	}
	return
}

func SentNotification() {

	if !wxInit() {
		return
	}

	tpl := message.NewTemplate(wc.Context)
	db := env.GetGorm()

	//在这之前没有发送的通知都发出去
	key := env.KeyCourseNotifyZSet
	val, err := env.RedisCli.ZRangeByScoreWithScores(key, redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprint(time.Now().Unix() + 90),
	}).Result()
	if err != nil {
		logs.Error(err)
		return
	}

	if len(val) == 0 {
		logs.Debug("没有上课提醒通知", time.Now().String())
		return
	}

	for _, v := range val {
		t := time.Unix(int64(v.Score), 0)
		logs.Info("通知任务 digest: %s %s", v.Member, t.String())

		var n env.TNotify
		err := db.Where("digest=?", v.Member).First(&n).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			logs.Error(err, v.Member)
			continue
		}
		if err == gorm.ErrRecordNotFound {
			logs.Debug(v.Member, "任务已经删除")
			err = env.RedisCli.ZRem(key, v.Member).Err()
			if err != nil {
				logs.Error(err, key, v.Member)
				continue
			}
			continue
		}

		data, err := json.Marshal(n)
		if err != nil {
			logs.Error(err, n)
			continue
		}
		if n.ID <= 0 || n.ToUser.String == "" {
			err = fmt.Errorf("illegal record: %s", v.Member)
			env.RedisCli.ZRem(key, v.Member)
			logs.Error(err, v.Member)
			return
		}
		var msg *message.Message
		err = json.Unmarshal(data, &msg)
		if err != nil {
			logs.Error(err)
			continue
		}
		_, err = tpl.Send(msg)
		if err != nil {
			if strings.Contains(err.Error(), "43004") { //已经取消关注
				logs.Warn(msg.ToUser + " 取消关注")
			} else {
				logs.Error(err, v.Member)
				continue
			}
		}
		logs.Info("%s通知成功 to:%s id:%d", n.Type.String, n.ToUser.String, n.ID)

		//更新数据库
		err = db.Model(&env.TNotify{ID: n.ID}).UpdateColumn("status", 2).Error
		if err != nil {
			logs.Error(err, n.ID)
			continue
		}
		err = env.RedisCli.ZRem(key, v.Member).Err()
		if err != nil {
			logs.Error(err, key, v.Member)
			continue
		}
	}
}
