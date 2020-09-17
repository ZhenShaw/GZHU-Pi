package gzhu_jw

import (
	"fmt"
	"net/http"
)

var (
	LoginError = fmt.Errorf("账号或密码错误，请重新输入，如需修改请前往 http://my.gzhu.edu.cn")
	AuthError  = fmt.Errorf("认证失败，可能是缓存失效，请重试")

	SemCode = []string{"3", "12"} //3是第一学期，12是第二学期
	Year    = "2020"
	FirstMonday="2020-09-14"

	//jsonHeader       = http.Header{"Content-Type": []string{"application/json"}}
	urlencodedHeader = http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}
)

const baseUrl = "http://jwxt.gzhu.edu.cn/jwglxt"

var Urls = map[string]string{
	//广大统一认证登录（get post）
	"jw-login": "https://cas.gzhu.edu.cn/cas_server/login?service=http://jwxt.gzhu.edu.cn/sso/lyiotlogin",
	//个人信息页面（get）
	"info": baseUrl + "/xsxxxggl/xsgrxxwh_cxXsgrxx.html?gnmkdm=N100801&layout=default",
	//选课-学生课表查询（post）
	"course": baseUrl + "/kbcx/xskbcx_cxXsKb.html?gnmkdm=N2151",
	//个人信息-选课信息（post）
	"id-credit": baseUrl + "/xsxxxggl/xsxxwh_cxXsxkxx.html?gnmkdm=N100801",
	//个人信息-成绩信息（post）
	"grade": baseUrl + "/cjcx/cjcx_cxDgXscj.html?doType=query&gnmkdm=N100801",
	//信息查询-考试信息查询（post）
	"exam": baseUrl + "/kwgl/kscx_cxXsksxxIndex.html?doType=query&gnmkdm=N358105",
	//信息查询-空教室查询（post）
	"empty-room": baseUrl + "/cdjy/cdjy_cxKxcdlb.html?doType=query&gnmkdm=N2155",
	//信息查询-全校实时课表（post）
	"all-course": baseUrl + "/design/funcData_cxFuncDataList.html?func_widget_guid=DA1B5BB30E1F4CB99D1F6F526537777B&gnmkdm=N219904",
	//学业情况页面（get）
	"achieve-get": baseUrl + "/xsxy/xsxyqk_cxXsxyqkIndex.html?gnmkdm=N105515",
	//学业情况课程列表接口（post）
	"achieve-post": baseUrl + "/xsxy/xsxyqk_cxJxzxjhxfyqKcxx.html?gnmkdm=N105515",
}

var WeekdayMatcher = map[string]int64{
	"星期一": 1,
	"星期二": 2,
	"星期三": 3,
	"星期四": 4,
	"星期五": 5,
	"星期六": 6,
	"星期日": 7,
	"星期天": 7,
}
