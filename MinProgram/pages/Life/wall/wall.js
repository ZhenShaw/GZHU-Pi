const Page = require('../../../utils/sdk/ald-stat.js').Page;
Page({

  data: {
    navTitle: "广大墙",
    limit: 10, //每页数量
    offset: 0, //偏移量
    loadDone: false, //加载完毕
    queryStr: "", //搜索的字符串
    loading: false,
    category: ["全部", "图书文具", "生活用品", "电子产品", "化妆用品", "服装鞋包", "其它"],
    categoryIndex: 0,

    dataSet: [],
    brick_option: {
      backgroundColor: "rgb(245, 245, 245)",
      forceRepaint: true,
      imageFillMode: 'aspectFill',
      columns: 2,
    },

    gridCol: 2,
    iconList: [{
      icon: 'https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/tmp/WechatIMG200.png',
      name: '广大水墙'
    }, {
      icon: 'https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/tmp/WechatIMG1971.png',
      name: '广大情墙'
    }, {
      icon: 'https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/tmp/WechatIMG201.png',
      name: '悄悄话'
    }, {
      icon: 'https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/tmp/WechatIMG202.png',
      name: '校园市场'
    }],

  },


  onLoad: function(options) {

    if (wx.$param["mode"] != "prod") {
      this.setData({
        normal: false
      })
      return
    } else {
      this.setData({
        normal: true
      })
    }

    this.getTopics()
  },

  onShareAppMessage: function() {

  },

  // 下拉刷新
  onPullDownRefresh: function() {
    this.setData({
      offset: 0, //恢复偏移量
      loadDone: false, //加载完毕
      queryStr: ""
    })
    this.getTopics()
    setTimeout(function() {
      wx.stopPullDownRefresh()
    }, 3000)
  },

  // 点击卡片，获取商品id，转跳详情页面
  tapCard: function(event) {
    console.log("商品ID：", event.detail.card_id)
    wx.navigateTo({
      url: '/pages/Life/wall/detail?id=' + event.detail.card_id,
    })
  },
  // 点击头像
  tapUser: function(e) {
    console.log("用户id:", e.detail.user_id)
    wx.navigateTo({
      url: '/pages/Life/oldthings/mine?id=' + e.detail.user_id,
    })
  },

  navToPost() {
    wx.navigateTo({
      url: '/pages/Life/wall/post',
    })
  },

  // 读取搜索内容
  searchInput: function(e) {
    this.data.queryStr = e.detail.value
  },

  search() {
    this.setData({
      categoryIndex: -1
    })
    this.getTopics()
  },

  // 触底加载更多，需改变offset，判断有无更多
  onReachBottom: function() {
    if (this.data.loadDone) return
    console.log('加载更多')
    this.data.offset = this.data.offset + this.data.limit
    // this.getTopics(true)
  },

  // 切换分类
  switchCategory(e) {
    let id = Number(e.currentTarget.id)

    switch (this.data.iconList[id].name) {
      case "广大水墙":
        this.setData({
          "brick_option.columns": 2,
          navTitle:"广大墙"
        })
        break
      case "广大情墙":
        break
      case "悄悄话":
        this.setData({
          "brick_option.columns": 1,
          navTitle: "悄悄话"
        })
        break
      case "校园市场":
        wx.$navTo("/pages/Life/oldthings/index")
        break
      default:
        console.error("unknown type")
        return
    }

    this.setData({
      offset: 0, //恢复偏移量
      loadDone: false, //加载完毕
      queryStr: "",
      categoryIndex: id
    })
    // this.getTopics()
  },

  // 获取商品
  getTopics(loadMore = false) {
    // return
    wx.$ajax({
        url: wx.$param.server["prest"] + "/postgres/public/v_topic",
        method: "get",
        loading: true,
        checkStatus: false,
      })
      .then(res => {
        console.log(res)
        this.setData({
          dataSet: res.data,
          loading: false,
        })
      }).catch(err => {
        console.log(err)
      })


  },


})