from spider.lib_handler import *
from lxml import html
import time


class Lib(object):
    def __init__(self):
        self.data = {
            'index': 'default',
            'matchesPerPage': '10',
            'displayPages': '15',
            'searchPage': '1',
            'query': '',
            'submit': 'Bookle 搜索',
            'minPublishYear': '2005',
            'maxPublishYear': ''
        }
        self.client = requests.Session()

    # 书本查询
    def search(self, query, page="1"):
        """
        根据字符串查询书本
        :param query:查询字符串
        :param page:翻页数
        :return:书本信息（包含书籍列表）
        """
        url = 'http://lib.gzhu.edu.cn:8080/bookle/'
        self.data['query'] = query
        self.data['searchPage'] = page

        # 监测网络连通性
        try:
            res = self.client.post(url, data=self.data, timeout=10)
            return get_books(res.text)
        except:
            return 0

    # 获取
    def holdings(self, book_id, source):
        """
        获取书本馆藏信息
        :param book_id: 图书馆书本id
        :param source: 藏书来源
        :return: 馆藏信息列表
        """
        url = 'http://lib.gzhu.edu.cn:8080/bookle/search2/detail/' + book_id + '?source=' + source

        try:
            res = self.client.get(url, timeout=10)
            return get_holdings(res.text)
        except:
            return 0

    # 图书馆进馆人次
    def get_visit(self):
        hour = int(time.strftime("%H"))
        if 0 <= (hour + 8) % 24 < 6:  # UCT时区
            visit = {"master": 0, "branch": 0}
            return visit

        url = "http://lib.gzhu.edu.cn/w/"
        res = self.client.get(url, timeout=10)

        html_text = html.fromstring(res.text)
        view1 = html_text.xpath('//*[@id="form1"]/div[4]/div/p[1]/span[4]')[0].text
        view2 = html_text.xpath('//*[@id="form1"]/div[4]/div/p[1]/span[5]')[0].text

        master = re.findall("\d+", view1)[0]  # 总馆
        branch = re.findall("\d+", view2)[0]  # 分馆

        visit = {"master": master, "branch": branch}

        return visit

