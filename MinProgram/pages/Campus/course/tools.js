const Page = require('../../../utils/sdk/ald-stat.js').Page;
// let rewardedVideoAd = null
let interstitialAd = null

Page({

  data: {
    id: "open",
    title: "任意门",

    semList: wx.$param.school["sem_list"],
    semIndex: wx.$param.school["sem_list"].indexOf(wx.$param.school["year_sem"]), //默认显示的学期索引

    yearIndex: 2,
    collegeIndex: 5,
    majorIndex: 1,
    classIndex: 0,
    majors: [],
    years: ["2022","2021","2020","2019", "2018", "2017", "2016"],
    colleges: ["地理科学学院", "法学院", "公共管理学院", "环境科学与工程学院",
      "化学化工学院", "计算机科学与网络工程学院", "机械与电气工程学院", "教育学院", "建筑与城市规划学院",
      "旅游学院", "美术与设计学院", "人文学院", "生命科学学院", "数学与信息科学学院",
      "土木工程学院", "体育学院", "外国语学院", "物理与电子工程学院", "新闻与传播学院",
      "音乐舞蹈学院", "国际教育学院", "马克思主义学院", "经济与统计学院",
      "工商管理学院", "卫斯理安学院", "创新创业学院",
      "教师培训学院继续教育学院", "第二专业"
    ]
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

    this.getClassList()
    // this.rewardAD()
    this.insertAD()
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


  // rewardAD() {
  //   if (wx.createRewardedVideoAd) {
  //     rewardedVideoAd = wx.createRewardedVideoAd({
  //       adUnitId: 'adunit-bdbbdebf506ba881'
  //     })
  //     rewardedVideoAd.onLoad(() => {
  //       console.log('onLoad event emit')
  //     })
  //     rewardedVideoAd.onError((err) => {
  //       console.log('onError event emit', err)
  //     })
  //     rewardedVideoAd.onClose((res) => {
  //       console.log('onClose event emit', res)
  //       if (res && res.isEnded) {
  //         console.log("正常播放结束")
  //       } else {
  //         console.log("播放中途退出")
  //       }
  //     })
  //   }
  // },

  nav() {
    // if (rewardedVideoAd) {
    //   rewardedVideoAd.show().catch(() => {
    //     // 失败重试
    //     rewardedVideoAd.load()
    //       .then(() => rewardedVideoAd.show())
    //       .catch(err => {
    //         console.log('激励视频 广告显示失败',err)
    //       })
    //   })
    // }

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


  pickerChange(e) {
    let id = e.currentTarget.id
    let value = Number(e.detail.value)
    if (id == "year") {
      this.setData({
        yearIndex: value,
        majorIndex: 0
      })
      this.getClassList(this.data.colleges[this.data.collegeIndex], this.data.years[value])
    }
    if (id == "college") {
      this.setData({
        collegeIndex: value,
        majorIndex: 0
      })
      this.getClassList(this.data.colleges[value], this.data.years[this.data.yearIndex])
    }
    if (id == "major") {
      this.setData({
        majorIndex: value
      })
      this.handle(this.data.list)
    }
    if (id == "class")
      this.setData({
        classIndex: value,
        target: this.data.classes[value]
      })
  },

  // 获取未处理的班级信息
  getClassList(college = "计算机", year = this.data.years[this.data.yearIndex]) {
    let that = this
    let Obj = new wx.BaaS.TableObject("college_major_class")
    let query = new wx.BaaS.Query()
    query.contains('jgmc', college) //学院
    query.contains('njmc', year) //年级

    Obj.setQuery(query).find().then(res => {
      that.data.list = res.data.objects
      that.handle(res.data.objects)
    })
  },

  // 提取专业班级列表
  handle(list) {

    let majors = []
    let classes = []
    list.forEach(item => {
      if (majors.indexOf(item.zymc) == -1)
        majors.push(item.zymc)
    })
    this.data.majors = majors

    let major = majors[this.data.majorIndex]

    list.forEach(item => {
      if (major == item.zymc && classes.indexOf(item.bj) == -1)
        classes.push(item.bj)
    })

    this.setData({
      classIndex: 0,
      majors: majors,
      classes: classes,
      target: classes[0]
    })

  },


})