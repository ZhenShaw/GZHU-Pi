package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/jinzhu/gorm"
	"math"
	"reflect"
	"time"
)

type TGrade struct {
	ID       int64  `json:"id,omitempty" remark:"id" gorm:"primary_key"`
	StuID    string `json:"stu_id" remark:"学号" gorm:"type:varchar;not null"`
	CourseID string `json:"course_id" remark:"课程ID" gorm:"type:varchar;not null"`
	JxbID    string `json:"jxb_id" remark:"教学班ID" gorm:"type:varchar;not null"`

	Credit     float64 `json:"credit" remark:"学分" gorm:"type:numeric(5,2)"`
	CourseGpa  float64 `json:"course_gpa" remark:"课程绩点" gorm:"type:numeric(5,2)"`
	GradeValue float64 `json:"grade_value" remark:"成绩分数" gorm:"type:numeric(5,2)"`
	Grade      string  `json:"grade" remark:"成绩" gorm:"type:varchar"`
	CourseName string  `json:"course_name" remark:"课程名称" gorm:"type:varchar"`
	CourseType string  `json:"course_type" remark:"课程类型" gorm:"type:varchar"`
	ExamType   string  `json:"exam_type" remark:"考试类型" gorm:"type:varchar"`
	Invalid    string  `json:"invalid" remark:"是否作废" gorm:"type:varchar"`
	Semester   string  `json:"semester" remark:"学期" gorm:"type:varchar"`
	Teacher    string  `json:"teacher" remark:"教师" gorm:"type:varchar"`
	Year       string  `json:"year" remark:"学年如2018-2019" gorm:"type:varchar"`
	YearSem    string  `json:"year_sem" remark:"学年学期" gorm:"type:varchar"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"default:current_timestamp"`
}

func SaveOrUpdateGrade(grades []*TGrade) {

	for _, v := range grades {
		//根据主键查询
		var res = TGrade{}
		result := db.Where("stu_id = ? and course_id = ? and jxb_id = ?",
			v.StuID, v.CourseID, v.JxbID).First(&res)

		//不存在记录则插入
		if result.Error == gorm.ErrRecordNotFound {
			logs.Debug("create record for course_id %s", v.CourseID)
			db.Create(v)
			continue
		}
		//存在记录但没有变动，跳过
		if math.Round(res.CourseGpa*10)/10 == v.CourseGpa &&
			res.GradeValue == v.GradeValue &&
			res.Grade == v.Grade &&
			res.Credit == v.Credit &&
			res.Invalid == v.Invalid {
			continue
		}
		//更新记录 结构体转换为map
		m := make(map[string]interface{})
		elem := reflect.ValueOf(v).Elem()
		relType := elem.Type()
		for i := 0; i < relType.NumField(); i++ {
			m[relType.Field(i).Name] = elem.Field(i).Interface()
		}
		delete(m, "CreatedAt")
		delete(m, "UpdatedAt")

		result = db.Model(&res).Where("stu_id = ? and course_id = ? and jxb_id = ?",
			v.StuID, v.CourseID, v.JxbID).Updates(m)
		if result.Error != nil {
			logs.Error(result.Error)
			continue
		}
		logs.Debug("update record: %s %s %s ", v.StuID, v.CourseID, v.JxbID)
	}
}
