package inet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"strings"
	"github.com/mmcdole/gofeed"
	"time"
)

//Data struct for inet status
type Data struct {
	Name         string
	Status       []string
	URL          string
	Selector     int
	Date         int
	Rootselector string
}

//InetDB contains all inet status info
var InetDB []Data

//ReadInetDB return weatherdb
func ReadInetDB() []Data {
	//return []string{"test","2222"}
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
	//fmt.Println("attempting to read: ", dbname)
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("read inet db, unmarshalling")
	err2 := json.Unmarshal(jsonFile, &InetDB)
	if err2 != nil {
		fmt.Println("experienced error")
		fmt.Println("error reading inet db: ", err2)
		//fmt.Println("dump:", jsonFile)
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
		//thisTime := time.Now()
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", InetDB[i].URL, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			fmt.Println("inet netclient error")
		}
		defer response.Body.Close()
		if InetDB[i].Name == "Facebook" {
			var jsonResponse map[string]interface{}
			err := json.NewDecoder(response.Body).Decode(&jsonResponse)
			if err != nil {
				fmt.Println("Error decoding response from Facebook:", err)
				fmt.Println("dump:", response)
			} else {
				//fmt.Println("response:", jsonResponse)
				tdb1, ok := jsonResponse["current"].(map[string]interface{})
				if !ok {
					fmt.Println("Error decoding current response from Facebook")
				} else {
					//fmt.Println("tdb1 success - inet", tdb1)
					fbookTemp, ok := tdb1["subject"].(string)
					if !ok {
						fmt.Println("Error decoding subject response from Facebook")
					} else {
						//fmt.Println("tdb1 input", InetDB[i].Status[0])
						InetDB[i].Status[0] = fbookTemp
					}
				}
			}
		} else {
			InetDB[i].Status = []string{}
			fp := gofeed.NewParser()
			feed, err := fp.Parse(response.Body)
			if err != nil {
				fmt.Println("Error parsing: "+InetDB[i].Name+" : ", err)
				InetDB[i].Status = []string{"OK"}
			} else {
				fmt.Println("for item: ", InetDB[i].Name, " length is: ", len(InetDB[i].Status))
				for y := range feed.Items {
					//fmt.Println("checking item ", y, " with values: ", feed.Items[y])
					now := time.Now()
					checktime := time.Now()
					if feed.Items[y].UpdatedParsed != nil {
						checktime = *feed.Items[y].UpdatedParsed
					} else {
						checktime = *feed.Items[y].PublishedParsed
					}
					//fmt.Println("checking time value ", checktime)
					if checktime.After(now.Add(time.Duration(12) * time.Hour)) {
						//fmt.Println("item is after")
						InetDB[i].Status = append(InetDB[i].Status, feed.Items[y].Title)
						fmt.Println("appended")
					} else {
						//fmt.Println("item is not after")
					}
					if InetDB[i].Status == nil {
						InetDB[i].Status = append(InetDB[i].Status, "OK")
					}
				}
				fmt.Println("inet: " + InetDB[i].Name + " parsed successfully")
			}
		}
	}
}
