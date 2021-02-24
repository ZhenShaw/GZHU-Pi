import requests
import time
from lxml import html
import pandas as pd
import json
import pprint

url = {
    "login": "https://cas.gzhu.edu.cn/cas_server/login?service=http%3A%2F%2Fjwxt.gzhu.edu.cn%2Fsso%2Flyiotlogin",
    "all-course": "http://jwxt.gzhu.edu.cn/jwglxt/design/funcData_cxFuncDataList.html?func_widget_guid=DA1B5BB30E1F4CB99D1F6F526537777B&gnmkdm=N219904"
}


class JW(object):

    def __init__(self, username, password):
        self.username = username
        self.password = password
        self.client = requests.session()
        self.headers = {
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 "
                          "(KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"
        }
        self.login()

    # 登录
    def login(self):
        get_res = self.client.get(url["login"], headers=self.headers)
        selector = html.fromstring(get_res.text)  # 将html文件转换为xpath可以识别的结构
        target = selector.xpath('//div[@class="row btn-row"]/input/@value')
        lt = target[0]
        execution = target[1]
        form_data = {
            "username": self.username,
            "password": self.password,
            "captcha": "",
            "warn": "true",
            "lt": lt,
            "execution": execution,
            "_eventId": "submit",
            "submit": "登录"
        }

        res = self.client.post(
            url["login"], data=form_data, headers=self.headers)
        if "账号或密码错误" in res.text:
            return 0
        else:
            return 1

    # 查询全校课表
    def get_all_course(self, page='1', count="15"):

        post_data = {
            'xnm': '2019',
            'xqm': '3',
            '_search': 'false',
            'nd': int(round(time.time() * 1000)),
            'queryModel.showCount': count,
            'queryModel.currentPage': page,
            'queryModel.sortName': '',
            'queryModel.sortOrder': 'asc'
        }
        res = self.client.post(url=url["all-course"], data=post_data, headers=self.headers)
        course_json = json.loads(res.text)["items"]

        return course_json

    # 爬取全校课表，利用pandas导出CSV
    def export_all_course(self):
        import pandas as pd
        raw_list = self.get_all_course(count="10000")
        table = pd.DataFrame(raw_list)
        pd.DataFrame(table).to_csv('2019-2020-1.csv')
        print("文件导出完成")

    # 获取全部学院及其编号
    def get_all_college(self):
        res = self.client.get(
            "http://jwxt.gzhu.edu.cn/jwglxt/design/viewFunc_cxDesignFuncPageIndex.html?gnmkdm=N219904")
        selector = html.fromstring(res.text)
        college_id = selector.xpath('//*[@id="jg_id"]/option/@value')
        college_name = selector.xpath('//*[@id="jg_id"]/option')
        print(college_name[0].text)
        colleges = []
        for i, item in enumerate(college_id):
            college = {
                "jg_id": item,
                "jgmc": college_name[i].text
            }
            colleges.append(college)
        return colleges

    # 根据学院编号获取专业信息
    def get_all_major(self, jg_id="01"):
        res = self.client.get(
            "http://jwxt.gzhu.edu.cn/jwglxt/xtgl/comm_cxZydmList.html?jg_id=" + jg_id + "&gnmkdm=N219904")
        res_json = json.loads(res.text)
        majors = []
        for item in res_json:
            major = {
                "jg_id": item["jg_id"],
                "jgmc": item["jgmc"],
                "zyh": item["zyh"],
                "zyh_id": item["zyh_id"],
                "zymc": item["zymc"]
            }
            majors.append(major)
        return majors

    # 根据学院编号获取班级信息
    def get_all_class(self, jg_id="01"):
        res = self.client.get(
            "http://jwxt.gzhu.edu.cn/jwglxt/xtgl/comm_cxBjdmList.html?jg_id=" + jg_id + "&gnmkdm=N219904")
        res_json = json.loads(res.text)
        classes = []
        for item in res_json:
            del item["queryModel"]
            del item["userModel"]
            del item["rangeable"]
            del item["totalResult"]
            del item["pageable"]
            del item["listnav"]
            del item["localeKey"]
            del item["jgpxzd"]
            tmp_class = item
            classes.append(tmp_class)
        pprint.pprint(classes)
        return classes


# 爬取全校 学院-专业-班级 信息
def get_college_major_class():
    spider = JW('1706300042', '059055')
    spider.login()
    colleges = spider.get_all_college()

    majors = []
    classes = []
    for item in colleges:
        print(item["jg_id"])
        if item["jg_id"] == "":
            continue
        major = spider.get_all_major(item["jg_id"])
        majors = majors + major

        tmp_class = spider.get_all_class(item["jg_id"])
        classes = classes + tmp_class
    # 匹配专业名称
    for i, item in enumerate(classes):
        for item2 in majors:
            try:
                if item["zyh_id"] == item2["zyh_id"]:
                    classes[i]["zymc"] = item2["zymc"]
            except:
                pass
    raw_list = classes
    table = pd.DataFrame(raw_list)
    pd.DataFrame(table).to_csv('class.csv')

