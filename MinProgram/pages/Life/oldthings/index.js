const Page = require('../../../utils/sdk/ald-stat.js').Page;
Page({

  data: {
    limit: 10, //每页数量
    offset: 0, //偏移量
    loadDone: false, //加载完毕
    queryStr: "", //搜索的字符串
    loading: false,
    category: ["全部", "图书文具", "生活用品", "电子产品", "化妆用品", "服装鞋包", "其它"],
    categoryIndex: 0,

    dataSet: [],
    brick_option: {
      showFullContent: true,
      backgroundColor: "rgb(235, 246, 250)",
      forceRepaint: true,
      defaultExpandStatus: false,
      imageFillMode: 'aspectFill',
      columns: 1,
      icon: {
        fill: 'https://images.ifanr.cn/wp-content/uploads/2018/08/640-90-1024x576.jpeg',
        default: 'https://images.ifanr.cn/wp-content/uploads/2018/08/640-90-1024x576.jpeg',
      },
      fontColor: 'black'
    },

    gridCol: 4,
    iconList: [{
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/house.svg',
      name: '全部'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/books.svg',
      name: '图书文具'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/yashua.svg',
      name: '生活用品'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/dianzi.svg',
      color: 'olive',
      name: '电子产品'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/kouhong.svg',
      name: '化妆用品'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/clothes.svg',
      color: 'blue',
      name: '服装鞋包'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/chanping.svg',
      name: '其它'
    }, {
      icon: 'https://cos.ifeel.vip/gzhu-pi/images/icon/wode.svg',
      name: '我的发布'
    }],

  },

  fake() {
    let mode = wx.$param["mode"]
    this.setData({
      mode: mode
    })
    if (mode == "prod") {
      return false
    } else return true
  },

  onLoad: function (options) {

    if (this.fake()) return

    this.getGoods()
  },

  onShareAppMessage: function () {

  },

  // 下拉刷新
  onPullDownRefresh: function () {
    this.setData({
      offset: 0, //恢复偏移量
      loadDone: false, //加载完毕
      queryStr: ""
    })
    this.getGoods()
    setTimeout(function () {
      wx.stopPullDownRefresh()
    }, 3000)
  },

  // 点击卡片，获取商品id，转跳详情页面
  tapCard: function (event) {
    console.log("商品ID：", event.detail.card_id)
    wx.navigateTo({
      url: '/pages/Life/oldthings/detail?id=' + event.detail.card_id,
    })
  },
  // 点击头像
  tapUser: function (e) {
    console.log("用户id:", e.detail.user_id)
    wx.navigateTo({
      url: '/pages/Life/oldthings/mine?id=' + e.detail.user_id,
    })
  },

  navToPost() {
    wx.navigateTo({
      url: '/pages/Life/oldthings/post',
    })
  },

  // 读取搜索内容
  searchInput: function (e) {
    this.data.queryStr = e.detail.value
  },

  search() {
    if (this.data.queryStr=="") return
    this.setData({
      categoryIndex: -1
    })
    this.getGoods()
  },

  // 触底加载更多，需改变offset，判断有无更多
  onReachBottom: function () {
    if (this.data.loadDone) return
    console.log('加载更多')
    this.data.offset = this.data.offset + this.data.limit
    this.getGoods(true)
  },

  // 切换分类
  switchCategory(e) {
    let id = Number(e.currentTarget.id)

    if (this.data.iconList[id].name == "我的发布") {
      wx.navigateTo({
        url: '/pages/Life/oldthings/mine'
      })
      return
    }
    if (id == this.data.categoryIndex) return

    this.setData({
      offset: 0, //恢复偏移量
      loadDone: false, //加载完毕
      queryStr: "",
      categoryIndex: id
    })
    this.getGoods()
  },

  // 获取商品
  getGoods(loadMore = false) {
    let that = this
    this.setData({
      loading: true
    })
    let table = new wx.BaaS.TableObject('flea_market')


    let query = new wx.BaaS.Query()
    query.compare('status', '=', 0) //筛选条件，0表示正常
    // 类别筛选
    let id = this.data.categoryIndex
    if (id != 0 && id < 7) {
      query.compare('category', '=', this.data.category[id]) //筛选类别
    }

    // 关键词模糊搜索
    let queryStr = this.data.queryStr
    if (queryStr != "") {

      let query1 = new wx.BaaS.Query()
      query1.contains('title', queryStr) // 标题

      let query2 = new wx.BaaS.Query()
      query2.contains('content', queryStr) // 描述

      let query3 = new wx.BaaS.Query()
      query3.in('label', [queryStr]) // 标签数组包含

      query = wx.BaaS.Query.or(query1, query2, query3)
    }


    let limit = this.data.limit
    table.setQuery(query)
      .limit(limit).offset(this.data.offset) //一次获取20条数据
      .orderBy('-refresh_time') //擦亮时间倒序
      .expand(['created_by']) //拓展用户信息
      .find().then(res => {
        console.log("商品列表", res.data.objects)

        if (loadMore) {
          // 触底加载更多，把获取到的数据拼接到原数据后面
          this.setData({
            dataSet: this.data.dataSet.concat(res.data.objects),
            loading: false,
            loadDone: res.data.objects.length < limit ? true : false //加载完毕
          })
        } else {
          // 切换类别，直接渲染
          this.setData({
            dataSet: res.data.objects,
            loading: false,
            loadDone: res.data.objects.length < limit ? true : false //加载完毕
          })
        }
        wx.stopPullDownRefresh()
      }, err => {
        wx.showToast({
          title: '获取商品错误',
          icon: "none"
        })
        that.setData({
          loading: false
        })
      })
  },


})