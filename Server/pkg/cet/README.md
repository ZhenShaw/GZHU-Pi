# 四六级历史成绩接口文档

## 获取验证码图片

### 请求示例

HTTP方法：GET

请求接口：/pastCetCaptcha

### 请求参数

无需任何参数

### 返回说明

响应头信息

| 参数       | 值            |
| ---------- | ------------- |
| Set-Cookie | verify=xxxxxx |

示例：

Set-Cookie ：verify=df8cd030f255f18cd8934faa95526053

注：获取成绩列表时带回，用于验证验证码

返回参数

| 参数名 | 必选 | 类型   | 说明                                             |
| ------ | ---- | ------ | ------------------------------------------------ |
| status | 是   | int    | 响应码                                           |
| msg    | 是   | string | 响应信息                                         |
| data   | 是   | string | 服务端返回的数据，该接口返回验证码的base64字符串 |
| ···    | ···  | ···    | ···                                              |



返回示例

```bash
{
    "status": 200,
    "msg": "request ok",
    "data": 图片的base64编码,
    "api": "/pastCetCaptcha",
    "method": "GET",
    "count": 0,
    "time": 2073,
    "update_time": "2020-11-04 08:48:47"
}
```

## 获取成绩列表

### 请求示例

HTTP方法：GET

请求接口：/ pastCet 

### 请求参数

URL参数：

| 参数    | 必选 | 类型   | 说明                               |
| ------- | ---- | ------ | ---------------------------------- |
| subject | 是   | string | 考试类型，如：CET4，CET6，必须大写 |
| xm      | 是   | string | 考生姓名                           |
| sfz     | 是   | string | 考生身份证号                       |
| captcha | 是   | string | 用户输入验证码                     |

完整URL示例：

 /pastCet?subject=CET4&xm=姓名&sfz=身份证&captcha=验证码

### 返回说明

返回参数

| 参数名 | 必选 | 类型   | 说明                                         |
| ------ | ---- | ------ | -------------------------------------------- |
| status | 是   | int    | 响应码                                       |
| msg    | 是   | string | 响应信息                                     |
| data   | 是   | string | 服务端返回的数据，该接口返回考生历史全部成绩 |
| ···    | ···  | ···    | ···                                          |

data数据内容：

| 参数名     | 必选 | 类型   | 说明         |
| ---------- | ---- | ------ | ------------ |
| subject    | 是   | string | 考试名称     |
| person_id  | 是   | string | 考生身份证号 |
| name       | 是   | string | 考生姓名     |
| score_list | 是   | array  | 历史成绩列表 |

sorceList数据内容：

| 参数名    | 必选 | 类型   | 说明         |
| --------- | ---- | ------ | ------------ |
| school    | 是   | string | 考试学校名称 |
| test_id | 是   | string | 笔试准考证号 |
| total     | 是   | int    | 总成绩       |
| listening | 是   | int    | 听力         |
| reading   | 是   | int    | 阅读         |
| writing   | 是   | int    | 写作         |
| voice_id  | 是   | string | 口语准考证号 |
| voice     | 是   | string | 口语等级   |
| score_id  | 是   | string | 成绩单编号   |
| exam_date       | 是   | string | 笔试日期     |
| exam_voice_date | 是   | string | 口试日期     |



返回示例

```bash
{
    "status": 200,
    "msg": "request ok",
    "data": {
        "subject": "全国大学英语四级考试(CET4)",
        "person_id": 身份证号码,
        "name": 姓名,
        "score_list": [
            {
                "school": 考场学校名称,
                "test_id": 准考证号,
                "total": 500,
                "listening": 150,
                "reading": 200,
                "writing": 150,
                "voice_id": "--",
                "voice": "--",
                "score_id": 成绩单编号,
                "exam_date": "2019年12月",
                "exam_voice_date": "--"
            },
            {
				···
            }
        ]
    },
    "api": "/pastCet",
    "method": "GET",
    "count": 0,
    "time": 193,
    "update_time": "2020-11-04 23:43:17"
}
```
