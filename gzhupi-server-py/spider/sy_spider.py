import requests
from spider.sy_handler import *

url = {
    "login": "http://202.192.67.23/",
    "home": "http://202.192.67.23//Index.aspx",
    "exp_list": "http://202.192.67.23//_GD002_Teaching_Teaching/Teaching_SYAdvanceStudentInfoList.aspx"
}

# url = {
#     "login": "http://202.192.67.23/",
#     "home": "http://202.192.67.23/Index.aspx",
#     "exp_list": "http://202.192.67.23/_GD002_Teaching_Teaching/Teaching_SYAdvanceStudentInfoList.aspx"
# }


# 物理实验爬虫
class SY(object):

    def __init__(self, username, password):
        self.username = username
        self.password = password
        self.client = requests.session()
        self.headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 '
                          '(KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36'
        }

    # 登录
    def login(self):
        get_res = self.client.get(url=url["login"], headers=self.headers)
        form_data = get_login_form(get_res.text, self.username, self.password)

        self.client.post(url=url["login"], data=form_data, headers=self.headers)
        # 获取新cookie
        res = self.client.get(url=url["home"], headers=self.headers)

        if "验证码" in res.text:
            return 0  # 登录失败
        else:
            return 1

    # 获取物理实验信息
    def get_experiment(self):
        res = self.client.get(
            url=url["exp_list"],
            headers=self.headers)

        row_data = re.findall(r'padding-bottom:5px;">(.+?)</td>', res.content.decode())

        return get_experiment(row_data)
