package main

import (
	"./module/finance"
	"./module/inet"
	"./module/news"
	"./module/weather"
	"fmt"
	"github.com/webview/webview"
	"io/ioutil"
	"log"
	"net/url"
)

func readShell() string {
	content, err := ioutil.ReadFile("./shell.html")
	if err != nil {
		log.Fatal(err)
	}
	return "data:text/html," + url.QueryEscape(string(content))
}

func startup() {
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
}

func main() {
	fmt.Println("starting up")
	startup()
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("newsweather")
	//w.SetSize(1280, 720, webview.HintFixed)
	w.SetSize(1920, 1080, webview.HintFixed)
	//shell := "data:text/html,<html><body><p style=\"width:10%;\">Test</p></body></html>"
	//shell := "data:text/html,<html><body>Test</body></html>"
	//w.Navigate(shell)
	//w.Navigate("https://en.m.wikipedia.org/wiki/Main_Page")
	w.Bind("readWeatherDB", weather.ReadWeatherDB)
	w.Bind("readHeadlineDB", news.ReadHeadlineDB)
	w.Bind("readInetDB", inet.ReadInetDB)
	w.Bind("readStockDB", finance.ReadStockDB)
	w.Bind("readCryptoDB", finance.ReadCryptoDB)
	w.Navigate(readShell())
	w.Run()
}
