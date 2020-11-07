// pages/Campus/cet/cet.js
Page({
  data: {
    name: "Waua",
    scoreList: [],
    // 请求信息
    cetInfo: null,
    // 显示验证
    isCaptcha: false,
    captchaimg: null,
    captcha: "",
    // 请求4/6级
    subject: 4,
    subjectName: "",
    curCard: 0,
  },
  onLoad: function (options) {
    if (JSON.stringify(options) != "{}") {
      this.setData({
        subject: Number(options.cet),
      });
    }
  },
  onShow: function () {
    this.setData({
      cetInfo: wx.getStorageSync("cet"),
    });
    this.checkStorge();
  },
  onShareAppMessage: function () { },
  getCaptcha() {
    let that = this;
    this.setData({
      captcha: ""
    });
    if (this.data.cetInfo) {
      wx.$ajax({
        method: "GET",
        url: "/pastCetCaptcha",
        loading: true,
      })
        .then((res) => {
          that.setData({
            isCaptcha: true,
            captchaimg: res.data,
          });
        })
        .catch((err) => {
          console.log(err);
        });
    } else {
      // 未填写信息
      this.noCetInfo();
    }
  },
  getScore() {
    let that = this;
    let _name = String(this.data.cetInfo.name);
    let _idc = String(this.data.cetInfo.idc);
    let _captcha = String(this.data.captcha);
    let _subject = String(this.data.subject);

    wx.$ajax({
      method: "GET",
      url: `/pastCet?subject=CET${_subject}&xm=${_name}&sfz=${_idc}&captcha=${_captcha}`,
      loading: true,
    })
      .then((res) => {
        wx.setStorageSync(`cet${_subject}Score`, res.data);
        that.checkStorge();
        that.setData({
          isCaptcha: false,
        });
      })
      .catch((err) => {
        that.getCaptcha();
      });
  },
  checkStorge() {
    let _cetStore = wx.getStorageSync(`cet${String(this.data.subject)}Score`);
    let _cetInfoStore = wx.getStorageSync("cet");

    // 不存在缓存
    if (!_cetStore) {
      this.getCaptcha();
      return;
    }

    // 请求名称||身份证变更
    if ((_cetStore.name != _cetInfoStore.name) || (_cetStore.person_id != _cetInfoStore.idc)) {
      this.getCaptcha();
      return;
    }

    this.render(this.data.subject);
  },
  render(subject) {
    // score_list
    // 1.score
    // 2.listening
    // 3.reading
    // 4.writing
    let _data = wx.getStorageSync(`cet${subject}Score`);
    this.setData({
      subjectName: _data.subject,
      name: _data.name,
      scoreList: _data.score_list,
    });
  },
  noCetInfo() {
    wx.showModal({
      title: "错误",
      content: "未填信息",
      cancelText: "先不填",
      confirmText: "现在填",
      cancelColor: "cancelColor",
      success(res) {
        if (res.confirm) {
          wx.navigateTo({
            url: "/pages/Setting/login/sync?id=2",
          });
        } else {
          wx.switchTab({
            url: "/pages/Campus/home/home",
            fail: function (res) {
              console.log("fail: ", res);
            },
          });
        }
      },
    });
  },
  cardSwiper(e) {
    this.setData({
      curCard: e.detail.current,
    });
  },
  bindKeyInput(e) {
    this.setData({
      captcha: e.detail.value,
    });
  },
});
