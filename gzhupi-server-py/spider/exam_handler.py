import re
import random
from urllib.parse import urlencode
import requests


# 处理普通话考试得到的数据
def ch_test_handler(res):
    # 获取未处理的目标数据
    meg_list = re.findall(r'<span>(.*?)</span>', res.text)
    meg_list.append(re.findall(r'您共有<em class="clor10">(\d)</em>次考试成绩记录', res.text)[0])
    # 合成字典
    result = {
        'kssj': meg_list[0].split('：')[-1],
        'cssf': meg_list[1].split('：')[-1],
        'csz': meg_list[2].split('：')[-1],
        'name': meg_list[3],
        'score': meg_list[4],
        'sex': meg_list[5],
        'level': meg_list[6],
        'zkzh': meg_list[7],
        'zsbh': meg_list[8],
        'id': meg_list[9],
        'kscs': meg_list[10],
        'image': re.findall(r'<img src="(.*?)"', res.text)[0]
    }

    return result


# 获取图片地址
def get_img(Session, id_numm):
    try:
        headers = {
            'Connection': 'keep - alive',
            'Host': 'cache.neea.edu.cn',
            'Referer': 'http://cet.neea.edu.cn/cet',
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3486.0 Safari/537.36',
        }
        Session.headers = headers
        get_url = 'http://cache.neea.edu.cn/Imgs.do?c=CET&ik='
        params = {
            'c': 'CET',
            'ik': id_numm,
            't': str(random.random())[0:-2]
        }
        print(params['ik'])
        get_url = get_url + params['ik'] + '&t=' + str(params['t'])
        response = Session.get(get_url, data=params)
        print(response.text)
        img_url = re.compile('"(.*?)"').findall(response.text)[0]
        # img = requests.get(img_url, timeout=None)
        '''with open('img.png', 'wb') as f:
            f.write(img.content)'''
        return img_url
    except Exception as e:
        return ''


# 获取分数
def get_score(Session, id_num, name, capcha, cookies):
    Session.cookies = requests.utils.cookiejar_from_dict(cookies)
    level = id_num[9]
    returnData = {}

    headers = {
        'Connection': 'keep - alive',
        'Host': 'cache.neea.edu.cn',
        'Origin': 'http://cet.neea.edu.cn',
        'Referer': 'http://cet.neea.edu.cn/cet',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3486.0 Safari/537.36',
    }

    query_url = "http://cache.neea.edu.cn/cet/query"

    test = {
        '1': 'CET4_182_DANGCI',
        '2': 'CET6_182_DANGCI',
    }
    data = {
        'data': test.get(level) + ',' + id_num + ',' + name,
        'v': capcha
    }
    data = urlencode(data)
    response = Session.post(query_url, data=data, headers=headers)
    if 'error' in response.text:
        e = re.compile("'error':'(.*?)'|error:'(.*?)'").findall(response.text)[0]
        if e is not None:
            # print(e)
            if '验证码错误' in e[1]:
                return '验证码错误'
            else:
                print(e)
                return '查询不到该人成绩'
    else:
        id_num = re.compile("z:'(.*?)'").findall(response.text)[0]
        name = re.compile("n:'(.*?)'").findall(response.text)[0]
        school = re.compile("x:'(.*?)'").findall(response.text)[0]
        score = re.compile("s:(.*?),").findall(response.text)[0]
        listening = re.compile("l:(.*?),").findall(response.text)[0]
        reading = re.compile("r:(.*?),").findall(response.text)[0]
        writing = re.compile("w:(.*?),").findall(response.text)[0]
        rank = re.compile("kys:'(.*?)'").findall(response.text)[0]
        if level == '1':
            returnData['level'] = '4'
        elif level == '2':
            returnData['level'] = '6'
        returnData['id'] = str(id_num)
        returnData['name'] = str(name)
        returnData['school'] = str(school)
        returnData['score'] = str(score)
        returnData['listening'] = str(listening)
        returnData['reading'] = str(reading)
        returnData['writing'] = str(writing)
        returnData['speakscore'] = str(rank)
        return returnData
