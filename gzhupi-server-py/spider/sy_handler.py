from lxml import html
import re
import requests
import os
import json


def get_login_form(text, username, password):
    """
    获取登录表单
    :param text: 登录页面html文本
    :param username: 用户名
    :param password: 密码
    :return: 用于POST的登录表单
    """

    selector = html.fromstring(text)  # 将html文件转换为xpath可以识别的结构
    form_data = {
        'txtUserName': username,
        'txtPSW': password,
        '__EVENTTARGET': '',
        '__EVENTARGUMENT': '',
        '__ASYNCPOST': 'true',
        'btnLogin': '',
        'NewPassWord1': '',
        'NewPassWord2': '',
        'ASPxTimer1S': '1;1000',
        '__VIEWSTATE': selector.xpath('//*[@id="__VIEWSTATE"]/@value')[0],
        '__VIEWSTATEGENERATOR': 'CA0B0334',
        '__EVENTVALIDATION': selector.xpath('//*[@id="__EVENTVALIDATION"]/@value')[0],
        'hfLoginState$I': '12|#|state|4|2|10#',
        'txtVF': selector.xpath('//*[@id="lbRandom"]/text()')[0],
        'ScriptManager1': 'ScriptManager1|btnLogin',
        'DXScript': '1_141,1_79,1_114,1_134,1_137,1_77,1_127,1_125,1_90',
        'MessageBox1_ASPxPCAlertWS': '0:0:-1:-10000:-10000:0:-10000:-10000:1:0:0:0;0:0:-1:-10000:-10000'
                                     ':0:-10000:-10000:1:0:0:0;0:0:-1:-10000:-10000:0:-10000:-10000:1:0:0:0',
    }
    return form_data


def sortByDate(Date):
    i = 1
    # 选择排序法
    for a in range(len(Date) - 1):
        for each in range(len(Date) - i):
            if int(Date[each]['date'].replace('-', '')) > int(Date[each + 1]['date'].replace('-', '')):
                temp = Date[each]
                Date[each] = Date[each + 1]
                Date[each + 1] = temp
        i += 1
    return (Date)


def numtran(ChStr):
    if ChStr == '一':
        return 1
    elif ChStr == '二':
        return 2
    elif ChStr == '三':
        return 3
    elif ChStr == '四':
        return 4
    elif ChStr == '五':
        return 5
    elif ChStr == '六':
        return 6
    elif ChStr == '日':
        return 7


# 整理物理实现信息
def get_experiment(data):
    exp_list = []

    # 去除多余字符
    while '&nbsp;' in data:
        data.remove('&nbsp;')
    for each in range(len(data)):
        data[each] = data[each].replace(r'<br/>', '')
    # print(data)
    if (data != []):
        student_info = {
            "name": data[3],
            "student_id": data[2],
            "year": '',
            "college": '',
            "major": '',
            "major_class": data[4]
        }
        set_log(student_info, api_type="物理实验")
        # print(student_info)
    for each in range(int(len(data) / 9)):
        exp_list.append({})
        exp_list[each]['class_place'] = data[6 + 9 * each]
        exp_list[each]['color'] = 20
        exp_list[each]['course_name'] = re.findall(r'\)(.+)\[', data[0 + 9 * each])[0]
        exp_list[each]['course_id'] = re.findall(r'\((\d+)\)', data[0 + 9 * each])[0]
        sy_contin_time = re.findall(r'星期./(.+?节)', data[8 + 9 * each])[0]
        exp_list[each]['course_time'] = sy_contin_time
        exp_list[each]['last'] = int(re.findall(r'-(.+)节', sy_contin_time)[0]) - int(
            re.findall(r'(.+)-', sy_contin_time)[0]) + 1
        exp_list[each]['start'] = int(re.findall(r'(.+)-', sy_contin_time)[0])
        exp_list[each]['teacher'] = data[5 + 9 * each]
        exp_list[each]['weeks'] = re.findall(r'(.+周)', data[8 + 9 * each])[0]
        exp_list[each]['which_day'] = re.findall(r'(星期.)', data[8 + 9 * each])[0]
        exp_list[each]['weekday'] = numtran(re.findall(r'星期(.)', data[8 + 9 * each])[0])
        exp_list[each]['type'] = "exp"  # 课表类型
        exp_list[each]['date'] = data[7 + 9 * each]

    return sortByDate(exp_list)


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
