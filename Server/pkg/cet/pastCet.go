/**
 * @File: pastCet
 * @Author: Crayon
 * @Date: 2020/10/24 22:52 PM
 * @Desc

 */

package cet

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
)

// 历次四六级成绩

type PastCetClient struct {
	Client    *http.Client     `json:"-"`
	Captcha   string           `json:"-"`          // 验证码
	Subject   string           `json:"subject"`    // 考试类别 eg：CET4
	PersonID  string           `json:"person_id"`  // 身份证
	Name      string           `json:"name"`       // 姓名
	ScoreList []*PastCetDetail `json:"score_list"` // 历次考试详细信息列表
}

type PastCetDetail struct {
	School        string `json:"school"`          // 学校
	TestID        string `json:"test_id"`         // 准考证号
	Total         int    `json:"total"`           // 总分
	Listening     int    `json:"listening"`       // 听力
	Reading       int    `json:"reading"`         // 阅读
	Writing       int    `json:"writing"`         // 写作
	VoiceID       string `json:"voice_id"`        // 口语考试准考证号
	Voice         string `json:"voice"`           // 口语
	ScoreID       string `json:"score_id"`        // 成绩单编号
	ExamDate      string `json:"exam_date"`       // 笔试考试日期
	ExamVoiceDate string `json:"exam_voice_date"` // 口语考试日期
}

// 请求头设置
var pastCetHeader = http.Header{
	"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36"},
	"Referer":    []string{"http://cjcx.neea.edu.cn/html1/folder/20051/1156-1.htm"},
}

func NewPastCetClient() *PastCetClient {
	cookieJar, _ := cookiejar.New(nil)
	var c = &PastCetClient{}
	c.Client = &http.Client{
		Jar:       cookieJar,
		Timeout:   time.Minute,
		Transport: transport,
	}
	return c
}

func (this *PastCetClient) doRequest(method, url string, header http.Header, body io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	t1 := time.Now()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	if cookies != nil {
		this.Client.Jar.SetCookies(req.URL, cookies)
	}
	resp, err := this.Client.Do(req)

	logs.Debug("请求耗时：", time.Since(t1), url)
	return resp, err
}

func (this *PastCetClient) GetPastCetDetail(token string) (err error) {
	baseUrl := "http://appquery.neea.edu.cn/api/result/data"
	params := url.Values{
		"token": []string{token},
	}
	fullUrl := baseUrl + "?" + params.Encode()
	resp, err := this.doRequest("GET", fullUrl, pastCetHeader, nil, nil)
	if err != nil {
		logs.Error(err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	text := string(body)
	reg := regexp.MustCompile(`"data":(\{.*?\})`)
	match := reg.FindStringSubmatch(text)
	res := strings.Split(match[1], ",")

	detail := &PastCetDetail{}
	for _, v := range res {
		sp := strings.Split(v, ":")
		if len(sp) != 2 {
			err = fmt.Errorf("获取成绩信息失败: %s", v)
			logs.Error(err)
			return
		}
		val := strings.ReplaceAll(sp[1], `"`, "")
		switch {
		case strings.Contains(v, "SN"):
			this.Subject = val
		// 考场学校名称
		case strings.Contains(v, "KS_SSXXMC"):
			detail.School = val
		// 准考证号
		case strings.Contains(v, "ZKZH"):
			detail.TestID = val
		// 总分
		case strings.Contains(v, "SCORE"):
			detail.Total, err = strconv.Atoi(val)
		// 听力
		case strings.Contains(v, "SCO_LC"):
			detail.Listening, err = strconv.Atoi(val)
		// 阅读
		case strings.Contains(v, "SCO_RD"):
			detail.Reading, err = strconv.Atoi(val)
		// 写作
		case strings.Contains(v, "SCO_WT"):
			detail.Writing, err = strconv.Atoi(val)
		// 口语准考证号
		case strings.Contains(v, "KY_ZKZ"):
			detail.VoiceID = val
		// 口语
		case strings.Contains(v, "KY_SCO"):
			detail.Voice = val
		// 成绩单编号
		case strings.Contains(v, "ID"):
			detail.ScoreID = val
		// 考试时间
		case strings.Contains(v, "EXAM_DT"):
			detail.ExamDate = val
		// 口语考试时间
		case strings.Contains(v, "EXAM_KY_DT"):
			detail.ExamVoiceDate = val
		}
		if err != nil {
			logs.Error(err)
			return
		}
	}
	this.ScoreList = append(this.ScoreList, detail)
	return
}

// 获取历次考试成绩
func (this *PastCetClient) GetPastCetInfo(captchaCookie *http.Cookie) (err error) {
	baseUrl := "http://appquery.neea.edu.cn/api/result/list"
	params := url.Values{
		"verify":  []string{this.Captcha},
		"subject": []string{this.Subject},
		"xm":      []string{this.Name},
		"sfz":     []string{this.PersonID},
	}
	fullUrl := baseUrl + "?" + params.Encode()
	resp, err := this.doRequest("GET", fullUrl, pastCetHeader, nil, []*http.Cookie{captchaCookie})
	if err != nil {
		logs.Error(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	text := string(body)
	// 去空白
	reg := regexp.MustCompile(`\s+`)
	text = reg.ReplaceAllString(text, "")
	reg = regexp.MustCompile(`"data":(\[.*\])`)
	match := reg.FindStringSubmatch(text)
	if len(match) != 2 || match[1] == "" {
		reg = regexp.MustCompile(`"message":"(.*?)"`)
		message := reg.FindStringSubmatch(text)
		err = fmt.Errorf("成绩查询失败 failed: %s", message[1])
		logs.Error(err)
		return
	}
	reg = regexp.MustCompile(`"token":"(.*?)"`)
	tokenList := reg.FindAllStringSubmatch(match[1], -1)
	for _, v := range tokenList {
		this.GetPastCetDetail(v[1])
	}

	return
}
