var utils = require("../../utils/date.js")
var Config = require("../../utils/config.js")
Component({
  options: {
    addGlobalClass: true
  },
  /**
   * 组件的属性列表
   */
  properties: {

  },

  /**
   * 组件的初始数据
   */
  data: {
    dataSet: [],
    brick_option: {
      backgroundColor: "white",
      forceRepaint: false,
      imageFillMode: 'aspectFill',
      columns: 1,
    },
    open: Config.get("hot_topic") === false ? false : true
  },

  methods: {

    fake() {
      let mode = wx.$param["mode"]
      this.setData({
        mode: mode
      })
      if (mode == "prod") {
        return false
      } else return true
    },

    radioChange(e) {
      Config.set("hot_topic", e.detail.value)
      this.setData({
        open: e.detail.value
      })
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
    tapUser: function (e) {},
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

    // 获取列表
    getTopics(loadMore = false) {

      let query = {
        _page: 1,
        _page_size: 5,
        type: "$in.日常,情墙,悄悄话",
        _order: "-created_at",
      }
      query = wx.$objectToQuery(query)

      let url = wx.$param.server["prest"] + wx.$param.server["scheme"] + "/v_topic" + query


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
  },

  lifetimes: {
    attached: function () {
      this.fake()
      let that = this
      wx.getNetworkType({
        success(res) {
          if (res.networkType == "none" || !wx.getStorageSync("gzhupi_cookie")) return
          that.getTopics()
        }
      })
    }
  },
})