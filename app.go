package main

//noinspection ALL
import (
	"encoding/json"
	"github.com/Tnze/CoolQ-Golang-SDK/v2/cqp"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type dwz struct {
	Code     int    `json:"Code"`
	ShortUrl string `json:"ShortUrl"`
	LongUrl  string `json:"LongUrl"`
	ErrMsg   string `json:"ErrMsg"`
}

type news struct {
	Id           int    `json:"id"`
	PubDate      int64  `json:"pubDate"`
	PubDateStr   string `json:"pubDateStr"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	InfoSource   string `json:"infoSource"`
	SourceUrl    string `json:"sourceUrl"`
	ProvinceId   string `json:"provinceId"`
	ProvinceName string `json:"provinceName"`
	CreateTime   int64  `json:"createTime"`
	ModifyTime   int64  `json:"modifyTime"`
}

var (
	group    int64 = 683750159
	first    bool  = true
	enabled  bool  = true
	nCoVnews []news
)

//go:generate cqcfg -c .
// cqp: 名称: 2019-nCoV
// cqp: 版本: 1.0.0:0
// cqp: 作者: sbbtd
// cqp: 简介: 2019-nCoV监控
func main() { /*此处应当留空*/ }

func init() {
	cqp.AppID = "com.sbbtd.ncov"
	cqp.GroupMsg = onGroupMsg
	cqp.Enable = onEnable
	cqp.Disable = onDisable
}

func onEnable() int32 {
	enabled = true
	cqp.AddLog(cqp.Info, "启动", "群号："+strconv.FormatInt(group, 10))
	go check()
	return 0
}

func onDisable() int32 {
	enabled = false
	return 0
}

func check() {
	refresh(first)
	for {
		if enabled {
			refresh(first)
		}
		time.Sleep(30 * time.Second)
	}
}

func refresh(f bool) {
	cqp.AddLog(cqp.Info, "刷新", "刷新2019-nCov信息...")
	var nCoVnews1 []news
	r, err := http.Get("https://3g.dxy.cn/newh5/view/pneumonia")
	if err == nil && r != nil {
		b, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			return
		}
		s := string(b)
		s = s[strings.Index(s, `getTimelineService = [`)+len(`getTimelineService = `):]
		s = s[:strings.Index(s, `}catch(e){}</script>`)]
		err = json.Unmarshal([]byte(s), &nCoVnews1)
		if err != nil {
			log.Println(err)
			return
		}
		if len(nCoVnews1) < 5 {
			return
		}
		if f {
			nCoVnews = nCoVnews1
			first = false
		} else {
			cqp.AddLog(cqp.Info, "结果", "刷新"+strconv.Itoa(len(nCoVnews1))+"条")
			for _, a := range nCoVnews1 {
				if !isIn(nCoVnews, a) {
					msg := "[" + strings.ReplaceAll(a.ProvinceName, "省", "") + "]" + a.Title + "\nVia:" + a.InfoSource + ", " + a.PubDateStr + "\n"
					if a.ProvinceName == "湖北省" || a.ProvinceName == "黑龙江省" {
						msg += a.Summary + "\n"
					}
					msg += tryGetShortURL(a.SourceUrl)
					cqp.SendGroupMsg(group, msg)
					nCoVnews = append(nCoVnews, a)
				}
			}
		}
	}
}

func tryGetShortURL(lurl string) string {
	appkey := "" //TODO get from https://dwz.cn/
	bdurl := "https://dwz.cn/admin/v2/create"
	postData := strings.NewReader(`{"Url":"` + lurl + `","TermOfValidity":"1-year"}`)
	var dwzResp dwz
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, bdurl, postData)
	if err != nil {
		return lurl
	}
	req.Header.Add("Token", appkey)
	resp, err := client.Do(req)
	if err != nil {
		return lurl
	}
	respStr, err := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	err = json.Unmarshal(respStr, &dwzResp)
	if err != nil {
		return lurl
	}
	return dwzResp.ShortUrl
}

func isIn(s []news, a news) bool {
	for _, n := range s {
		if n.Id == a.Id {
			return true
		}
	}
	return false
}

func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if fromGroup != group {
		return 0
	}
	if msg == "开启追踪" {
		cqp.SendGroupMsg(fromGroup, "已"+msg)
		enabled = true
	} else if msg == "停止追踪" {
		cqp.SendGroupMsg(fromGroup, "已"+msg)
		enabled = false
	}
	return 0
}
