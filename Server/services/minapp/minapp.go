/**
 * @File: minapp
 * @Author: Shaw
 * @Date: 2020/9/17 3:25 PM
 * @Desc

 */

package minapp

import (
	"GZHU-Pi/env"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sync"
)

type MinApp struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	Client       *http.Client
}

var minAppClient *MinApp
var once sync.Once

func GetMinApp() (app *MinApp, err error) {
	if minAppClient != nil {
		return minAppClient, nil
	}
	once.Do(func() {
		minAppClient, err = newMinApp(env.Conf.MinApp.ClientID, env.Conf.MinApp.ClientSecret)
	})
	if minAppClient == nil {
		err = fmt.Errorf("get minapp nil")
	}
	return minAppClient, err
}

func newMinApp(clientID, clientSecret string) (app *MinApp, err error) {
	if clientID == "" || clientSecret == "" {
		err = fmt.Errorf("minapp auth config not set")
		logs.Error(err)
		return
	}
	cookieJar, _ := cookiejar.New(nil)
	app = &MinApp{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  "",
		Client: &http.Client{
			Jar: cookieJar,
		},
	}
	err = app.getMinAppToken()
	if err != nil {
		return nil, err
	}
	return
}

func (app *MinApp) getMinAppToken() (err error) {

	//获取code
	url := "https://cloud.minapp.com/api/oauth2/hydrogen/openapi/authorize/"
	form := map[string]string{
		"client_id":     app.ClientID,
		"client_secret": app.ClientSecret,
	}
	body, _ := json.Marshal(&form)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		logs.Error(err)
		return
	}
	req.Header["Content-Type"] = []string{"application/json"}
	resp, err := app.Client.Do(req)
	if err != nil {
		logs.Error(err)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return
	}

	var m = make(map[string]interface{})
	if err = json.Unmarshal(body, &m); err != nil {
		logs.Error(err)
		return
	}
	code := m["code"].(string)

	//获取access_token
	url = "https://cloud.minapp.com/api/oauth2/access_token/"
	var postData = map[string]string{
		"client_id":     app.ClientID,
		"client_secret": app.ClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
	}
	body, _ = json.Marshal(&postData)
	req, err = http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		logs.Error(err)
		return
	}
	req.Header["Content-Type"] = []string{"application/json"}
	resp, err = app.Client.Do(req)
	if err != nil {
		logs.Error(err)
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code: %d", resp.StatusCode)
		logs.Error(err)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	if err = json.Unmarshal(body, &m); err != nil {
		logs.Error(err)
		return
	}
	token := m["access_token"].(string)
	app.AccessToken = token
	return
}

func (app *MinApp) GetTableData(tableID, recordID string) (data map[string]interface{}, err error) {

	url := fmt.Sprintf("https://cloud.minapp.com/oserve/v2.4/table/%s/record/%s/", tableID, recordID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error(err)
		return
	}
	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", app.AccessToken)}

	resp, err := app.Client.Do(req)
	if err != nil {
		logs.Error(err)
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code: %d", resp.StatusCode)
		logs.Error(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		logs.Error(err)
		return
	}
	return
}
