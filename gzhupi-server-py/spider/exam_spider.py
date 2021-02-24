from spider.exam_handler import *
import requests


# 考试相关查询爬虫

class Exam(object):

    def __init__(self):
        self.url = {
            'ch_test': 'http://www.cltt.org/StudentScore/ScoreResult',
            'admit_query': 'http://zsjy.gzhu.edu.cn/gklqcxjgy.jsp?wbtreeid=1080'
        }
        self.client = requests.session()

    # 广州大学高考录取查询
    def admit_query(self, stu_id, stu_name):
        post_data = {
            'stuID1': stu_id,
            'stuName1': stu_name
        }
        res = self.client.post(self.url["admit_query"], post_data)
        get_result = re.findall(r'align="left" >(.+?)</td>', res.text)
        try:
            admit_result = {
                'stu_id': get_result[0],
                'stu_name': get_result[1],
                'major': get_result[2]
            }
        except:
            admit_result = {}
        return admit_result

    # 普通话水平测试查询
    def ch_test_query(self, post_data):
        res = self.client.post(
            url=self.url['ch_test'], data=post_data)
        if ('对不起没有查询到相关信息' in res.text):
            return '对不起没有查询到相关信息'
        else:
            return ch_test_handler(res)

    # 四六级获取验证码图片
    def cetTestQueryGetImg(self, id_num, name):
        imgCookiesData = {}
        imgCookiesData['img'] = get_img(self.client, id_num)
        imgCookiesData['cookies'] = requests.utils.dict_from_cookiejar(self.client.cookies)
        return imgCookiesData

    # 四六级获取分数
    def cetTestQueryGetScore(self, id_num, name, capcha, cookies):
        return get_score(self.client, id_num, name, capcha, cookies)

    # 四六级获取验证码图片
    def cet_get_captcha(self, id_num, name):
        return get_img(self.client, id_num)

    # 四六级获取分数
    def cet_get_score(self, id_num, name, capcha):
        return get_score(self.client, id_num, name, capcha)


# 普通话考试测试
# test = Exam()
# testData = {
#     'name': '***REMOVED***',
#     'stuID': '',
#     'idCard': '***REMOVED***'
# }
# print(test.ch_test_query(testData))

# #cet考试测试
# test=EX()
# print(test.cetTestQueryGetImg('***REMOVED***','***REMOVED***'))
# capcha=input()

# cookiesTest={'BIGipServercache.neea.edu.cn_pool': '2543896586.39455.0000'}
# test1=Exam()
# print(test1.cetTestQueryGetScore('***REMOVED***','***REMOVED***',cookiesTest))

'''
#录取查询
test=EX()
test.admitQuery('***REMOVED***','***REMOVED***')
'''
