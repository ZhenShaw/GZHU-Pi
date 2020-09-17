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
    top: 50,
    left: 50,
  },

  lifetimes: {
    // 在组件实例进入页面节点树时执行
    attached: function () {
      // 进入动画
      this.animate("#weekChoiceWin", [{ opacity: 0 }, { opacity: 0.9 }], 300);
      this.animate(
        "#box",
        [
          { opacity: 0, scale: [0, 0] },
          { opacity: 0.9, scale: [1, 1] },
          { offset: 0.9, opacity: 0.9, scale: [1.01, 1.01] },
          { opacity: 0.9, scale: [1, 1] },
        ],
        300
      );

      // 初始化周表
      let _allweek = [];
      for (let i = 0; i < this.properties.propMaxWeek; i++) {
        _allweek[i] = i + 1;
      }

      // attached阶段methods中函数挂接在__proto__上
      // 初始化位置
      let position = this.__proto__.getPosition(this.properties.propCurentWeek);

      this.setData({
        allweek: _allweek,
        left: position[0],
        top: position[1],
      });
    },
  },

  pageLifetimes: {},

  methods: {
    // 根据id调tab所在位置
    getPosition: function (id) {
      // xMove
      let _left;
      if (id % 5 != 0) {
        _left = ((id % 5) - 1) * 20;
      } else {
        _left = 80;
      }

      // yMove
      let _top = (Math.ceil(id / 5) - 1) * 20;
      return [_left, _top];
    },

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

    //跳转周表
    jumpToWeek: function (e) {
      let position = this.getPosition(e.target.id);
      this.setData({
        left: position[0],
        top: position[1],
      });

      // 传递数据给父级方法
      this.triggerEvent("neweek", e.target.id);
      setTimeout(() => {
        this.closeWinAnimation();
      }, 600);
    },
  },
});
