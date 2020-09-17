/**
 * @File: demo
 * @Author: Shaw
 * @Date: 2020/3/24 11:18 PM
 * @Last Modified by: Shaw
 * @Last Modified by: 2020/3/24 11:18 PM
 * @Desc

 */

package pkg

import (
	"GZHU-Pi/env"
	"GZHU-Pi/pkg/gzhu_jw"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"math/rand"
	"net/http"
	"time"
)

type Demo struct {
	Username  string
	Password  string
	ExpiresAt time.Time
	Client    *http.Client
}

func (d Demo) GetRank(stuID string) (result map[string]interface{}, err error) {
	res := GetDemoCache("rank")

	e := json.Unmarshal([]byte(res), &result)
	if e != nil {
		logs.Error(e)
	}
	return
}

func (d Demo) GetCourse(year, semester string) (courseData *gzhu_jw.CourseData, err error) {
	res := GetDemoCache("course")
	if len(res) < 2 {
		return
	}
	courseData = new(gzhu_jw.CourseData)
	e := json.Unmarshal([]byte(res), courseData)
	if e != nil {
		logs.Error(e)
	}
	for _, v := range courseData.CourseList {
		v.StuID = ""
	}
	return
}

func (d Demo) GetExam(year, sem string) (exams []*gzhu_jw.Exam, err error) {
	res := GetDemoCache("exam")
	if len(res) < 2 {
		return
	}
	e := json.Unmarshal([]byte(res), &exams)
	if e != nil {
		logs.Error(e)
	}
	return
}

func (d Demo) GetAllGrade(year, sem string) (gradeData *gzhu_jw.GradeData, err error) {
	res := GetDemoCache("grade")
	if len(res) < 2 {
		return
	}
	gradeData = new(gzhu_jw.GradeData)
	e := json.Unmarshal([]byte(res), &gradeData)
	if e != nil {
		logs.Error(e)
	}
	gradeData.StuInfo = nil

	for _, v1 := range gradeData.SemList {
		for _, v2 := range v1.GradeList {
			v2.StuID = ""
		}
	}

	return
}

func (d Demo) GetEmptyRoom(r *http.Request) (data *gzhu_jw.RoomData, err error) {
	res := GetDemoCache("empty-room")
	if len(res) < 2 {
		return
	}
	data = new(gzhu_jw.RoomData)
	e := json.Unmarshal([]byte(res), data)
	if e != nil {
		logs.Error(e)
	}
	return
}

func (d Demo) GetAchieve() (achieves []*gzhu_jw.Achieve, err error) {
	res := GetDemoCache("achieve")
	if len(res) < 2 {
		return
	}
	e := json.Unmarshal([]byte(res), &achieves)
	if e != nil {
		logs.Error(e)
	}
	return
}

func (d Demo) SearchAllCourse(xnm, xqm string, page, count int) (data []gzhu_jw.RawCourse, csvData []byte, err error) {
	res := GetDemoCache("all-course")
	if len(res) < 2 {
		return
	}
	e := json.Unmarshal([]byte(res), &data)
	if e != nil {
		logs.Error(e)
	}
	return
}

func (d Demo) GetExpiresAt() time.Time {
	return d.ExpiresAt
}

func (d Demo) SetExpiresAt(t time.Time) {
	d.ExpiresAt = t
}

func (d Demo) GetUsername() string {
	return d.Username
}

//设置体验用户缓存数据，体验用户随机从缓存提取结果返回
func SetDemoCache(keyType, subKey string, data interface{}) {
	if keyType == "" || fmt.Sprint(data) == "<nil>" {
		return
	}
	cache, err := json.Marshal(data)
	if err != nil {
		logs.Error(err)
		return
	}
	if len(subKey) >= 10 {
		subKey = subKey[:6]
	}

	key := fmt.Sprintf("gzhupi:demo:%s:%s", keyType, subKey)
	_, err = env.RedisCli.Set(key, cache, 120*24*time.Hour).Result()
	if err != nil {
		logs.Error(err)
	}
}

func GetDemoCache(keyType string) (data string) {

	key := fmt.Sprintf("gzhupi:demo:%s:*", keyType)
	val, _, err := env.RedisCli.Scan(0, key, 20000).Result()
	if err != nil {
		logs.Error(err)
		return
	}
	if len(val) == 0 {
		return
	}

	key = val[rand.Intn(len(val))]
	logs.Info(key)

	data, err = env.RedisCli.Get(key).Result()
	if err != nil {
		logs.Error(err)
		return
	}
	return
}
