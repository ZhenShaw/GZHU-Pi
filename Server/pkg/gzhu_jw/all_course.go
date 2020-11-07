package gzhu_jw

import (
	"GZHU-Pi/env"
	"fmt"
	"github.com/astaxie/beego/logs"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RawCourse struct {
	env.TRawCourse
}

//查询全校课表
func (c *JWClient) SearchAllCourse(xnm, xqm string, page, count int) (data []RawCourse, csvData []byte, err error) {

	if xnm == "" {
		year := time.Now().Year()
		month := time.Now().Month()
		if month < 8 {
			year = year - 1
		}
		xnm = fmt.Sprint(year)
	}

	if page <= 0 {
		page = 1
	}
	if count <= 0 {
		count = 15
	}

	if xqm == "1" || xqm == "3" {
		xqm = "3"
	} else {
		xqm = "12"
	}

	nd := time.Now().Unix() * 1000 //时间戳
	var form = url.Values{
		"xnm":                    {xnm}, //2019
		"xqm":                    {xqm}, //3 是第一学期，12 是第二学期
		"_search":                {"false"},
		"nd":                     {strconv.Itoa(int(nd))},
		"queryModel.showCount":   {strconv.Itoa(count)},
		"queryModel.currentPage": {strconv.Itoa(page)},
		"queryModel.sortName":    {""},
		"queryModel.sortOrder":   {"asc"},
	}

	resp, err := c.doRequest("POST", Urls["all-course"], urlencodedHeader, strings.NewReader(form.Encode()))
	if err != nil {
		logs.Error(err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	//检查登录状态
	if strings.Contains(string(body), "登录") {
		return nil, nil, AuthError
	}

	type Data struct {
		Items []RawCourse `json:"items"`
	}
	var d Data

	json1 := jsoniter.ConfigCompatibleWithStandardLibrary
	err = json1.Unmarshal(body, &d)
	if err != nil {
		logs.Error(err)
		return
	}
	data = d.Items

	csvData = ToCsvFormat(data)

	return
}

func ToCsvFormat(all []RawCourse) (data []byte) {

	header := []string{"cdbh", "cdlbmc", "cdmc", "cdqsjsz", "cdskjc", "jgh", "jgh_id", "jslxdh",
		"jsxy", "jxb_id", "jxbmc", "jxbrs", "jxbzc", "jxdd", "jxlmc", "kch", "kch_id", "kcmc",
		"kcxzmc", "kkbm_id", "kkxy", "qsjsz", "rwzxs", "skjc", "sksj", "xbmc", "xf", "xkrs",
		"xm", "xnm", "xn", "xq", "xqh_id", "xqj", "xqm", "xqmc", "zcmc", "zgxl", "zhxs",
		"zjxh", "zyzc", "zcd", "jc", "cdjc", "zws", "lch"}

	var lines []string
	lines = append(lines, strings.Join(header, ","))

	for _, v := range all {
		var values []string
		values = append(values, v.Cdbh, v.Cdlbmc, v.Cdmc, v.Cdqsjsz, v.Cdskjc, v.Jgh, v.JghID, v.Jslxdh,
			v.Jsxy, v.JxbID, v.Jxbmc, fmt.Sprint(v.Jxbrs), v.Jxbzc, v.Jxdd, v.Jxlmc, v.Kch, v.KchID, v.Kcmc, v.Kcxzmc,
			v.KkbmID, v.Kkxy, v.Qsjsz, v.Rwzxs, v.Skjc, v.Sksj, v.Xbmc, fmt.Sprint(v.Xf), fmt.Sprint(v.Xkrs), v.Xm, v.Xnm, v.Xn,
			v.Xq, v.XqhID, fmt.Sprint(v.Xqj), v.Xqm, v.Xqmc, v.Zcmc, v.Zgxl, v.Zhxs, fmt.Sprint(v.Zjxh),
			v.Zyzc, fmt.Sprint(v.Zcd), fmt.Sprint(v.Jc), fmt.Sprint(v.Cdjc), fmt.Sprint(v.Zws), fmt.Sprint(v.Lch))

		for k, v := range values {
			if strings.Contains(v, ",") {
				values[k] = fmt.Sprintf(`"%s"`, v) //去除字符串内部逗号对csv的影响
			}
		}

		csvLine := strings.Join(values, ",")
		lines = append(lines, csvLine)
	}
	csv := strings.Join(lines, "\n")

	data = []byte(csv)
	return
}
