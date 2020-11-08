package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/iris/v12"
	_ "github.com/kataras/iris/v12"
	"io/ioutil"
	"net/http"
)

func getConnect() *sql.DB {
	connection, _ := sql.Open("mysql", "bathapp:bathapp123@tcp(106.15.195.193:3306)/bathapp")
	return connection
}

func main() {
	app := iris.Default()
	app.Use(myMiddleware)
	app.Post("/api/wxlogin", wxLogin)
	//app.Post("/api/hook", hook)
	app.Get("/api/getUserStatus", getUserStatus)
	app.Post("/api/getUserStatus", getUserStatus)
	app.Get("/api/getDormStatus", getDormStatus)
	app.Post("/api/getDormStatus", getDormStatus)
	app.Post("/api/appoint", appoint)
	app.Run(iris.Addr(":5000"), iris.WithConfiguration(iris.Configuration{
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		TimeFormat:                        "Mon, 02 Jan 2006 15:04:05 GMT",
		Charset:                           "UTF-8",
	}))
	fmt.Println("hello")
}

type user struct {
	userid int
	name   string
	dormid int
	openid string
}

func myMiddleware(ctx iris.Context) {
	ctx.Application().Logger().Infof("Runs before %s", ctx.Path())
	ctx.Next()
}
func wxLogin(ctx iris.Context) {
	appid := "wx8018dd5d3153547c"
	secret := "d0daa59c08b75198d8870fb470529694"
	js_code := ctx.PostValue("js_code")
	resp, _ := http.Get("https://api.weixin.qq.com/sns/jscode2session?appid=" + appid + "&secret=" + secret + "&js_code=" + js_code)
	body, _ := ioutil.ReadAll(resp.Body)
	var data iris.Map
	json.Unmarshal(body, &data)
	openid := data["openid"].(string)
	s := "SELECT COUNT(*) FROM `user` WHERE `openid`='" + openid + "'"
	res := getConnect().QueryRow(s)
	var i int
	res.Scan(&i)
	if i == 0 {
		ctx.JSON(iris.Map{"status": "hook", "openid": openid})
	} else {
		s := "SELECT `userid` FROM `user` WHERE `openid`='" + openid + "'"
		res := getConnect().QueryRow(s)
		var userid string
		res.Scan(&userid)
		ctx.JSON(iris.Map{"status": "success", "userid": userid})
	}
}
func hook(ctx iris.Context) {
	var userid string
	var openid string
	if ctx.Method() == "POST" {
		userid = ctx.PostValue("userid")
		openid = ctx.PostValue("openid")
	} else {
		ctx.HTML("<h1>请求方式错误！</h1>")
	}
	s := "UPDATE `user` SET `openid`='" + openid + "' WHERE `userid`" + userid
	getConnect().Exec(s)
	ctx.JSON(iris.Map{"status": "success"})
}
func getUserStatus(ctx iris.Context) {
	var userid string
	var name string
	var dormid string
	if ctx.Method() == "POST" {
		userid = ctx.PostValue("userid")
	} else if ctx.Method() == "GET" {
		userid = ctx.URLParam("userid")
	} else {
		ctx.HTML("<h1>请求方式错误！</h1>")
	}
	s := "SELECT `name`,`dormid` FROM `user` WHERE `userid`=" + userid
	res := getConnect().QueryRow(s)
	res.Scan(&name, &dormid)
	userData := iris.Map{"name": name, "userid": userid, "dormid": dormid}
	s = "SELECT COUNT(*) FROM `appointment` WHERE `userid`=" + userid
	res = getConnect().QueryRow(s)
	var i int
	res.Scan(&i)
	appointment := iris.Map{"con": false}
	if i == 1 {
		s = "SELECT `bathid`,`starttime`,`endtime` FROM `appointment` WHERE `userid`=" + userid
		res = getConnect().QueryRow(s)
		var (
			bathid    string
			starttime string
			endtime   string
		)
		res.Scan(&bathid, &starttime, &endtime)
		appointment["bathid"] = bathid
		appointment["starttime"] = starttime
		appointment["endtime"] = endtime
		appointment["con"] = true
	}
	ctx.JSON(iris.Map{"status": "success", "userData": userData, "appointment": appointment})
}

func getDormStatus(ctx iris.Context) {
	var dormid string
	var bathroom []iris.Map
	if ctx.Method() == "POST" {
		dormid = ctx.PostValue("dormid")
	} else if ctx.Method() == "GET" {
		dormid = ctx.URLParam("dormid")
	} else {
		ctx.HTML("<h1>请求方式错误！</h1>")
	}
	s := "SELECT `bathid`,`con` FROM `dorm` WHERE `dormid`=" + dormid
	res, _ := getConnect().Query(s)
	now := 0
	total := 0
	var (
		bathid string
		con    string
	)
	for res.Next() {
		res.Scan(&bathid, &con)
		singleroom := iris.Map{"bathid": bathid, "con": con}
		bathroom = append(bathroom, singleroom)
		bathroom[total] = singleroom
		total += 1
		if con == "empty" {
			now += 1
		}
	}
	dormData := iris.Map{"dormid": dormid, "now": now, "total": total, "bathroom": bathroom}
	ctx.JSON(iris.Map{"status": "success", "dormData": dormData})
}

func appoint(ctx iris.Context) {
	var userid string
	var bathid string
	var startTime string
	var endTime string
	if ctx.Method() == "POST" {
		userid = ctx.PostValue("userid")
		bathid = ctx.PostValue("bathid")
		startTime = ctx.PostValue("startTime") + ":00"
		endTime = ctx.PostValue("endTime") + ":00"
	} else {
		ctx.HTML("<h1>请求方式错误！</h1>")
	}
	getConnect().Exec("INSERT INTO `appointment` (`userid`,`bathid`,`starttime`,`endtime`) VALUES (?,?,?,?)", userid, bathid, startTime, endTime)
	ctx.JSON(iris.Map{"status": "success"})
}
func userGetIn(ctx iris.Context) {

}
func userGetOut(ctx iris.Context) {

}
