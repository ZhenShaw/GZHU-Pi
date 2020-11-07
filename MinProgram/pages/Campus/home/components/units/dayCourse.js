var utils = require("../../../../../utils/utils.js")
var showTimes = 0
Component({

  properties: {

  },

  data: {
    todayCourse: utils.getTodayCourse(),
    week: utils.getSchoolWeek(), //周数
    schoolWeek: utils.getSchoolWeek(), //校历周
    weekDays: ['日', '一', '二', '三', '四', '五', '六', ],
    weekday: new Date().getDay()
  },

  methods: {
    nav() {
      wx.navigateTo({
        url: '/pages/Setting/login/bindStudent',
      })
    },
  },

  lifetimes: {
    attached: function () {
      let account = wx.getStorageSync('account')
      this.setData({
        haveAccount: !(account == "" || !account)
      })
    },
  },
  pageLifetimes: {
    show() {
      this.setData({
        week: utils.getSchoolWeek(), //周数
        schoolWeek: utils.getSchoolWeek(), //校历周
      })
      // 初次onshow不执行
      if (showTimes) {
        this.setData({
          todayCourse: utils.getTodayCourse()
        })
      }
      showTimes++
    }
  }
})