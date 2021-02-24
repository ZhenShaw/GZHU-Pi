import requests
from spider.jw_handler import *
import time
import os

url = {
    "login": "https://cas.gzhu.edu.cn/cas_server/login?service=http%3A%2F%2Fjwxt.gzhu.edu.cn%2Fsso%2Flyiotlogin",
    "info": "http://jwxt.gzhu.edu.cn/jwglxt/xsxxxggl/xsgrxxwh_cxXsgrxx.html?gnmkdm=N100801&layout=default",
    "course": "http://jwxt.gzhu.edu.cn/jwglxt/kbcx/xskbcx_cxXsKb.html?gnmkdm=N2151",
    "grade": "http://jwxt.gzhu.edu.cn/jwglxt/cjcx/cjcx_cxDgXscj.html?doType=query&gnmkdm=N100801",
    "exam": "http://jwxt.gzhu.edu.cn/jwglxt/kwgl/kscx_cxXsksxxIndex.html?doType=query&gnmkdm=N358105",
    "id-credit": "http://jwxt.gzhu.edu.cn/jwglxt/xsxxxggl/xsxxwh_cxXsxkxx.html?gnmkdm=N100801",  # 获取课程学分
    "empty-room": "http://jwxt.gzhu.edu.cn/jwglxt/cdjy/cdjy_cxKxcdlb.html?doType=query&gnmkdm=N2155",
    "all-course": "http://jwxt.gzhu.edu.cn/jwglxt/design/funcData_cxFuncDataList.html?func_widget_guid=DA1B5BB30E1F4CB99D1F6F526537777B&gnmkdm=N219904"
}

"""

教务系统爬虫类

"""


class JW(object):

    def __init__(self, username, password):
        self.username = username
        self.password = password
        self.client = requests.session()
        self.headers = {
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 "
                          "(KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"
        }

    # 登录
    def login(self):
        get_res = self.client.get(url["login"], headers=self.headers)
        form_data = get_login_form(get_res.text, self.username, self.password)

        res = self.client.post(
            url["login"], data=form_data, headers=self.headers)

        if "账号或密码错误" in res.text:
            return 0
        else:
            return 1

    # 获取学生信息
    def get_info(self):
        get_res = self.client.get(url["info"], headers=self.headers)
        return get_student_info(get_res.text)

    # 获取课表
    def get_course(self, year_sem="2019-2020-1"):
        """
        获取课表
        :param year_sem: 学年-学期
        :return: 处理后的课表数据
        """
        # 学年段的第一年
        year = year_sem.split("-")[0]
        # 第一学期：3，第二学期：12
        if year_sem.split("-")[2] == "1":
            semester = "3"
        else:
            semester = "12"

        data = {
            "xnm": year,
            "xqm": semester,
            "queryModel.showCount": 50
        }
        res = self.client.post(url["course"], data=data, headers=self.headers)
        course = get_course(res.text)

        credit = self.client.post(
            url["id-credit"], data=data, headers=self.headers)
        course = add_credit(credit.text, course)
        course["update_time"] = time.strftime(" %Y-%m-%d %H:%M:%S")
        try:
            set_log(self.get_info(), "课表查询")
        except:
            pass
        return course

    # 获取成绩
    def get_grade(self):
        data = {
            "xh_id": self.username,
            "xnm": "",
            "xqm": "",
            "_search": "false",
            "nd": int(round(time.time() * 1000)),
            "queryModel.showCount": 150,
            "queryModel.currentPage": 1,
            "queryModel.sortName": "",
            "queryModel.sortOrder": "asc",
            "time": 0,
        }
        res = self.client.post(url["grade"], data=data, headers=self.headers)
        try:
            set_log(self.get_info(), "成绩查询")
        except:
            pass
        return get_grade(res.text, self.password)

    def get_exam(self, year_sem="2019-2020-1"):
        """
        获取考试信息
        :param year_sem: 学年-学期
        :return: 处理后的考试数据
        """
        # 学年段的第一年
        year = year_sem.split("-")[0]
        # 第一学期：3，第二学期：12
        if year_sem.split("-")[2] == "1":
            semester = "3"
        else:
            semester = "12"

        data = {
            "xnm": year,
            "xqm": semester,
            "ksmcdmb_id": "",
            "kch": "",
            "kc": "",
            "ksrq": "",
            "_search": False,
            "nd": int(round(time.time() * 1000)),  # 当前时间戳,
            "queryModel.showCount": 15,
            "queryModel.currentPage": 1,
            "queryModel.sortName": "",
            "queryModel.sortOrder": "asc",
            "time": 0
        }
        res = self.client.post(url["exam"], data=data, headers=self.headers)

        try:
            set_log(self.get_info(), "考试查询")
        except:
            pass

        return get_exam(res.text)

    # 查询空教室

    def get_empty_room(self, request):
        # 处理表单参数
        post_data = empty_room_form_handle(request)

        res = self.client.post(
            url=url["empty-room"], data=post_data, headers=self.headers)
        return get_empty_room(res.text)

    # 查询全校课表
    def get_all_course(self, request_form, page='1'):
        post_data = all_course_form_handle(request_form, page)
        res = self.client.post(url=url["all-course"], data=post_data, headers=self.headers)
        course_data = get_all_course(res.text)

        return course_data


# 把API请求记录写入数据库
def set_log(student_info, api_type="其它"):
    """
    把API请求记录写入知晓云
    :param student_info: 学生基础信息
    :param api_type: API请求类型
    :return: 状态码201为写入成功
    """

    student_info["api_type"] = api_type
    # token有效期至2020年2月1号，从环境变量读取
    token = os.getenv('minapp_token')
    if token == None:
        token = "please set token to environment value"
    api_url = "https://cloud.minapp.com/oserve/v1/table/65445/record/"
    headers = {
        "Authorization": "Bearer " + token,
        "Content-Type": 'application/json'
    }
    data = json.dumps(student_info)
    res = requests.post(url=api_url, data=data, headers=headers)
    return res.status_code  # 201为写入成功
