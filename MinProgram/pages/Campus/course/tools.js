const Page = require('../../../utils/sdk/ald-stat.js').Page;
// let rewardedVideoAd = null
let interstitialAd = null

Page({

  data: {
    id: "open",
    title: "任意门",

    semList: wx.$param.school["sem_list"],
    semIndex: wx.$param.school["sem_list"].indexOf(wx.$param.school["year_sem"]), //默认显示的学期索引

  },

  onLoad: function (options) {
    console.log(options)
    let title = {
      eval: "课程评价",
      query: "任意门",
      favorite: "课表收藏"
    }
    this.setData({
      id: options.id,
      title: title[options.id]
    })

    this.insertAD()

    this.search_class("计算机")
  },

  onShareAppMessage: function () {
    return {
      title: this.data.title ? this.data.title : "",
      path: '/pages/Campus/course/tools?id=query',
      imageUrl: imageUrl ? imageUrl : "",
      success: function (res) {
        wx.showToast({
          title: '分享成功',
          icon: "none"
        });
      }
    }
  },

  insertAD() {
    if (wx.createInterstitialAd) {
      interstitialAd = wx.createInterstitialAd({
        adUnitId: 'adunit-fba645195084f056'
      })
      interstitialAd.onLoad(() => {
        console.log('onLoad event emit')
      })
      interstitialAd.onError((err) => {
        console.log('onError event emit', err)
      })
      interstitialAd.onClose((res) => {
        console.log('onClose event emit', res)
      })
    }
  },
  onShow() {
    if (interstitialAd) {
      interstitialAd.show().catch((err) => {
        console.error('show ad error ', err)
      })
    }
  },

  nav() {

    if (!this.data.target) return

    let str = this.data.semList[this.data.semIndex]
    let sp = str.split("-")
    let year = sp[0] + "-" + sp[1]
    let sem = sp[2]

    wx.navigateTo({
      url: '/pages/Campus/course/schedule?classNmae=' + this.data.target + "&year=" + year + "&sem=" + sem,
    })
  },

  actionSheet(e) {
    let that = this
    wx.showActionSheet({
      itemList: this.data.semList,
      success(res) {
        that.setData({
          semIndex: res.tapIndex
        })
      }
    })

  },

  search() {
    this.search_class(this.keyword)
  },

  search_class(keyword) {
    if (!keyword) {
      this.setData({
        class_list: []
      })
      return
    }

    let Obj = new wx.BaaS.TableObject("college_major_class")

    // 班级
    let query1 = new wx.BaaS.Query()
    query1.contains('bj', keyword)
    // 专业
    let query2 = new wx.BaaS.Query()
    query2.contains('zymc', keyword)
    // 学院
    let query3 = new wx.BaaS.Query()
    query3.contains('jgmc', keyword)

    let orQuery = wx.BaaS.Query.or(query1, query2, query3)

    Obj.setQuery(orQuery).orderBy('-njmc').limit(10).find().then(res => {
      let list = res.data.objects
      this.setData({
        class_list: list
      })
      if (list.length > 0) {
        this.setData({
          target: list[0].bj
        })
      }
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

  searchInput(e) {
    this.keyword = e.detail.value
    if (!this.isDebounce(300)) {
      this.search_class(this.keyword)
    }
  },
  selectClass(e) {
    this.setData({
      target: e.currentTarget.dataset.target
    })
  }

})