/**
 * @File: dbInit
 * @Author: Shaw
 * @Date: 2020/5/3 5:36 PM
 * @Desc

 */

package env

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jmoiron/sqlx"
	"time"
)

const (
	vTopic = `
			create or replace VIEW v_topic as
			select t.*, u.open_id, u.gender,
				   --匿名头像
				   (CASE WHEN anonymous = true THEN u.profile_pic ELSE u.avatar END) as avatar,
				   --匿名昵称
				   (CASE WHEN anonymous = true THEN anonymity ELSE u.nickname END) as nickname,
				   --留言数量
				   (select count(*) from t_discuss where object_id = t.id)         as discussed,
				   --点赞数量
				   (select count(*) from t_relation where object_id = t.id and object = 't_topic'
					  and type = 'star')                                           as liked,
				   --查询当前主题有关用户的点赞记录
				   (select json_agg(result)
					from (select r.*, t_user.nickname, t_user.avatar
						  from t_relation r, t_user where r.created_by = t_user.id
							and r.object = 't_topic' and r.type = 'star' ) result
					where object_id = t.id)                                        as star_list,
				   --查询当前主题有关用户的认领记录
				   (select json_agg(result)
					from (select r.*, t_user.nickname, t_user.avatar
						  from t_relation r, t_user where r.created_by = t_user.id
							and r.object = 't_topic' and r.type = 'claim' ) result
					where object_id = t.id)                                        as claim_list,
					--最新留言时间作为排序时间
       				(coalesce((select created_at from t_discuss d where d.object_id = t.id and object='t_topic'
					order by created_at desc limit 1), t.updated_at))              as order_time
			from t_topic as t, t_user as u where t.created_by = u.id;
			comment on view v_topic is '主题/帖子视图';
	`
	vGrade = `
			create or replace view v_grade (stu_id, class_id, major_class, major_id, major, stu_name, college_id, college, admit_year, year,
						semester, course_id, course_name, credit, grade_value, grade, course_gpa, course_type, exam_type, invalid, jxb_id,
						teacher, year_sem, created_at) as SELECT s.stu_id, s.class_id, s.major_class, s.major_id, s.major, s.stu_name,
						  s.college_id, s.college,s.admit_year, g.year, g.semester, g.course_id, g.course_name, g.credit,g.grade_value,
						  g.grade, g.course_gpa, g.course_type, g.exam_type, g.invalid,g.jxb_id, g.teacher, g.year_sem, g.created_at
			FROM t_stu_info s, t_grade g WHERE ((s.stu_id)::text = (g.stu_id)::text);
			comment on view v_grade is '学生成绩视图';
	`
	vDiscuss = `
			create or replace VIEW v_discuss as
			select d.*, u.open_id, u.gender,
				   (CASE WHEN anonymous = true THEN u.profile_pic ELSE u.avatar END) as avatar,
				   (CASE WHEN anonymous = true THEN anonymity ELSE u.nickname END) as nickname
			from t_discuss as d, t_user as u where d.created_by = u.id;
			comment on view v_discuss is '评论视图';
	`
	vUser = `
			create or replace view v_user as (
			select u.*, s.stu_name, s.admit_year, s.class_id, s.college, s.college_id, 
				s.major, s.major_class, s.major_id
			from t_user u left join t_stu_info s on u.stu_id = s.stu_id);
			comment on view v_discuss is '学生用户视图';
	`
	VTeachEvaluation = `
			create or replace VIEW v_teach_evaluation as
			select t.*,
				   coalesce(rc.course_name, t.addi ->> 'course_name'::varchar)                                AS course_name,
				   coalesce(rc.course_type, t.addi ->> 'course_type'::varchar)                                AS course_type,
				   rc.credit, rc.constitute, rc.period,
				   coalesce(rc.teacher, t.addi ->> 'teacher'::varchar)                                        AS teacher,
				   rc.sex, rc.campus, rc.college, rc.title, rc.education, rc.phone,
				   -- 评分统计
				   (select cast(avg(mark) AS decimal(5, 2)) from t_discuss where object_id = t.id and object = 't_teach_evaluation') as mark,
				   --留言数量
				   (select count(*) from t_discuss where object_id = t.id and object = 't_teach_evaluation')  as discussed,
				   --点赞数量
				   (select count(*) from t_relation where object_id = t.id and object = 't_teach_evaluation'
					  and type = 'star')                                                                      as liked,
				   --查询当前主题有关用户的点赞记录
				   (select json_agg(result) from ( select r.*, t_user.nickname, t_user.avatar from t_relation r, t_user
						  where r.created_by = t_user.id and r.object = 't_teach_evaluation' and r.type = 'star' limit 50) result
					where object_id = t.id)                                                                   as star_list,
					--最新留言时间作为排序时间
       				(coalesce((select created_at from t_discuss d where d.object_id = t.id and object='t_teach_evaluation'
					order by created_at desc limit 1), t.updated_at))              as order_time
			from t_teach_evaluation as t left join
				 -- rc表:根据 课程号+教工号 去重，选出教师课程信息
				 (select kch_id course_id,
						 kcmc   course_name,
						 kcxzmc course_type,
						 xf     credit,
						 zyzc   constitute,--专业组成
						 zhxs   period,--周学时
						 jgh_id teacher_id,
						 xm     teacher,
						 xbmc   sex,
						 xqmc   campus,--校区
						 jsxy   college,-- 教师学院
						 zcmc   title,--职称
						 zgxl   education,--最高学历
						 jslxdh phone
				  from t_raw_course where id in (select max(id) from t_raw_course group by kch_id, jgh_id)) as rc
			on t.course_id = rc.course_id and t.teacher_id = rc.teacher_id;
			comment on view v_topic is '教评视图';
	`
)

var db *gorm.DB
var sqlxDB *sqlx.DB

func InitDb() error {
	d := Conf.Db
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Dbname, d.SslMode)

	var err error
	db, err = gorm.Open("postgres", dbInfo)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("数据库：%s:%d/%s", d.Host, d.Port, d.Dbname)

	sqlxDB = sqlx.MustOpen("postgres", dbInfo)
	err = sqlxDB.Ping()
	if err != nil {
		logs.Error(err)
		return err
	}

	// SetMaxIdleCons 设置连接池中的最大闲置连接数。
	db.DB().SetMaxIdleConns(10)

	// SetMaxOpenCons 设置数据库的最大连接数量。
	db.DB().SetMaxOpenConns(5)

	// SetConnMaxLifetime 设置连接的最大可复用时间。
	db.DB().SetConnMaxLifetime(time.Hour)

	//关闭复数表名
	db.SingularTable(true)

	if Conf.App.InitModels {
		t := time.Now()
		modelsInit()
		logs.Info("init models in:", time.Since(t))
	}

	return nil
}

func GetGorm() *gorm.DB {
	return db
}

func GetSqlx() *sqlx.DB {
	return sqlxDB
}

func modelsInit() {
	logs.Info("models initializing ...")
	t := time.Now()
	//自动迁移 只会 创建表、缺失的列、缺失的索引，不会 更改现有列的类型或删除未使用的列
	e1 := db.AutoMigrate(&TStuInfo{}, &TGrade{}, &TApiRecord{}, &TUser{},
		&TTopic{}, &TDiscuss{}, &TRelation{}, &TNotify{}, &TRawCourse{}, &TTeachEvaluation{}).Error

	e2 := db.Model(&TGrade{}).AddUniqueIndex("t_grade_stu_id_course_id_jxb_id_idx",
		"stu_id", "course_id", "jxb_id").Error

	e3 := db.Model(&TRelation{}).AddUniqueIndex("t_relation_object_object_id_type_created_by_idx",
		"object", "object_id", "type", "created_by").Error

	e4 := db.Model(&TRawCourse{}).AddUniqueIndex("t_raw_course_unique_idx",
		"kch_id", "cdbh", "jgh_id", "jxb_id", "xn", "qsjsz", "skjc", "xqj").Error

	e5 := db.Model(&TTeachEvaluation{}).AddUniqueIndex("t_teach_evaluation_course_id_teacher_id",
		"course_id", "teacher_id").Error

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
		err := fmt.Errorf("初始化表失败：%v %v %v %v %v", e1, e2, e3, e4, e5)
		panic(err)
	}
	e1 = db.Exec(vTopic).Error
	e2 = db.Exec(vDiscuss).Error
	e3 = db.Exec(vUser).Error
	e4 = db.Exec(vGrade).Error
	e4 = db.Exec(VTeachEvaluation).Error

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		err := fmt.Errorf("初始化视图失败：%v %v %v %v", e1, e2, e3, e4)
		panic(err)
	}

	logs.Debug("models inited in", time.Since(t))
}
