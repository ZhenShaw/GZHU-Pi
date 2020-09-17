/**
 * @File: msg-queue
 * @Author: Shaw
 * @Date: 2020/5/3 6:03 PM
 * @Desc: 创建消息队列

 */

package env

import (
	"GZHU-Pi/services/kafka"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"math"
	"reflect"
	"strconv"
	"time"
)

const (
	//消息队列主题
	QueueTopicGrade   = "kafka-queue-grade"
	QueueTopicStuInfo = "kafka-queue-student-info"
	QueueTopicCourse  = "kafka-queue-course"

	//消息队列offset缓存key
	keyQueueGrade   = "gzhupi:queue:" + QueueTopicGrade
	keyQueueStuInfo = "gzhupi:queue:" + QueueTopicStuInfo
	keyQueueCourse  = "gzhupi:queue:" + QueueTopicCourse

	KeyCourseNotifyZSet = "gzhupi:notify:course_zset" //课程通知有序集合key

)

func ErrHandler(h *kafka.CustomHandler, err error) {
	if err != nil && h != nil {
		logs.Error("%s 消息发生错误：%s", h.Topic, err)
	}
}

//注册成绩保存更新队列
func GradeQueue() (err error) {

	val, err := RedisCli.Get(keyQueueGrade).Result()
	if err != nil && err != redis.Nil {
		logs.Error(err)
		return
	}
	var offset int64
	if val != "" {
		offset, _ = strconv.ParseInt(val, 10, 64)
		offset = offset + 1
	}
	logs.Info("%s offset: %d", QueueTopicGrade, offset)

	h := &kafka.CustomHandler{
		Topic:     QueueTopicGrade,
		CustomFun: CustomGradeMsg,
		ErrorFun:  ErrHandler,
		Offset:    offset,
	}

	err = Kafka.AddCustomer(h)
	if err != nil {
		logs.Error(err)
		return
	}
	return
}

//注册学生信息存储队列
func StuInfoQueue() (err error) {

	val, err := RedisCli.Get(keyQueueStuInfo).Result()
	if err != nil && err != redis.Nil {
		logs.Error(err)
		return
	}
	var offset int64
	if val != "" {
		offset, _ = strconv.ParseInt(val, 10, 64)
		offset = offset + 1
	}
	logs.Info("%s offset: %d", QueueTopicStuInfo, offset)

	h := &kafka.CustomHandler{
		Topic:     QueueTopicStuInfo,
		CustomFun: CustomInfo,
		ErrorFun:  ErrHandler,
		Offset:    offset,
	}

	err = Kafka.AddCustomer(h)
	if err != nil {
		logs.Error(err)
		return
	}
	return
}

//注册课表处理队列
func CourseQueue() (err error) {

	val, err := RedisCli.Get(keyQueueCourse).Result()
	if err != nil && err != redis.Nil {
		logs.Error(err)
		return
	}
	var offset int64
	if val != "" {
		offset, _ = strconv.ParseInt(val, 10, 64)
		offset = offset + 1
	}
	logs.Info("%s offset: %d", QueueTopicCourse, offset)

	h := &kafka.CustomHandler{
		Topic:     QueueTopicCourse,
		CustomFun: CustomNotify,
		ErrorFun:  ErrHandler,
		Offset:    offset,
	}

	err = Kafka.AddCustomer(h)
	if err != nil {
		logs.Error(err)
		return
	}
	return
}

func CustomGradeMsg(msg *sarama.ConsumerMessage) (err error) {
	if msg == nil {
		return
	}
	var grades []*TGrade
	err = json.Unmarshal(msg.Value, &grades)
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info("消费成功 topic: %s offset: %d", msg.Topic, msg.Offset)
	_ = SaveOrUpdateGrade(grades)

	//消费成功 记录更新offset
	err = RedisCli.Set(keyQueueGrade, fmt.Sprint(msg.Offset), 0).Err()
	if err != nil {
		logs.Error(err)
		return
	}
	return
}

func CustomInfo(msg *sarama.ConsumerMessage) (err error) {
	if msg == nil {
		return
	}
	var info *TStuInfo
	err = json.Unmarshal(msg.Value, &info)
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info("消费成功 topic: %s offset: %d", msg.Topic, msg.Offset)
	_ = SaveStuInfo(info)

	//消费成功 记录更新offset
	err = RedisCli.Set(keyQueueStuInfo, fmt.Sprint(msg.Offset), 0).Err()
	if err != nil {
		logs.Error(err)
		return
	}
	return
}

func CustomNotify(msg *sarama.ConsumerMessage) (err error) {
	if msg == nil {
		return
	}
	var notifies []TNotify
	err = json.Unmarshal(msg.Value, &notifies)
	if err != nil {
		logs.Error(err)
		return
	}
	err = SaveCourseNotify(notifies)
	if err != nil {
		return
	}

	logs.Info("消费成功 topic: %s offset: %d", msg.Topic, msg.Offset)
	//消费成功 记录更新offset
	err = RedisCli.Set(keyQueueCourse, fmt.Sprint(msg.Offset), 0).Err()
	if err != nil {
		logs.Error(err)
		return
	}
	return
}

func SaveOrUpdateGrade(grades []*TGrade) (err error) {
	db := GetGorm()
	for _, v := range grades {
		//根据主键查询
		var res = TGrade{}
		err = db.Where("stu_id = ? and course_id = ? and jxb_id = ?",
			v.StuID, v.CourseID, v.JxbID).First(&res).Error
		//不存在记录则插入
		if err == gorm.ErrRecordNotFound {
			logs.Debug("create record for course_id %s", v.CourseID)
			err = db.Create(v).Error
			if err != nil {
				logs.Error(err, v)
			}
			continue
		}
		if err != nil || res.ID <= 0 {
			logs.Error(err, res)
			return
		}
		//存在记录但没有变动，跳过
		if math.Round(res.CourseGpa*10)/10 == v.CourseGpa &&
			res.GradeValue == v.GradeValue &&
			res.Grade == v.Grade &&
			res.Credit == v.Credit &&
			res.Invalid == v.Invalid {
			continue
		}
		v.ID = res.ID
		//更新记录 结构体转换为map
		m := make(map[string]interface{})
		elem := reflect.ValueOf(v).Elem()
		relType := elem.Type()
		for i := 0; i < relType.NumField(); i++ {
			m[relType.Field(i).Name] = elem.Field(i).Interface()
		}
		delete(m, "CreatedAt")
		delete(m, "UpdatedAt")

		err = db.Model(&res).Where("stu_id = ? and course_id = ? and jxb_id = ?",
			v.StuID, v.CourseID, v.JxbID).Updates(m).Error
		if err != nil {
			logs.Error(err, v)
			continue
		}
		logs.Debug("update record: %s %s %s ", v.StuID, v.CourseID, v.JxbID)
	}
	return
}

func SaveStuInfo(info *TStuInfo) (err error) {
	db := GetGorm()
	//获取匹配的第一条记录, 否则根据给定的条件创建一个新的记录 (仅支持 struct 和 map 条件)
	//db.FirstOrCreate(info, &info)
	var stu TStuInfo
	err = db.Where("stu_id = ?", info.StuID).First(&stu).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = db.FirstOrCreate(info, &info).Error
			if err != nil {
				logs.Error(err)
				return
			}
			return
		}
		logs.Error(err)
		return
	}
	if stu.ClassID != info.ClassID || stu.College != info.College ||
		stu.CollegeID != info.CollegeID || stu.MajorID != info.MajorID {
		logs.Info("%s 更新信息", stu.StuID)
		err = db.Model(&stu).Where("stu_id = ?", info.StuID).Update(info).Error
		if err != nil {
			logs.Error(err)
			return
		}
	}
	return
}

func SaveCourseNotify(notifies []TNotify) (err error) {

	db := GetGorm()

	for _, v := range notifies {
		if !v.Digest.Valid || v.Digest.String == "" {
			err = fmt.Errorf("enpty digest %v", v)
			logs.Error(err)
			continue
		}

		//分布式锁
		key := fmt.Sprintf("gzhupi:lock:%s", v.Digest.String)
		ok, err := RedisCli.SetNX(key, v.Digest.String, 2*time.Second).Result()
		if !ok {
			logs.Warn("get distributed lock failed: %s", key)
			continue
		}

		var val int64
		val, err = RedisCli.ZRank(KeyCourseNotifyZSet, v.Digest.String).Result()
		if err != nil && err != redis.Nil {
			logs.Error(err)
			continue
		}
		if err == nil {
			logs.Warn("重复消费跳过 zset member: %s rank: %v", v.Digest.String, val)
			continue
		}

		err = db.Create(&v).Error
		if err != nil {
			logs.Error(err, v)
			continue
		}

		//设置有序集合 发送通知时间为排序依据
		err = RedisCli.ZAdd(KeyCourseNotifyZSet, redis.Z{
			Member: v.Digest.String,
			Score:  float64(v.SentTime.Unix()),
		}).Err()
		if err != nil {
			logs.Error(err, v.Digest.String)
			continue
		}
	}
	return
}
