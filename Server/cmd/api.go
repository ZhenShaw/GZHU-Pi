/**
 * @File: api
 * @Author: Shaw
 * @Date: 2020/5/31 8:37 PM
 * @Desc

 */

package cmd

import (
	"GZHU-Pi/env"
	rt "GZHU-Pi/routers"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/mux"
	"github.com/mo7zayed/reqip"
	"github.com/prest/adapters/postgres"
	"github.com/prest/cmd"
	"github.com/prest/config"
	"github.com/prest/config/router"
	"github.com/prest/middlewares"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strings"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "api 服务",
	Long:  `api 服务`,
	Run: func(cmd *cobra.Command, args []string) {
		webApi()
	},
}

func webApi() {

	env.EnvInit()

	r := mux.NewRouter()

	//获取阿里云函数计算容器内部端口
	port := os.Getenv("FC_SERVER_PORT")
	if port == "" {

		port = fmt.Sprint(env.Conf.App.Port)
		logs.Info("自主部署 Server on port: " + port)
		r = r.PathPrefix("/api/v1").Subrouter()

	} else {

		viper.Set("http.port", port)
		logs.Info("阿里云云函数部署 Server on port: " + port)
		r = r.PathPrefix("/2016-08-15/proxy/GZHU-API/go/api/v1").Subrouter()

	}

	if env.Conf.App.PRest == true {

		logs.Info("启用pRest接口服务")
		runWithPRest(r)

	} else {

		customRouter(r)
		if err := http.ListenAndServe(":"+port, r); err != nil {
			log.Fatal(err)
		}

	}
}

//pRest服务内嵌方式运行（把自定义路由和pRest内置路由合并）
func runWithPRest(r *mux.Router) {

	//同步应用配置 覆盖pRest部分配置
	viper.Set("pg.host", viper.GetString("db.host"))
	viper.Set("pg.user", viper.GetString("db.user"))
	viper.Set("pg.pass", viper.GetString("db.password"))
	viper.Set("pg.port", viper.GetInt("db.port"))
	viper.Set("pg.database", viper.GetString("db.dbname"))
	viper.Set("ssl.mode", viper.GetString("db.sslmode"))
	viper.Set("jwt.key", viper.GetString("secret.jwt"))
	viper.Set("http.port", viper.GetString("app.port"))

	// load config for pREST
	config.Load()

	// Load Postgres Adapter
	postgres.Load()

	// Get pREST app
	n := middlewares.GetApp()

	// Register custom middleware
	n.UseFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		//该中间件用于消除路由前缀对pRest内部路由的影响
		if !strings.Contains(r.URL.Path, "/api/v") {
			_, _ = w.Write([]byte("path must contains api version, such as /api/v1"))
			return
		}
		reg := regexp.MustCompile(`[\s\S]*/api/v\d`)
		match := reg.FindStringSubmatch(r.URL.Path)
		if len(match) > 0 {
			r.URL.Path = strings.ReplaceAll(r.URL.Path, match[0], "")
		}
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		next(w, r)
	})

	n.UseFunc(rt.TableAccessHandle)

	// Get pREST router
	r = router.Get()

	// just for test
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(reqip.GetClientIP(r)))
	})

	// Register custom routes
	customRouter(r)

	os.Args = []string{os.Args[0]}
	// Call pREST cmd
	cmd.Execute()
}

//自定义路由初始化
func customRouter(r *mux.Router) *mux.Router {

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/", Hello)

	r.HandleFunc("/auth", rt.Recover(rt.Auth)).Methods("POST")
	r.HandleFunc("/param", rt.Recover(rt.Param))
	r.HandleFunc("/upload", rt.Recover(rt.Upload))

	//微信公众号接口
	//r.HandleFunc("/wx/check", rt.Recover(rt.WeChatCheck))
	r.HandleFunc("/wx/check", rt.Recover(rt.WxMessage))

	//教务系统
	r.HandleFunc("/jwxt/course", rt.Recover(rt.JWMiddleWare(rt.Course))).Methods("GET", "POST")
	r.HandleFunc("/jwxt/exam", rt.Recover(rt.JWMiddleWare(rt.Exam))).Methods("GET", "POST")
	r.HandleFunc("/jwxt/grade", rt.Recover(rt.JWMiddleWare(rt.Grade))).Methods("GET", "POST")
	r.HandleFunc("/jwxt/classroom", rt.Recover(rt.JWMiddleWare(rt.EmptyRoom))).Methods("GET", "POST")
	r.HandleFunc("/jwxt/achieve", rt.Recover(rt.JWMiddleWare(rt.Achieve))).Methods("GET", "POST")
	r.HandleFunc("/jwxt/all-course", rt.Recover(rt.JWMiddleWare(rt.AllCourse))).Methods("GET", "POST")
	r.HandleFunc("/jwxt/rank", rt.Recover(rt.Rank)).Methods("GET")

	//图书馆
	r.HandleFunc("/library/search", rt.Recover(rt.BookSearch)).Methods("GET")
	r.HandleFunc("/library/holdings", rt.Recover(rt.BookHoldings)).Methods("GET")

	//第二课堂学分系统
	r.HandleFunc("/second/my", rt.Recover(rt.SecondMiddleWare(rt.MySecond))).Methods("GET", "POST")
	r.HandleFunc("/second/search", rt.Recover(rt.SecondMiddleWare(rt.SecondSearch))).Methods("GET", "POST")
	r.HandleFunc("/second/image", rt.Recover(rt.SecondMiddleWare(rt.SecondImage))).Methods("GET", "POST")

	//物理实验平台
	//r.HandleFunc("/exp", test).Methods("POST")

	//r.HandleFunc("/jwxt/course2", rt.Recover(rt.Course2)).Methods("POST")

	//四六级、普通话考试查询
	r.HandleFunc("/cet", rt.Recover(rt.GetCet)).Methods("GET")
	r.HandleFunc("/pastCetCaptcha", rt.Recover(rt.GetPastCetCaptcha)).Methods("GET")
	r.HandleFunc("/pastCet", rt.Recover(rt.GetPastCet)).Methods("GET")
	//r.HandleFunc("/exam/chinese", test).Methods("POST")
	return r
}

func Hello(w http.ResponseWriter, r *http.Request) {
	logs.Debug(r)
	m := make(map[string]interface{})
	m["header"] = r.Header
	m["url"] = r.URL.String()

	_, err := json.Marshal(m)
	if err != nil {
		logs.Error(err)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logs.Error(err)
	}
	logs.Info(string(body), 111)
	w.Write(body)
}
