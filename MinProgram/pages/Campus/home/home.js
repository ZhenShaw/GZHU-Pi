const Page = require('../../../utils/sdk/ald-stat.js').Page;
var app = getApp()
var Config = require("../../../utils/config.js")
var Setting = require("../../../utils/setting.js")
Page({
  data: {
    schedule: Config.get("schedule_mode") == "week" ? true : false,
    navColor: Config.get("schedule_mode") == "week" ? "rgba(221, 221, 221, 0.7)" : "rgba(255, 255, 255, 0.8)",
    out: "ami",
    showUpdate: false,
    showDrawer: false,
    arrowUrl: "https://cos.ifeel.vip/gzhu-pi/images/icon/right-arrow.svg",
  },

  onLoad: function (options) {},

  onShareAppMessage: function () {},

  onShareTimeline: function () {
    return {
      title: "广大派-广大人必备的掌中宝"
    }
  },

  onReady() {
    wx.hideLoading()
    this.updateCheck()
    this.setData({
      mode: wx.$param.mode
    })
  },

  updateCheck() {
    let version = Config.get("version")
    if (version < Config.config['version']) {
      this.setData({
        showUpdate: true
      })
      Config.set("version", Config.config['version'])
    }
  },

  // 切换课表模式，点解悬浮图标
  switchModel() {
    if (this.data.schedule) {
      this.setData({
        schedule: !this.data.schedule,
        navColor: "",
      })
    } else {
      this.setData({
        schedule: !this.data.schedule,
        navColor: "rgba(221, 221, 221, 0.7)",
      })
    }

  },

  // 打开抽屉弹窗
  openDrawer() {
    this.setData({
      showDrawer: true,
      userInfo: wx.getStorageSync("ifx_baas_userinfo")
    })
  },

  // 抽屉选项
  tapDrawer(e) {
    let drawerItem = e.currentTarget.id
    const schedule = this.selectComponent('#schedule')
    switch (drawerItem) {
      case "changeBg":
      case "changeMode":
        this.setData({
          drawerItem: drawerItem == this.data.drawerItem ? null : drawerItem,
          checkedBlur: Config.get("blur"),
          mode: Config.get("schedule_mode")
        })
        break
      case "selectImg":
        Setting.setBg().then(res => {
          Config.set("schedule_bg", res)
          schedule.updateBg()
        })
        break
      case "white,white":
      case "#ddd,#ddd":
      case "#d299c2,#fef9d7":
      case "#a8edea,#fed6e3":
        Config.set("schedule_bg", drawerItem)
        schedule.updateBg()
        break
      case "navToAbout":
        wx.navigateTo({
          url: '/pages/Setting/about/data',
        })
        break
      case "navToSync":
        wx.navigateTo({
          url: "/pages/Setting/login/sync",
        })
        break
      case "navToHelp":
        wx.navigateTo({
          url: "/pages/Setting/help/help",
        })
        break
      case "navToBind":
        wx.navigateTo({
          url: "/pages/Setting/login/auth",
        })
        break
    }
  },

  // 开启关闭高斯模糊
  switchChange(e) {
    if (e.detail.value) Config.set("blur", 8)
    else Config.set("blur", 0)
    const schedule = this.selectComponent('#schedule')
    schedule.updateBg()
  },

  // 切换课表模式
  radioChange(e) {
    Config.set("schedule_mode", e.detail.value)

    if (e.detail.value == "day") this.data.schedule = true
    else this.data.schedule = false
    this.switchModel()
  },
  catchtap(e) {},

  navTo(e) {
    wx.$navTo(e)
    // wx.switchTab({
    //   url: '/pages/Life/wall/wall',
    // })
  }

})