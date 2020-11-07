const Page = require('../../../utils/sdk/ald-stat.js').Page
var utils = require("../../../utils/date.js")
Page({

  data: {
    navTitle: "广大墙",
    pageSize: 20, //每页数量
    page: 1, //页数
    loadDone: false, //加载完毕
    queryStr: "", //搜索的字符串
    loading: true,

    dataSet: [],
    brick_option: {
      backgroundColor: "white",
      forceRepaint: false,
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
    let that = this
    setTimeout(() => {
      that.setData({
        loading: false
      })
    }, 1000);

    if (this.fake()) return

    this.getTopics()
  },

  onShareAppMessage: function () {},

  onShareTimeline: function () {
    return {
      title: "广大墙-属于广大人的专属票圈"
    }
  },

  // 下拉刷新
  onPullDownRefresh: function () {
    this.setData({
      page: 1, //恢复页数
      loadDone: false, //加载完毕
      queryStr: ""
    })
    this.getTopics()
    setTimeout(function () {
      wx.stopPullDownRefresh()
    }, 3000)
  },

  // true 说明在防抖期间，应该停止执行
  isDebounce(timeout = 2000) {
    let that = this
    if (this.data.debounce) {
      console.log("触发防抖")
      return true
    }
    this.data.debounce = true
    setTimeout(() => {
      that.data.debounce = false
    }, timeout)
    return false
  },

  // 点击卡片，获取id，转跳详情页面
  tapCard: function (event) {
    console.log("ID：", event.detail.card_id)
    let args = {
      id: event.detail.card_id
    }
    wx.$navTo('/pages/Life/wall/detail', args)
  },
  // 点击头像
  tapUser: function (e) {
    // console.log("用户id:", e.detail.user_id)
    // let args = {
    //   id: e.detail.user_id,
    // }
    // wx.$navTo('/pages/Life/oldthings/mine', args)
  },
  // 点击喜欢
  tapLike: function (e) {
    if (this.isDebounce(1500)) return

    console.log("点赞:", e.detail.card_id)
    let cur_uid = wx.getStorageSync('gzhupi_user').id
    let topic_id = e.detail.card_id

    let topic_index = -1
    let star_list = []
    for (let i in this.data.dataSet) {
      if (this.data.dataSet[i].id == topic_id) {
        topic_index = i
        star_list = this.data.dataSet[i].star_list
        for (let j in star_list) {
          if (star_list[j].created_by == cur_uid && star_list[j].type == "star") {
            console.log("用户已点赞，取消")
            wx.$ajax({
                url: wx.$param.server["prest"] + wx.$param.server["scheme"] + "/t_relation?id=$eq." + star_list[j].id,
                method: "delete",
              })
              .then(res => {
                star_list.splice(j, 1)
                this.setData({
                  ["dataSet[" + i + "].star_list"]: star_list
                })
              }).catch(err => {
                console.error(err)
              })
            return
          }
        }
      }
    }
    if (topic_index < 0) return
    console.log("用户未点赞，点赞")
    wx.$ajax({
      url: wx.$param.server["prest"] + wx.$param.server["scheme"] + "/t_relation",
      method: "post",
      data: {
        object_id: Number(topic_id),
        object: "t_topic",
        type: "star"
      },
      header: {
        "content-type": "application/json"
      }
    }).then(res => {
      if (typeof star_list != "object" || star_list == null) star_list = []
      if (res.data && res.data.id) {
        res.data.avatar = wx.getStorageSync('gzhupi_user').avatar
        star_list.push(res.data)
        this.setData({
          ["dataSet[" + topic_index + "].star_list"]: star_list
        })
      }
    })
  },

  navToPost() {
    wx.$subscribe()
    wx.$navTo('/pages/Life/wall/post')
  },

  // 读取搜索内容
  searchInput: function (e) {
    this.data.queryStr = e.detail.value
  },

  search() {
    if (this.data.queryStr == "") {
      this.setData({
        page: 1, //恢复页数
        loadDone: false, //加载完毕
        type: "",
        "brick_option.columns": 2,
        // loading: true
      })

    } else {
      this.setData({
        page: 1, //恢复页数
        loadDone: false, //加载完毕
        dataSet: [],
        // loading: true
      })
    }
    this.getTopics()
  },

  // 触底加载更多，需改变offset，判断有无更多
  onReachBottom: function () {
    if (this.data.loadDone) return
    console.log('加载更多')
    this.data.page = this.data.page + 1
    this.getTopics(true)
  },

  // 切换分类
  switchCategory(e) {
    let id = Number(e.currentTarget.id)

    switch (this.data.iconList[id].name) {
      case "广大水墙":
        this.setData({
          "brick_option.columns": 2,
          navTitle: "广大墙"
        })
        this.data.type = "$eq.日常"
        break
      case "广大情墙":
        this.data.type = "$eq.情墙"
        this.setData({
          "brick_option.columns": 1,
          navTitle: "广大情墙"
        })
        break
      case "悄悄话":
        this.setData({
          "brick_option.columns": 1,
          navTitle: "悄悄话"
        })
        this.data.type = "$eq.悄悄话"
        break
      case "校园市场":
        wx.$navTo("/pages/Life/oldthings/index")
        return
      default:
        console.error("unknown type")
        return
    }

    this.setData({
      page: 1, //恢复页数
      loadDone: false, //加载完毕
      queryStr: "",
      dataSet: [],
      loading: true
    })

    this.getTopics()

  },

  // 获取列表
  getTopics(loadMore = false) {

    let query = {
      _page: this.data.page,
      _page_size: this.data.pageSize,
      type: this.data.type ? this.data.type : "$in.日常,情墙,悄悄话",
      _order: "-order_time",
    }
    query = wx.$objectToQuery(query)

    let url = wx.$param.server["prest"] + wx.$param.server["scheme"] + "/v_topic" + query

    if (this.data.queryStr != "") {
      url = wx.$param.server["prest"] + "/_QUERIES/topic/v_topic_search?match=" + this.data.queryStr
    }

    wx.$ajax({
        url: url,
        method: "get",
      })
      .then(res => {
        console.log("主题列表", res)
        // 格式化时间
        for (let i = 0; i < res.data.length; i++) {
          let time = new Date(res.data[i].updated_at)
          res.data[i].updated_at = utils.relativeTime(time.getTime() / 1000)
        }
        if (loadMore) {
          this.data.dataSet = this.data.dataSet.concat(res.data)
        } else {
          this.data.dataSet = res.data
        }
        this.setData({
          dataSet: this.data.dataSet,
          loading: false,
          loadDone: res.data.length < this.data.pageSize ? true : false //加载完毕
        })

      }).catch(err => {
        console.log(err)
        this.setData({
          loading: false
        })
      })
  },

})