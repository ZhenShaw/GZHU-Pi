import requests
import re
import math
import json
from jsonpath_rw import parse


# 处理网页文本，提取书本列表数据
def get_books(text):
    books_info = {}

    # 搜索结果数
    try:
        total = int(re.findall(r'(\d+)条结果', text)[0])
        books_info['pages'] = int(math.ceil(total / 10))
        books_info['total'] = total
    except:
        books_info['total'] = 0
        books_info['pages'] = 0

    my_list = re.findall(
        r'noreferrer">\r\n\r\n +(.*?)\r*\n*</a>\r\n +<span>(.*?)</span>\r\n +</h2>\r\n +<h4>(.*?)\r\n +&nbsp;&nbsp;(.*?)\r\n',
        text)

    # 书本列表
    books_info['books'] = []
    for each in range(len(my_list)):
        books_info['books'].append({})
        books_info['books'][each]['book_name'] = my_list[each][0].replace('&middot;', '·')
        books_info['books'][each]['author'] = my_list[each][1].replace('&middot;', '·')
        # 获取出版社
        if my_list[each][2] != '':
            books_info['books'][each]['publisher'] = re.findall(r'：(.+)', my_list[each][2])[0].replace(',', '').replace(
                '&middot;', '·')
        else:
            books_info['books'][each]['publisher'] = my_list[each][2]

        # 获取ISBN和封面图片
        try:
            books_info['books'][each]['ISBN'] = ISBNFilter(my_list[each][3])
        except:
            books_info['books'][each]['ISBN'] = ''
        try:
            books_info['books'][each]['image'] = get_image(books_info['books'][each]['ISBN'])
        except:
            books_info['books'][each]['image'] = ''

    my_list = re.findall(r'</h4>\r\n +<h4>\r\n +(.*?)\r\n +&nbsp;&nbsp;(.*?)\r\n +&nbsp;&nbsp;(.*?)\r\n', text)
    for each in range(len(my_list)):
        books_info['books'][each]['call_No'] = re.findall(r'索书号：(.*)', my_list[each][0])[0]
        books_info['books'][each]['copies'] = re.findall(r'复本数：(.*)', my_list[each][1])[0]
        books_info['books'][each]['loanable'] = re.findall(r'在馆数：(.*)', my_list[each][2])[0]

    my_list = re.findall(r'/bookle/search2/detail/(.+)\?index=defaul.*source=(.*)"', text)
    for each in range(len(my_list)):
        books_info['books'][each]['id'] = my_list[each][0]
        books_info['books'][each]['source'] = my_list[each][1]

    return books_info


# 处理网页文本，提取馆藏数据
def get_holdings(text):
    infolist = re.findall(r' <td>(.*?)</td>\r\n', text)

    holdings = []
    # 添加除地点外信息（地点信息第一次未能提取到）
    for each in range(int(len(infolist) / 6)):
        holdings.append({})

        holdings[each]['bar_code'] = infolist[0 + each * 6]
        holdings[each]['status'] = infolist[1 + each * 6]
        holdings[each]['loan_date'] = infolist[2 + each * 6]
        holdings[each]['due_back'] = infolist[3 + each * 6]
        holdings[each]['circulate'] = infolist[4 + each * 6]
        holdings[each]['explain'] = infolist[5 + each * 6]

    infolist = re.findall(r' (.*?)\r\n +\r\n +<a target', text)
    for each in range(len(infolist)):
        holdings[each]['location'] = infolist[each].replace(' ', '')

    return holdings


# 精提取ISBN
def ISBNFilter(badISBN):
    ISBN = badISBN.replace('标准号', '').replace('、', '').replace('价格', '')
    ISBN = ISBN.replace(' ', '').replace(':', '').replace('：', '').replace('(pbk.)', '')
    ISBN = ISBN.replace('-', '')
    try:
        ISBN = ISBN[0:ISBN.index('CNY')]
    except:
        pass
    try:
        ISBN = ISBN[0:ISBN.index('.')]
    except:
        pass
    return ISBN


# 根据ISBN从豆瓣获取封面图片
def get_image(ISBN):
    if ISBN == "":
        return ""
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 "
                      "(KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"
    }
    res = requests.get(url='https://douban.uieee.com/v2/book/isbn/' + ISBN, headers=headers)
    res_json = json.loads(res.text)
    if not parse('$.code').find(res_json):
        image = parse('$.image').find(res_json)[0].value
    else:
        image = ""
    return image
