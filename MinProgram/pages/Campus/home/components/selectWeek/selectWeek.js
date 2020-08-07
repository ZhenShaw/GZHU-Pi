Component({
  properties: {
    propMaxWeek: {
      type: Number,
    },
    propCurentWeek: {
      type: Number,
    },
  },

  data: {
    allweek: [],
  },

  lifetimes: {
    // 在组件实例进入页面节点树时执行
    attached: function () {
      this.animate("#weekChoiceWin", [{ opacity: 0 }, { opacity: 0.9 }], 300);
      this.animate(
        "#box",
        [
          { opacity: 0, scale: [0, 0] },
          { opacity: 0.9, scale: [1, 1] },
          { opacity: 0.9, scale: [1.1, 1.1] },
          { opacity: 0.9, scale: [1, 1] },
        ],
        300
      );
      let _allweek = [];
      for (let i = 0; i < this.properties.propMaxWeek; i++) {
        _allweek[i] = i + 1;
      }
      this.setData({
        allweek: _allweek,
      });
    },
  },

  pageLifetimes: {},

  methods: {
    // 关闭窗口动画
    closeWinAnimation: function () {
      this.animate("#weekChoiceWin", [{ opacity: 0.9 }, { opacity: 0 }], 300);
      this.animate(
        "#box",
        [
          { opacity: 0.9, scale: [1, 1] },
          { opacity: 0, scale: [0, 0] },
        ],
        300
      );
      setTimeout(() => {
        // 调用父级方法隐藏窗口
        this.triggerEvent("closewin");
      }, 300);
    },
    // 关闭
    close: function (e) {
      if (e.target.id === "weekChoiceWin") {
        this.closeWinAnimation();
      }
    },
    jumpToWeek: function (e) {
      console.log(e.target.id);
      // 传递数据给父级方法
      this.triggerEvent("neweek", e.target.id);
      this.closeWinAnimation();
    },
  },
});
