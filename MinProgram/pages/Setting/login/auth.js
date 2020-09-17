const Page = require('../../../utils/sdk/ald-stat.js').Page;
Page({

  /**
   * 页面的初始数据
   */
  data: {

    v_user: {},
    bind_success: false,

    qrcode: "https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/resource/qrcode.jpg"
  },

  onLoad: function (options) {

    if (!wx.$checkUser(false)) {
      wx.showToast({
        title: '请先授权微信',
        icon: "none"
      })
    }
    this.auth()

    let mp_open_id = options.mp_open_id

    // 绑定公众号
    if (!!mp_open_id) {
      this.bind_mp(mp_open_id)
      return
    }

    this.setData({
      v_user: wx.getStorageSync('gzhupi_user')
    })

  },
  onShareAppMessage: function () {

  },

  viewImg(e) {
    wx.$viewImg([], e)
  },

  // 关联公众号请求
  bind_mp(mp_open_id) {
    console.log("绑定公众号：", mp_open_id)
    wx.$ajax({
        url: wx.$param.server["prest"] + "/auth?type=bind_mp&mp_open_id=" + mp_open_id,
        // url: "http://localhost:9000/api/v1/auth?type=bind_mp&mp_open_id=" + mp_open_id,
        method: "post",
        data: {},
      })
      .then(res => {
        console.log(res)

        if (res.data.id > 0) {
          this.setData({
            v_user: res.data,
            bind_success: true
          })
          wx.setStorage({
            key: 'gzhupi_user',
            data: res.data,
          })
        }
      })
  },

  userInfoHandler(data) {
    let that = this
    wx.showLoading({
      title: '授权中...',
    })
    wx.BaaS.auth.loginWithWechat(data, {
      createUser: true,
      syncUserProfile: "overwrite"
    }).then(user => {
      console.log("minapp", user)
      wx.hideLoading()

      this.auth(user)

    }).catch(err => {
      console.log("拒绝授权", err)
      wx.hideLoading()
      wx.showToast({
        title: '授权失败，可退出重试',
        icon: "none"
      })
    })

  },

  async auth(user) {

    if (!user || user == undefined || !user.user_id) {
      user = await wx.BaaS.auth.getCurrentUser().then(user => {
        return user
      }).catch(err => {
        console.error(err)
      })
    }

    if (user == undefined || !user.user_id) {
      return
    }

    let form = {
      minapp_id: user.id,
      open_id: user.openid,
      union_id: user.unionid,
      avatar: user.avatar,
      nickname: user.nickname,
      city: user.city,
      province: user.province,
      country: user.country,
      gender: user.gender,
      language: user.language,
      phone: user._phone,
    }
    wx.$ajax({
        url: wx.$param.server["prest"] + "/auth",
        // url: "http://localhost:9000/api/v1/auth",
        method: "post",
        data: form,
        header: {
          "content-type": "application/json"
        }
      })
      .then(res => {
        console.log("auth", res)
        if (res.data && res.data.id > 0 && res.data.open_id == user.openid) {
          wx.setStorage({
            key: 'gzhupi_user',
            data: res.data,
          })
          this.setData({
            v_user: res.data
          })
        }
      })
  },


  navTo(e) {
    wx.$navTo(e)
  },

  onClick() {
    wx.showModal({
      title: "提示",
      content: "在广大派公众号发送【绑定】"
    })
  }

})