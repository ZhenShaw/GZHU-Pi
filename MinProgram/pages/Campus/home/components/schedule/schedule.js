var utils = require("../../../../../utils/utils.js")
var Data = require("../../../../../utils/data.js")
var Config = require("../../../../../utils/config.js")
var showTimes = 0
var maxWeeks = 25
Component({
  properties: {
    show: {
      type: Boolean,
      value: false
    },
  },

  data: {
    hideTimeLine: !Config.get("tips").time_line,
    showDetail: false,
    current: 1,
    dis: "none",
    today: new Date().getDate(), //日期
    week: utils.getSchoolWeek(), //周数
    schoolWeek: utils.getSchoolWeek(), //校历周
    weekDate: utils.setWeekDate(), //一周日期
    bg: Config.get("schedule_bg"), // 获取背景
    blur: Config.get("blur"), //高斯模糊

    weekDays: Data.weekDays,
    timeLine: Data.timeLine,
    colors: Data.colors,
    kbList: Data.course_sample,

    // 选择周次
    showSelectWeek: false,
    // 最大周数25
    maxWeeks: maxWeeks,
  },

  methods: {
    changeWeek() {
      // this.animate(
      //   "#schedule",
      //   [
      //     { scale: [1, 1], ease: "ease" },
      //     { scale: [0.98, 0.98], ease: "ease-in-out" },
      //     { scale: [0.99, 0.99], ease: "ease" },
      //   ],
      //   100
      // );
      this.setData({
        // scale: 0.98,
        showSelectWeek: true,
      });
    },
    // 修改当前week
    setWeek(e) {
      this.setData({
        week: Number(e.detail),
        weekDate: utils.setWeekDate(e.detail - this.data.schoolWeek),
      });
    },

    closewin() {
      // this.animate(
      //   "#schedule",
      //   [{ scale: [0.99, 0.99] }, { scale: [1, 1] }],
      //   100
      // );
      this.setData({
        showSelectWeek: false,
      });
    },

    // 恢复校历周
    resetWeek() {
      let week = utils.getSchoolWeek()
      this.setData({
        week: week,
        weekDate: utils.setWeekDate(),
      })
      wx.showToast({
        title: "校历 " + String(week) + " 周",
        icon: "none",
        duration: 1000
      })
    },
    // 左右滑动切换周数
    switchWeek(e) {
      let value = e.detail.current - this.data.current
      let week
      if (value == 1 || value == -2) {
        // 下一周
        if (this.data.week + 1 > maxWeeks) {
          week = 0
        } else if (this.data.week < 0) {
          week = 1
        } else {
          week = this.data.week + 1
        }
      } else {
        // 上一周
        if (this.data.week - 1 < 0) {
          week = maxWeeks
        } else {
          week = this.data.week - 1
        }
      }

      if (week < 1 && this.data.schoolWeek < 1) {
        this.setData({
          // weekDate: utils.setWeekDate(week - this.data.schoolWeek),
          week: week,
          current: e.detail.current,
        })
      } else if (this.data.schoolWeek < 1) {
        this.setData({
          weekDate: utils.setWeekDate(week),
          week: week,
          current: e.detail.current,
        })
      } else {
        this.setData({
          weekDate: utils.setWeekDate(week - this.data.schoolWeek),
          week: week,
          current: e.detail.current,
        })
      }

      wx.showToast({
        title: "第 " + String(week) + " 周",
        icon: "none",
        duration: 1000
      })
    },
    // 展开时间轴
    tapSlideBar() {
      this.setData({
        hideTimeLine: !this.data.hideTimeLine,
      })

      // 修改时间轴不展开
      let tips = Config.get("tips")
      tips.time_line = false
      Config.set("tips", tips)
    },

    // 课程详情弹窗
    showDetail(e) {
      let that = this
      let id = Number(e.currentTarget.id)
      let day = this.data.kbList[id].weekday
      let start = this.data.kbList[id].start
      let detail = [this.data.kbList[id]]
      // 遍历课表，找出星期和开始节相同的课程
      this.data.kbList.forEach(function (item) {
        if (item.weekday == day && item.start == start) {
          if (that.data.kbList.indexOf(item) != id) detail.push(item)
        }
      })
      this.setData({
        detail: detail,
        showDetail: true,
        currentIndex: 0, //恢复滑动视图索引
      })
      this.showCourseId(0)
    },
    // 左右滑动切换课程
    switchCourse(e) {
      this.showCourseId(e.detail.current)
    },
    // 打开或者切换时更新显示的课程数组索引
    showCourseId(current) {
      let course = this.data.detail[current]
      for (let i = 0; i < this.data.kbList.length; i++) {
        if (course == this.data.kbList[i]) {
          this.data.openTarget = i
        }
      }
    },

    // 关闭课程详情弹窗
    cancelModal() {
      this.setData({
        showDetail: false,
      })
    },

    // 编辑课表
    navTo(e) {
      switch (e.currentTarget.id) {
        case "0": //编辑
          wx.navigateTo({
            url: "/pages/Campus/home/addCourse/addCourse",
          })
          break
        case "1": //添加
          wx.navigateTo({
            url: "/pages/Campus/evaluation/evaluation",
          })
          break
        case "2": //删除
          this.deleteCourse()
          break
      }
    },

    // 删除课程
    deleteCourse() {
      let that = this
      let id = this.data.openTarget
      let obj = wx.getStorageSync("course")
      if (obj == "") return
      obj.course_list.splice(id, 1)
      wx.showModal({
        title: "提醒",
        content: "是否删除当前课程?",
        success: function (e) {
          if (e.confirm) {
            wx.setStorage({
              key: "course",
              data: obj,
              success: function () {
                that.setData({
                  kbList: obj.course_list,
                  showDetail: false,
                })
              }
            })
          }
        }
      })
    },

    // 更新背景视图使用
    updateBg() {
      this.setData({
        bg: Config.get("schedule_bg"),
        blur: Config.get("blur"),
      })
    },
    viewUpdate() {
      let course = wx.getStorageSync("course")
      let exp = wx.getStorageSync("exp")
      if (course != "" || exp != "") {
        let kbList = course == "" ? [] : course.course_list;
        if (Config.get("showExp")) {
          kbList = kbList.concat(exp)
        }
        this.setData({
          kbList: kbList,
          sjkList: course == "" ? [] : course.sjk_course_list
        })

        // 获取最大周数，小于20取20
        for (let i = 0; i < kbList.length; i++) {
          let max = Math.max(...kbList[i].week_section)
          if (max > maxWeeks) {
            maxWeeks = max
          }
        }
      }
    }
  },

  lifetimes: {
    created: function () {},

    attached: function () {
      this.viewUpdate()
    },

    ready: function () {}
  },

  pageLifetimes: {
    show() {
      // 初次onshow不执行
      if (showTimes) {
        this.viewUpdate()
      }
      showTimes++

    }
  }

})