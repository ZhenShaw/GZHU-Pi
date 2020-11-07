const Page = require('../../../utils/sdk/ald-stat.js').Page;
var Request = require("../../../utils/request.js")
Page({

  data: {
    current: 1,
    sem_list: wx.$param.school["sem_list"],
    picker_index: wx.$param.school["sem_list"].indexOf(wx.$param.school["year_sem"]), //当前学期索引
    loading: false,
    exp_btn: "同步实验",
    kb_btn: "同步课表",

    exam_picker_index: 0,
    exam_list: ["大学英语四级-CET4", "大学英语六级-CET6"]
  },

  onLoad: function (options) {

    this.setData({
      update_time: wx.getStorageSync("course").update_time,
      cet: wx.getStorageSync("cet"),
      account: wx.getStorageSync("account"),
      exp_account: wx.getStorageSync("exp_account"),
    })
    // 根据来源的不同显示不同界面
    if (JSON.stringify(options) != "{}") {
      this.setData({
        current: Number(options.id)
      })
    }
  },

  navTo(e) {
    wx.$navTo(e)
  },

  switch (e) {
    this.setData({
      current: Number(e.target.id)
    })
  },
  swiperChange(e) {
    this.setData({
      current: e.detail.current
    })
  },
  pickerChange(e) {

    switch (e.target.id) {
      case "sem":
        this.setData({
          picker_index: Number(e.detail.value)
        })
        break
      case "exam":
        this.setData({
          exam_picker_index: Number(e.detail.value)
        })
    }

  },

  // 提交表单
  formSubmit(e) {
    let that = this
    let id = e.detail.target.id

    if (e.detail.value.username == "" || e.detail.value.password == "") {
      wx.showToast({
        title: '用户名、密码不能为空',
        icon: "none"
      })
      return
    }
    if (e.detail.value.name == "" || e.detail.value.idc == "") {
      wx.showToast({
        title: "姓名、身份证不能为空",
        icon: "none"
      })
      return
    }
    // 防止多次点击
    if (this.data.loading) {
      return
    }
    this.setData({
      loading: true
    })

    let account = {
      username: e.detail.value.username,
      password: e.detail.value.password,
      year_sem: that.data.sem_list[that.data.picker_index],
      first_monday: wx.$param.school["first_monday"]
    }

    switch (id) {
      // 课表
      case "kb":
        wx.showModal({
          title: '提示',
          content: '同步将会覆盖当前课表',
          success: function (res) {
            if (res.confirm) {
              wx.$ajax({
                  url: "/jwxt/course",
                  data: account,
                  loading: true
                })
                .then(res => {
                  wx.showToast({
                    title: "同步完成",
                    icon: "none"
                  })
                  that.setData({
                    loading: false,
                    update_time: res.update_time,
                    kb_btn: "已同步",
                  })
                  // 缓存账户信息
                  delete account["year_sem"]
                  delete account["first_monday"]
                  wx.setStorageSync("account", account)
                  // 缓存结果数据
                  res.data["update_time"] = res.update_time
                  wx.setStorageSync("course", res.data)
                }).catch(err => {
                  that.setData({
                    loading: false
                  })
                })
            } else {
              that.setData({
                loading: false
              })
            }
          }
        })
        break


        // 实验
      case "exp":
        wx.$ajax({
            url: "/exp",
            data: account,
            loading: true
          })
          .then(res => {
            wx.showToast({
              title: "同步完成",
              icon: "none"
            })
            wx.setStorageSync("exp", res.data)
            wx.setStorageSync("exp_account", account)

            that.setData({
              loading: false
            })

          }).catch(err => {
            console.error(err)
            that.setData({
              loading: false
            })
          })
        break


        // 四六级
      case "toCet":
        wx.setStorageSync("cet", e.detail.value)
        let cet = that.data.exam_list[that.data.exam_picker_index]
        cet = cet.indexOf("4") >= 0 ? 4 : 6
        wx.navigateTo({
          url: '/pages/Campus/cet/cet?cet=' + cet,
        })
        this.setData({
          loading: false
        })
        break
    }
  },
})