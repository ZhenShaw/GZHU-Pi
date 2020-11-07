/**
 * @File: models
 * @Author: Shaw
 * @Date: 2020/5/3 5:22 PM
 * @Desc

 */

package env

import (
	"github.com/jmoiron/sqlx/types"
	"gopkg.in/guregu/null.v3"
	"time"
)

type TApiRecord struct {
	ID        int64     `json:"id,omitempty" gorm:"primary_key"`
	Username  string    `json:"username,omitempty" gorm:"type:varchar"`
	Uri       string    `json:"uri,omitempty" gorm:"type:varchar"`
	Duration  int64     `json:"duration,omitempty" gorm:"type:real"` //耗时统计：毫秒
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"default:current_timestamp"`
}

type TDiscuss struct {
	ID       int64       `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	Object   string      `json:"object,omitempty" remark:"实体表名" gorm:"type:varchar;not null"`
	ObjectID int64       `json:"object_id,omitempty" remark:"主题对象记录ID" gorm:"type:bigint;not null"`
	Content  null.String `json:"content,omitempty" remark:"主体内容" gorm:"type:varchar"`
	ReplyID  null.Int    `json:"reply_id,omitempty" remark:"回复留言id" gorm:"type:bigint"` // reference to t_discuss(id)

	Mark      null.Float     `json:"mark,omitempty" remark:"打星、评分" gorm:"type:numeric(5,2)"`
	Image     types.JSONText `json:"image,omitempty" remark:"图片地址[string]" gorm:"type:varchar[]"`
	Anonymous null.Bool      `json:"anonymous,omitempty" remark:"是否匿名" gorm:"type:bool;default:false"`
	Anonymity null.String    `json:"anonymity,omitempty" remark:"匿名/化名" gorm:"type:varchar"`

	//Type      null.String    `json:"type,omitempty" remark:"留言类型(普通/互动)" gorm:"type:varchar"`
	Addi      types.JSONText `json:"addi,omitempty" remark:"附加信息" gorm:"type:jsonb"`
	Status    null.Int       `json:"status,omitempty" remark:"状态" gorm:"type:smallint;default:0"`
	CreatedBy null.Int       `json:"created_by,omitempty" remark:"创建者" gorm:"type:bigint"`
	CreatedAt time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
}

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

//用户与主题的关系记录 可以用以点赞、参与等
type TRelation struct {
	ID       int64  `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	Object   string `json:"object,omitempty" remark:"实体表名" gorm:"type:varchar;not null"`
	ObjectID int64  `json:"object_id,omitempty" remark:"主题对象记录ID" gorm:"type:bigint;not null"`
	Type     string `json:"type,omitempty" remark:"关系类型:star点赞 claim认领 favourite收藏" gorm:"type:varchar;not null"`

	CreatedBy null.Int  `json:"created_by,omitempty" remark:"创建者" gorm:"type:bigint"`
	CreatedAt time.Time `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
}

type TStuInfo struct {
	ID         int64  `json:"id,omitempty" remark:"id" gorm:"primary_key"`
	StuID      string `json:"stu_id,omitempty" remark:"学号" gorm:"type:varchar;unique_index;not null"`
	StuName    string `json:"stu_name,omitempty" remark:"姓名" gorm:"type:varchar"`
	AdmitYear  string `json:"admit_year,omitempty" remark:"年级" gorm:"type:varchar"`
	ClassID    string `json:"class_id,omitempty" remark:"班级id" gorm:"type:varchar"`
	College    string `json:"college,omitempty" remark:"学院" gorm:"type:varchar"`
	CollegeID  string `json:"college_id,omitempty" remark:"学院id" gorm:"type:varchar"`
	Major      string `json:"major,omitempty" remark:"专业" gorm:"type:varchar"`
	MajorClass string `json:"major_class,omitempty" remark:"专业班级" gorm:"type:varchar"`
	MajorID    string `json:"major_id,omitempty" remark:"专业id" gorm:"type:varchar"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"default:current_timestamp"`
}

type TTopic struct {
	ID      int64       `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	Type    null.String `json:"type,omitempty" remark:"主题类型" gorm:"type:varchar"`
	Title   null.String `json:"title,omitempty" remark:"标题" gorm:"type:varchar"`
	Content null.String `json:"content,omitempty" remark:"主体内容" gorm:"type:varchar"`
	//Mention   null.String `json:"mention,omitempty" remark:"提及@用户" gorm:"type:varchar"`
	Category  null.String    `json:"category,omitempty" remark:"归属类别" gorm:"type:varchar"`
	Image     types.JSONText `json:"image,omitempty" remark:"图片地址" gorm:"type:varchar[]"`
	Label     types.JSONText `json:"label,omitempty" remark:"标签" gorm:"type:varchar[]"`
	Viewed    null.Int       `json:"viewed,omitempty" remark:"浏览量" gorm:"type:int;default:0"`
	Anonymous null.Bool      `json:"anonymous,omitempty" remark:"是否匿名" gorm:"type:bool;default:false"`
	Anonymity null.String    `json:"anonymity,omitempty" remark:"匿名/化名" gorm:"type:varchar"`

	Addi      types.JSONText `json:"addi,omitempty" remark:"附加信息" gorm:"type:jsonb"`
	Status    null.Int       `json:"status,omitempty" remark:"状态" gorm:"type:smallint;default:0"`
	CreatedBy null.Int       `json:"created_by,omitempty" remark:"创建者" gorm:"type:bigint"`
	CreatedAt time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
	UpdatedAt time.Time      `json:"updated_at,omitempty" remark:"更新时间" gorm:"default:current_timestamp"`
}

type TUser struct {
	ID         int64          `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	MinappID   null.Int       `json:"minapp_id,omitempty" remark:"知晓云用户id" gorm:"type:bigint;unique_index"`
	OpenID     null.String    `json:"open_id,omitempty" remark:"微信openid" gorm:"type:varchar;unique_index;not null"`
	MpOpenID   null.String    `json:"mp_open_id,omitempty" remark:"微信公众号openid" gorm:"type:varchar"`
	UnionID    null.String    `json:"union_id,omitempty" remark:"微信unionid" gorm:"type:varchar;unique"`
	StuID      null.String    `json:"stu_id,omitempty" remark:"学号" gorm:"type:varchar"`
	RoleID     null.Int       `json:"role_id,omitempty" remark:"用户角色id" gorm:"type:smallint"`
	Avatar     null.String    `json:"avatar,omitempty" remark:"微信头像" gorm:"type:varchar"`
	ProfilePic null.String    `json:"profile_pic,omitempty" remark:"系统随机头像" gorm:"type:varchar"`
	Nickname   null.String    `json:"nickname,omitempty" remark:"昵称" gorm:"type:varchar"`
	City       null.String    `json:"city,omitempty" remark:"城市" gorm:"type:varchar"`
	Province   null.String    `json:"province,omitempty" remark:"省份" gorm:"type:varchar"`
	Country    null.String    `json:"country,omitempty" remark:"国家" gorm:"type:varchar"`
	Gender     null.Int       `json:"gender,omitempty" remark:"性别" gorm:"type:smallint"`
	Language   null.String    `json:"language,omitempty" remark:"语言" gorm:"type:varchar"`
	Phone      null.String    `json:"phone,omitempty" remark:"手机号码" gorm:"type:varchar"`
	Tag        types.JSONText `json:"tag,omitempty" remark:"身份标签" gorm:"type:varchar[]"`
	CreatedAt  time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty" remark:"更新时间" gorm:"default:current_timestamp"`
}

type VUser struct {
	ID         int64          `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	MinappID   null.Int       `json:"minapp_id,omitempty" remark:"知晓云用户id" gorm:"type:bigint;unique_index"`
	OpenID     null.String    `json:"open_id,omitempty" remark:"微信openid" gorm:"type:varchar;unique_index;not null"`
	MpOpenID   null.String    `json:"mp_open_id,omitempty" remark:"微信公众号openid" gorm:"type:varchar"`
	UnionID    null.String    `json:"union_id,omitempty" remark:"微信unionid" gorm:"type:varchar;unique"`
	StuID      null.String    `json:"stu_id,omitempty" remark:"学号" gorm:"type:varchar"`
	RoleID     null.Int       `json:"role_id,omitempty" remark:"用户角色id" gorm:"type:smallint"`
	Avatar     null.String    `json:"avatar,omitempty" remark:"头像" gorm:"type:varchar"`
	ProfilePic null.String    `json:"profile_pic,omitempty" remark:"系统随机头像" gorm:"type:varchar"`
	Nickname   null.String    `json:"nickname,omitempty" remark:"昵称" gorm:"type:varchar"`
	City       null.String    `json:"city,omitempty" remark:"城市" gorm:"type:varchar"`
	Province   null.String    `json:"province,omitempty" remark:"省份" gorm:"type:varchar"`
	Country    null.String    `json:"country,omitempty" remark:"国家" gorm:"type:varchar"`
	Gender     null.Int       `json:"gender,omitempty" remark:"性别" gorm:"type:smallint"`
	Language   null.String    `json:"language,omitempty" remark:"语言" gorm:"type:varchar"`
	Phone      null.String    `json:"phone,omitempty" remark:"手机号码" gorm:"type:varchar"`
	Tag        types.JSONText `json:"tag,omitempty" remark:"身份标签" gorm:"type:varchar[]"`
	CreatedAt  time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty" remark:"更新时间" gorm:"default:current_timestamp"`

	StuName    string `json:"stu_name,omitempty" remark:"姓名" gorm:"type:varchar"`
	AdmitYear  string `json:"admit_year,omitempty" remark:"年级" gorm:"type:varchar"`
	ClassID    string `json:"class_id,omitempty" remark:"班级id" gorm:"type:varchar"`
	College    string `json:"college,omitempty" remark:"学院" gorm:"type:varchar"`
	CollegeID  string `json:"college_id,omitempty" remark:"学院id" gorm:"type:varchar"`
	Major      string `json:"major,omitempty" remark:"专业" gorm:"type:varchar"`
	MajorClass string `json:"major_class,omitempty" remark:"专业班级" gorm:"type:varchar"`
	MajorID    string `json:"major_id,omitempty" remark:"专业id" gorm:"type:varchar"`
}

type TStuCourse struct {
	ID int64 `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`

	StuID       string  `json:"stu_id" remark:"学号" gorm:"type:varchar"`
	YearSem     string  `json:"year_sem" remark:"学年学期" gorm:"type:varchar"`
	CheckType   string  `json:"check_type" remark:"考核类型" gorm:"type:varchar"`
	ClassPlace  string  `json:"class_place" remark:"上课地点" gorm:"type:varchar"`
	Color       int64   `json:"color" remark:"课表颜色" gorm:"type:smallint"`
	CourseID    string  `json:"course_id" remark:"课程ID" gorm:"type:varchar"`
	CourseName  string  `json:"course_name" remark:"课程名称" gorm:"type:varchar"`
	CourseTime  string  `json:"course_time" remark:"上课时间" gorm:"type:varchar"`
	Credit      float64 `json:"credit" remark:"学分" gorm:"type:varchar"`
	TeacherID   string  `json:"teacher_id" remark:"教工号ID" gorm:"type:varchar"`
	Last        int64   `json:"last" remark:"持续节数" gorm:"type:smallint"`
	Start       int64   `json:"start" remark:"开始节数" gorm:"type:smallint"`
	Teacher     string  `json:"teacher" remark:"教师" gorm:"type:varchar"`
	Weekday     int64   `json:"weekday" remark:"星期几数值" gorm:"type:smallint"`
	Weeks       string  `json:"weeks" remark:"周段" gorm:"type:varchar"`
	WhichDay    string  `json:"which_day" remark:"星期几" gorm:"type:varchar"`
	WeekSection []int   `json:"week_section" remark:"周段[start,end,start,end]" gorm:"type:jsonb"`

	CreatedBy null.Int  `json:"created_by,omitempty" remark:"创建者" gorm:"type:bigint"`
	CreatedAt time.Time `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at,omitempty" remark:"更新时间" gorm:"default:current_timestamp"`
}

type TNotify struct {
	ID     int64       `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	Digest null.String `json:"digest,omitempty" remark:"哈希标识防止重复" gorm:"type:varchar;unique_index;not null"`
	Type   null.String `json:"type,omitempty" remark:"通知类型" gorm:"type:varchar"`

	SentTime time.Time `json:"sent_time,omitempty" remark:"指定发送时间"`

	//微信公众号通知
	ToUser      null.String    `json:"touser,omitempty" gorm:"type:varchar"`      // 必须, 接受者OpenID
	TemplateID  null.String    `json:"template_id,omitempty" gorm:"type:varchar"` // 必须, 模版ID
	URL         null.String    `json:"url,omitempty" gorm:"type:varchar"`         // 可选, 用户点击后跳转的URL, 该URL必须处于开发者在公众平台网站中设置的域中
	Color       null.String    `json:"color,omitempty" gorm:"type:varchar"`       // 可选, 整个消息的颜色, 可以不设置
	Data        types.JSONText `json:"data,omitempty" gorm:"type:jsonb"`          // 必须, 模板数据
	MiniProgram types.JSONText `json:"miniprogram,omitempty" gorm:"type:jsonb"`

	Addi      types.JSONText `json:"addi,omitempty" remark:"附加信息" gorm:"type:jsonb"`
	Status    null.Int       `json:"status,omitempty" remark:"状态0未通知，2已通知" gorm:"type:smallint;default:0"`
	CreatedBy null.Int       `json:"created_by,omitempty" remark:"创建者" gorm:"type:bigint"`
	CreatedAt time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
	UpdatedAt time.Time      `json:"updated_at,omitempty" remark:"更新时间" gorm:"default:current_timestamp"`
}

//全校实时课表原始数据
type TRawCourse struct {
	ID int64 `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`

	//唯一索引
	KchID string `json:"kch_id" gorm:"type:varchar"`
	Cdbh  string `json:"cdbh" gorm:"type:varchar"`
	JghID string `json:"jgh_id" gorm:"type:varchar"`
	JxbID string `json:"jxb_id" gorm:"type:varchar"`
	Xn    string `json:"xn" gorm:"type:varchar"`
	Qsjsz string `json:"qsjsz" gorm:"type:varchar"`
	Skjc  string `json:"skjc" gorm:"type:varchar"`
	Xqj   int    `json:"xqj" gorm:"type:int"`

	Jxbmc   string  `json:"jxbmc" gorm:"type:varchar"`
	Jgh     string  `json:"jgh" gorm:"type:varchar"`
	Kch     string  `json:"kch" gorm:"type:varchar"`
	Cdlbmc  string  `json:"cdlbmc" gorm:"type:varchar"`
	Cdmc    string  `json:"cdmc" gorm:"type:varchar"`
	Cdqsjsz string  `json:"cdqsjsz" gorm:"type:varchar"`
	Cdskjc  string  `json:"cdskjc" gorm:"type:varchar"`
	Jslxdh  string  `json:"jslxdh" gorm:"type:varchar"`
	Jsxy    string  `json:"jsxy" gorm:"type:varchar"`
	Jxbrs   int     `json:"jxbrs" gorm:"type:int"`
	Jxbzc   string  `json:"jxbzc" gorm:"type:varchar"`
	Jxdd    string  `json:"jxdd" gorm:"type:varchar"`
	Jxlmc   string  `json:"jxlmc" gorm:"type:varchar"`
	Kcmc    string  `json:"kcmc" gorm:"type:varchar"`
	Kcxzmc  string  `json:"kcxzmc" gorm:"type:varchar"`
	KkbmID  string  `json:"kkbm_id" gorm:"type:varchar"`
	Kkxy    string  `json:"kkxy" gorm:"type:varchar"`
	Rwzxs   string  `json:"rwzxs" gorm:"type:varchar"`
	Sksj    string  `json:"sksj" gorm:"type:varchar"`
	Xbmc    string  `json:"xbmc" gorm:"type:varchar"`
	Xf      float64 `json:"xf" gorm:"type:numeric(5,2)"`
	Xkrs    int     `json:"xkrs" gorm:"type:int"`
	Xm      string  `json:"xm" gorm:"type:varchar"`
	Xnm     string  `json:"xnm" gorm:"type:varchar"`
	Xq      string  `json:"xq" gorm:"type:varchar"`
	XqhID   string  `json:"xqh_id" gorm:"type:varchar"`
	Xqm     string  `json:"xqm" gorm:"type:varchar"`
	Xqmc    string  `json:"xqmc" gorm:"type:varchar"`
	Zcmc    string  `json:"zcmc" gorm:"type:varchar"`
	Zgxl    string  `json:"zgxl" gorm:"type:varchar"`
	Zhxs    string  `json:"zhxs" gorm:"type:varchar"`
	Zjxh    int     `json:"zjxh" gorm:"type:int"`
	Zyzc    string  `json:"zyzc" gorm:"type:varchar"`
	Zcd     int     `json:"zcd" gorm:"type:int"`
	Jc      int     `json:"jc" gorm:"type:int"`
	Cdjc    int     `json:"cdjc" gorm:"type:int"`
	Zws     int     `json:"zws" gorm:"type:int"`
	Lch     int     `json:"lch" gorm:"type:int"`

	CreatedAt time.Time `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
}

type TTeachEvaluation struct {
	ID int64 `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`

	CourseID  string `json:"course_id,omitempty" remark:"课程号" gorm:"type:varchar;not null"`
	TeacherID string `json:"teacher_id,omitempty" remark:"教师工号" gorm:"type:varchar;not null"`

	Addi      types.JSONText `json:"addi,omitempty" remark:"附加信息" gorm:"type:jsonb"`
	Status    null.Int       `json:"status,omitempty" remark:"状态" gorm:"type:smallint;default:0"`
	CreatedBy null.Int       `json:"created_by,omitempty" remark:"创建者" gorm:"type:bigint"`
	CreatedAt time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
	UpdatedAt time.Time      `json:"updated_at,omitempty" remark:"更新时间" gorm:"default:current_timestamp"`
}
