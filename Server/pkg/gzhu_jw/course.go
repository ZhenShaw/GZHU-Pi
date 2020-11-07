package gzhu_jw

import (
	"GZHU-Pi/env"
	"github.com/astaxie/beego/logs"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type CourseData struct {
	CourseList    []*env.TStuCourse `json:"course_list"  remark:"课程列表"`
	SjkCourseList []*SjkCourse      `json:"sjk_course_list" remark:"实践课"`
}

//type Course env.TStuCourse

type SjkCourse struct {
	SjkCourseName string `json:"sjk_course_name" remark:"实践课名"`
	SjkTeacher    string `json:"sjk_teacher" remark:"教师姓名"`
	SjkWeeks      string `json:"sjk_weeks" remark:"上课周次"`
}

func (c *JWClient) GetCourse(year, semester string) (courseData *CourseData, err error) {

	var form = url.Values{
		"xnm":                  {year},
		"xqm":                  {semester},
		"queryModel.showCount": {"50"},
	}
	var body1, body2 []byte

	var wg = sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var resp *http.Response
		resp, err = c.doRequest("POST", Urls["course"], urlencodedHeader, strings.NewReader(form.Encode()))
		if err != nil {
			logs.Error(err)
			return
		}
		body1, _ = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		wg.Done()
	}()
	go func() {
		resp, err0 := c.doRequest("POST", Urls["id-credit"], urlencodedHeader, strings.NewReader(form.Encode()))
		if err0 != nil {
			logs.Error(err0)
			return
		}
		body2, _ = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		wg.Done()
	}()
	wg.Wait()

	//检查登录状态
	if strings.Contains(string(body1), "登录") || strings.Contains(string(body2), "登录") {
		return nil, AuthError
	}
	//匹配课程号和学分
	creditMatcher := MatchCredit(body2)
	courseData = &CourseData{
		CourseList:    ParseCourse(body1, creditMatcher),
		SjkCourseList: ParseSjk(body1),
	}
	return
}

//获取到的课程信息不包含学分，从另一个请求结果进行匹配
func MatchCredit(body []byte) (matcher map[string]float64) {

	matcher = make(map[string]float64)
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	item := json.Get(body, "items")

	//遍历所有课程，提取课程号-学分
	for i := 0; true; i++ {
		courseID := item.Get(i).Get("kch").ToString()
		if courseID == "" {
			break
		}
		credit := item.Get(i).Get("xf").ToFloat64()
		matcher[courseID] = credit
	}
	return
}

//解析提取课程信息
func ParseCourse(body []byte, matcher map[string]float64) (courses []*env.TStuCourse) {

	courses = []*env.TStuCourse{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	kbList := json.Get(body, "kbList")

	var idSet = make(map[string]int64) //课程id集合，去重
	//遍历所有课程
	for i := 0; true; i++ {
		c := &env.TStuCourse{}
		c.CourseID = kbList.Get(i).Get("kch_id").ToString()
		if c.CourseID == "" {
			break
		}
		c.CheckType = kbList.Get(i).Get("khfsmc").ToString()
		c.ClassPlace = kbList.Get(i).Get("cdmc").ToString()
		c.CourseName = kbList.Get(i).Get("kcmc").ToString()
		c.CourseTime = kbList.Get(i).Get("jc").ToString()
		c.TeacherID = kbList.Get(i).Get("jgh_id").ToString()
		c.Weeks = kbList.Get(i).Get("zcd").ToString()
		c.WhichDay = kbList.Get(i).Get("xqjmc").ToString()
		c.Teacher = kbList.Get(i).Get("xm").ToString()
		//处理过长的多个教师姓名
		if len(c.Teacher) > 13 {
			c.Teacher = c.Teacher[:10] + "..."
		}
		//星期匹配为对应的数字
		c.Weekday = WeekdayMatcher[c.WhichDay]

		//提取并处理开始上课及持续节次
		reg, _ := regexp.Compile(`\d+`)
		res := reg.FindAllString(c.CourseTime, 2)
		if len(res) == 2 {
			c.Start, _ = strconv.ParseInt(res[0], 10, 64)
			end, _ := strconv.ParseInt(res[1], 10, 64)
			c.Last = end - c.Start + 1
		} else if len(res) == 1 {
			c.Start, _ = strconv.ParseInt(res[0], 10, 64)
			c.Last = 1
		}

		//通过集合为相同的课程分配唯一颜色id
		var ok bool
		c.Color, ok = idSet[c.CourseID]
		if !ok {
			idSet[c.CourseID] = int64(len(idSet))
			c.Color = idSet[c.CourseID]
		}
		//匹配课程学分
		c.Credit = matcher[c.CourseID]
		//生成周段数组
		c.WeekSection = WeekHandle(c.Weeks)

		courses = append(courses, c)
	}
	return
}

//解析提取实践课程信息
func ParseSjk(body []byte) (courses []*SjkCourse) {

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	kbList := json.Get(body, "sjkList")

	for i := 0; true; i++ {
		c := &SjkCourse{}
		//c.SjkCourseName = kbList.Get(i).Get("kcmc").ToString()
		c.SjkCourseName = kbList.Get(i).Get("sjkcgs").ToString()
		if c.SjkCourseName == "" {
			break
		}
		c.SjkTeacher = kbList.Get(i).Get("xm").ToString()
		c.SjkWeeks = kbList.Get(i).Get("qsjsz").ToString()

		courses = append(courses, c)
	}
	return
}

//周段提取为[start end start end]  如"4周,8-12周(单),14-16周" -> [4 4 9 9 11 11 14 16]
func WeekHandle(weekStr string) (weeks []int) {

	weekSlice := strings.Split(weekStr, ",")

	r, _ := regexp.Compile(`\d+`)

	for _, v := range weekSlice {

		//上课周段[start end]
		var section []int

		//提取 如8-12周 -> [8 12]
		res := r.FindAllString(v, 2)
		var start, end int
		if len(res) == 2 {
			start, _ = strconv.Atoi(res[0])
			end, _ = strconv.Atoi(res[1])
		} else if len(res) == 1 {
			start, _ = strconv.Atoi(res[0])
			end = start
		} else {
			start = 1
			end = 20
			logs.Error("match week failed %s", v)
			continue
		}
		section = []int{start, end}

		//特殊处理单双周 如8-12周(单) -> [9 9 11 11]
		if strings.Contains(v, "单") {
			section = []int{}
			for i := start; i <= end; i++ {
				if i%2 == 1 {
					section = append(section, i)
					section = append(section, i)
				}
			}
		}
		if strings.Contains(v, "双") {
			section = []int{}
			for i := start; i <= end; i++ {
				if i%2 == 0 {
					section = append(section, i)
					section = append(section, i)
				}
			}
		}
		weeks = append(weeks, section...)
	}
	return weeks
}
