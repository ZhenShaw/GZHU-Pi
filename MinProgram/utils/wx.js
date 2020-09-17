// 全局微信变量
wx.$param = require('param').param


// 封装的wx微信全局方法

wx.$ajax = function (option) {
  return new Promise(function (resolve, reject) {
    if (option.method == undefined || typeof option.method !== "string") {
      option.method = "POST"
    }
    if (option.url == undefined) {
      option.url = wx.$param.server["aliyun_go"]
    }
    if (option.header == undefined || typeof option.header != 'object') {
      option.header = {
        "content-type": "application/x-www-form-urlencoded"
      }
    }
    if (typeof option.url === 'string' && option.url.indexOf("http") == -1) {
      option.url = wx.$param.server["aliyun_go"] + option.url
    }
    if (typeof option.loading == "boolean" && option.loading) {
      wx.showLoading({
        title: '加载中',
        duration: 60000,
        mask: true,
      })
    } else if (typeof option.loading == "string") {
      wx.showLoading({
        title: option.loading,
        duration: 10000,
        mask: true,
      })
    }
    // 携带cookie
    option.header["Cookie"] = wx.getStorageSync("gzhupi_cookie")
    wx.request({
      url: option.url,
      data: option.data,
      method: option.method.toUpperCase(),
      header: option.header,
      success: (res) => {

        // http响应错误
        if (res.statusCode >= 400) {
          // if (res.statusCode == 401) wx.removeStorageSync('gzhupi_cookie')
          let msg = res.data.error
          msg = msg ? msg : res.errMsg
          msg = String(res.statusCode) + " 错误"
          reject({
            when: "http_status_error",
            error: msg,
            detail: msg,
          })
          if (option.showErr == false) return
          wx.showModal({
            title: '提示',
            content: msg,
            showCancel: false
          })
          return
        }

        // 缓存cookies
        if (res.header["Set-Cookie"] != undefined && res.header["Set-Cookie"] != "") {
          wx.setStorageSync("gzhupi_cookie", res.header["Set-Cookie"]);
        }

        // 自定义响应协议(只返回data)
        if (res.data && res.data.status) {
          // if (res.data.status == 401) wx.removeStorageSync('gzhupi_cookie')
          if (res.data.status == 0 || res.data.status == 200) {
            resolve(res.data)
            return
            // } else if (res.data.status == -1) {
          } else {
            let msg = "请求响应失败"
            if (res.data.msg != undefined) msg = res.data.msg
            reject({
              when: "status_error",
              error: msg,
              detail: res.data,
            })
            if (option.showErr == false) return
            wx.showModal({
              title: '提示',
              content: msg,
              showCancel: false
            })
            return
          }
        }
        // 没有使用自定义响应协议
        resolve(res)
        return
      },
      fail: (err) => {
        reject({
          when: "origin_error",
          error: err
        })
        wx.showModal({
          title: '错误提示',
          content: JSON.stringify(err),
          showCancel: false
        })
      },
      complete: (res) => {
        console.log("response :" + option.url, res)
        if (!!option.loading) {
          wx.hideLoading()
        }
      }
    })
  })
}

/**
 * 页面转跳封装
 * @method wx.$navTo
 * @param {object|string}  e    如果是字符串，直接跳转；对象，就解析到e.target.dataset.url
 * @param {object} args         页面参数
 */
wx.$navTo = function (e, args) {
  if (e == undefined && arg == undefined) return
  console.log('fun: navTo', e, args)
  let args_str = []
  if (typeof args === 'object') {
    for (let i in args) {
      args_str.push(i + '=' + encodeURIComponent(args[i]))
    }
    args_str = '?' + args_str.join("&")
  } else {
    args_str = ''
  }
  if (typeof e == 'object') {
    if (e.target.dataset && e.target.dataset.url) {
      wx.navigateTo({
        url: e.target.dataset.url + args_str,
        fail: err => {
          console.warn(err)
          wx.switchTab({
            url: e.target.dataset.url + args_str,
            fail: err => {
              console.err(err)
            }
          })
        }
      })
    } else if (e.currentTarget.dataset && e.currentTarget.dataset.url) {
      wx.navigateTo({
        url: e.currentTarget.dataset.url + args_str,
        fail: err => {
          console.warn(err)
          wx.switchTab({
            url: e.currentTarget.dataset.url + args_str,
            fail: err => {
              console.err(err)
            }
          })
        }
      })
    }
  } else {
    wx.navigateTo({
      url: e + args_str,
      fail: err => {
        console.warn(err)
        wx.switchTab({
          url: e + args_str,
          fail: err => {
            console.err(err)
          }
        })
      }
    })
  }
}


/**
 * 对象转url参数(带问号)
 * @method wx.$objectToQuery
 * @param {object}  obj
 * @return {string} query
 */
wx.$objectToQuery = function (obj = {}) {

  if (typeof obj != 'object') {
    console.error("not object")
    return
  }
  let args_str = []
  for (let i in obj) {

    if (!obj[i]) continue

    args_str.push(i + '=' + encodeURIComponent(obj[i]))
  }
  let query = '?' + args_str.join("&")
  return query
}

//收集订阅消息，回调成功数量
wx.$subscribe = function () {

  let tpls = {
    unread: 'mzClt2VmH5tlVqVpbaKaeGZ2XM2GIYrztuGjNIjRaZw', //未读消息提醒 发送人、消息内容、备注
    comment: "qXh2oaTKaNEBF1UJCjYkTqW4vs24yQCfmShdO4SyvXg", //留言通知	文章标题、留言人、留言内容
  }
  let tmplIds = [tpls["comment"], tpls["unread"]]

  return new Promise(function (resolve) {
    wx.requestSubscribeMessage({
      tmplIds: tmplIds,
      success: (res) => {
        console.log(res)
        let subscription = []
        for (let i in tmplIds) {
          if (res[tmplIds[i]] === 'accept') {
            subscription.push({
              template_id: tmplIds[i],
              subscription_type: 'once',
            })
          }
        }
        if (subscription.length > 0) {
          resolve(subscription.length)
          wx.BaaS.subscribeMessage({
            subscription
          }).then(res => {
            console.log(res)
          }, err => {
            console.error(err)
          })
        } else {
          resolve(0)
        }
      },
      fail: (err) => {
        console.error(err)
        resolve(0)
      }
    })
  })
}

/*
 * 伪双向绑定
 * wxml input 定义属性：data-field="field1.field2" value="{{field1.field2}}"
 * 输入内容将绑定到：this.data.field1.field2 = e.detail.value
 * 
 * <input bindinput="inputChange" data-field="navTitle" value="{{navTitle}}" ></input>
 * inputChange(e){
 *  wx.$bindInput.call(this,e)
 *  }
 */

wx.$bindInput = function (e) {
  if (typeof e.currentTarget.dataset.field != "string") return
  let field = e.currentTarget.dataset.field
  console.log("数据绑定：key：", field, " value:", e.detail.value)

  let data = {}
  data[field] = e.detail.value
  this.setData(data)
}

// 图片预览
/* <image data-url="{{curImg}}" ></image> */
wx.$viewImg = function (urls = [], e) {

  if (urls.length == 0) {
    urls.push(e.currentTarget.dataset.url)
  }

  wx.previewImage({
    urls: urls,
    current: e.currentTarget.dataset.url
  });
}


// 同步认证
wx.$authSync = async function (times = 1) {

  if (times > 3) {
    console.log("转跳认证")
    wx.$navTo("/pages/Setting/login/auth")
    return
  }

  let user = await wx.BaaS.auth.getCurrentUser().then(user => {
    console.log("minapp user", user)
    return user
  }).catch(err => {
    if (err.code === 604) {
      console.log('用户未登录，发送auth请求失败，重试', err)
    }
  })

  // 递归重试
  if (user == undefined || !user.user_id) {
    setTimeout(() => {
      wx.$authSync(times + 1)
    }, 3000);
    return
  }
  let stu_id = wx.getStorageSync("account").username
  if (stu_id == undefined || stu_id == null) {
    stu_id = ""
  }
  let form = {
    stu_id: stu_id,
    minapp_id: user.user_id,
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
  // console.log("auth data: ", form)

  let v_user = await wx.$ajax({
      url: wx.$param.server["prest"] + "/auth",
      // url: "http://localhost:9000/api/v1/auth",
      method: "post",
      showErr: false,
      data: form,
      header: {
        "content-type": "application/json"
      }
    })
    .then(res => {
      console.log("auth resp user:", res.data)
      if (res.data.open_id == user.openid && res.data.id > 0) {
        wx.setStorage({
          key: 'gzhupi_user',
          data: res.data,
        })
      }
    }).catch(err => {
      console.log("转跳认证", err)
      wx.$navTo("/pages/Setting/login/auth")
    })
}


// 登录、cookie检查
wx.$checkUser = function (nav = true) {
  let cookie = wx.getStorageSync("gzhupi_cookie")
  let v_user = wx.getStorageSync("gzhupi_user")

  if (!v_user || v_user.id <= 0 || !cookie) {
    wx.showToast({
      title: '未登录',
      icon: "none"
    })
    if (nav) {
      console.log("转跳认证")
      wx.$navTo("/pages/Setting/login/auth")
    }
    return false
  }

  return true
}

// 违规文本检测
wx.$checkText = async function (text) {
  if (text == "" || text == undefined) return
  var is_risk = await wx.BaaS.wxCensorText(text).then(res => {
    // console.log(res.data.risky)
    if (res.data.risky) {
      wx.showModal({
        title: '警告',
        content: '您的发布内容包含违规词语',
      })
      return res.data.risky
    }
    return res.data.risky
  }, err => {
    console.log(err)
    return false
  })
  return is_risk
}

// 违规图片检测
wx.$checkImg = async function (path) {
  if (text == "" || text == undefined) return
  var is_risk = await wx.BaaS.wxCensorImage(path).then(res => {
    // console.log(res.data.risky)
    if (res.data.risky) {
      wx.showModal({
        title: '警告',
        content: '您选择的图片包含敏感内容',
      })
      return res.data.risky
    }
    return res.data.risky
  }, err => {
    console.log(err)
    return false
  })
  return is_risk
}

// 线上配置 > 缓存配置 > 默认配置
wx.$syncParam = async function () {

  let param = wx.getStorageSync("app_param")
  if (param != "") wx.$param = param

  let app_param = await wx.$ajax({
      url: wx.$param.server["prest"] + "/param?type=app",
      // url: "http://localhost:9000/api/v1/param?type=app",
      method: "get",
      showErr: false,
    })
    .then(res => {
      if (!res || !res.data || !res.data.data) {
        console.error(res)
        return
      }
      console.log("在线配置：", res.data.data)
      // verify模式下 username有效且不是测试账号放行
      let account = wx.getStorageSync("account")
      if (res.data.data.mode == "verify") {
        if (!!account.username && account.username != "20200504" && account.username != "20180829") {
          res.data.data.mode = "prod"
        }
      }
      wx.setStorageSync("app_param", res.data.data)
      wx.$param = res.data.data

    }).catch(err => {
      console.error("app_param", err)
    })

}


wx.$update = function () {
  // 检测更新版本
  const updateManager = wx.getUpdateManager()
  updateManager.onCheckForUpdate(function (res) {
    // 请求完新版本信息的回调
    console.log("新版本：", res.hasUpdate)
  })
  updateManager.onUpdateReady(function () {
    wx.showModal({
      title: '更新提示',
      content: '新版本已经准备好，是否重启应用？\n如遇缓存丢失，请重启小程序。',
      success(res) {
        if (res.confirm) {
          // 新的版本已经下载好，调用 applyUpdate 应用新版本并重启
          updateManager.applyUpdate()
        }
      }
    })
  })
  updateManager.onUpdateFailed(function () {
    // 新版本下载失败
  })
}