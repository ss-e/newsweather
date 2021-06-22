package inet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
var downtimeLength int = 4

//ReadInetDB return inet status database
func ReadInetDB() []Data {
	return InetDB
}

//Startup starts authentication and headline scheduling
func Startup() error {
	readToDB("inet")
	getCurrentInetStatus()
	t1 := schedule(getCurrentInetStatus, 3*time.Minute)
	_ = t1
	return nil
}

//ReadToDB read cities in database
func readToDB(dbname string) {
	// open json file
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		fmt.Println(err)
	}
	err2 := json.Unmarshal(jsonFile, &InetDB)
	if err2 != nil {
		fmt.Println("error reading inet db: ", err2)
	}
	fmt.Println("readToDB completed successfully")
}

func schedule(f func(), interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			f()
		}
	}()
	return ticker
}

func getCurrentInetStatus() {
	for i := range InetDB {
		InetDB[i].Status = make([]StatusData, 0)
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", InetDB[i].URL, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			fmt.Println("inet netclient error")
			continue
		}
		defer response.Body.Close()
		if InetDB[i].Name == "Facebook" {
			var jsonResponse map[string]interface{}
			err := json.NewDecoder(response.Body).Decode(&jsonResponse)
			if err != nil {
				fmt.Println("Error decoding response from Facebook:", err)
				fmt.Println("dump:", response)
			} else {
				tdb1, ok := jsonResponse["current"].(map[string]interface{})
				if !ok {
					fmt.Println("Error decoding current response from Facebook")
				} else {
					fbookTemp, ok := tdb1["subject"].(string)
					if !ok {
						fmt.Println("Error decoding subject response from Facebook")
					} else {
						var temp StatusData
						temp.Title = fbookTemp
						temp.Content = ""
						InetDB[i].Status = append(InetDB[i].Status, temp)
					}
				}
			}
		} else {
			fp := gofeed.NewParser()
			feed, err := fp.Parse(response.Body)
			if err != nil {
				fmt.Println("Error parsing: "+InetDB[i].Name+" : ", err)
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
				fmt.Println("inet: ", InetDB[i].Name, " parsed successfully with ", strconv.Itoa(len(InetDB[i].Status)), "items")
			}
		}
	}
}
