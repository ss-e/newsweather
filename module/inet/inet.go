package inet

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"../debug"

	"github.com/mmcdole/gofeed"
)

//StatusData struct for inet status data
type StatusData struct {
	Title   string
	Content string
}

//Data struct for inet status
type Data struct {
	Name         string
	Status       []StatusData
	URL          string
	Selector     int
	Date         int
	Rootselector string
}

//InetDB contains all inet status info
var InetDB []Data

//downtimeLength when to prune entries, in hours
const (
	downtimeLength = 4
	userAgent      = "newsweather/0.1"
)

//ReadInetDB return inet status database
func ReadInetDB() []Data {
	return InetDB
}

func debugOutput(t string) {
	debug.Output("inet", t)
	return
}

//Startup starts authentication and headline scheduling
func Startup(nc *http.Client) error {
	readToDB("inet")
	getCurrentInetStatus(nc)
	t1 := schedule(getCurrentInetStatus, 3*time.Minute, nc)
	_ = t1
	return nil
}

//ReadToDB read cities in database
func readToDB(dbname string) {
	// open json file
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		debugOutput("Error reading db:" + err.Error())
	}
	err2 := json.Unmarshal(jsonFile, &InetDB)
	if err2 != nil {
		debugOutput("Error unmarshalling db:" + err2.Error())
	}
}

func schedule(f func(*http.Client), interval time.Duration, nc *http.Client) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			f(nc)
		}
	}()
	return ticker
}

func getCurrentInetStatus(nc *http.Client) {
	for i := range InetDB {
		InetDB[i].Status = make([]StatusData, 0)
		req, err := http.NewRequest("GET", InetDB[i].URL, nil)
		req.Header.Set("user-agent", userAgent)
		response, err := nc.Do(req)
		if err != nil {
			debugOutput("inet nc error")
			continue
		}
		defer response.Body.Close()
		if InetDB[i].Name == "Facebook" {
			var jsonResponse map[string]interface{}
			err := json.NewDecoder(response.Body).Decode(&jsonResponse)
			if err != nil {
				debugOutput("Error decoding response from Facebook:" + err.Error())
				continue
			} else {
				tdb1, ok := jsonResponse["current"].(map[string]interface{})
				if !ok {
					debugOutput("Error decoding current response from Facebook")
					continue
				} else {
					fbookTemp, ok := tdb1["subject"].(string)
					if !ok {
						debugOutput("Error decoding subject response from Facebook")
						continue
					} else {
						var temp StatusData
						temp.Title = fbookTemp
						temp.Content = ""
						InetDB[i].Status = append(InetDB[i].Status, temp)
						debugOutput("inet: " + InetDB[i].Name + " parsed successfully with " + strconv.Itoa(len(InetDB[i].Status)) + "items")
					}
				}
			}
		} else {
			fp := gofeed.NewParser()
			feed, err := fp.Parse(response.Body)
			if err != nil {
				debugOutput("Error parsing: " + InetDB[i].Name + " : " + err.Error())
				var temp StatusData
				temp.Title = "OK"
				temp.Content = "OK"
				InetDB[i].Status = append(InetDB[i].Status, temp)
			} else {
				for y := range feed.Items {
					var temp StatusData
					now := time.Now()
					checktime := time.Now()
					if feed.Items[y].UpdatedParsed != nil {
						checktime = *feed.Items[y].UpdatedParsed
					} else {
						checktime = *feed.Items[y].PublishedParsed
					}
					if checktime.After(now.Add(-(time.Duration(downtimeLength) * time.Hour))) {
						temp.Title = feed.Items[y].Title
						if feed.Items[y].Content != "" {
							temp.Content = feed.Items[y].Content
						} else if feed.Items[y].Description != "" {
							temp.Content = feed.Items[y].Description
						} else {
							temp.Content = "no content"
						}
						InetDB[i].Status = append(InetDB[i].Status, temp)
					}
					if InetDB[i].Status == nil {
						temp.Title = "OK"
						temp.Content = "no content"
						InetDB[i].Status = append(InetDB[i].Status, temp)
					}
				}
				debugOutput("inet: " + InetDB[i].Name + " parsed successfully with " + strconv.Itoa(len(InetDB[i].Status)) + "items")
			}
		}
	}
}
