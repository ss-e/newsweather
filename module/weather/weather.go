package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//Data contains critical data for weather
type Data struct {
	Name string
	ID   string
	Tz   string
	Lat  string
	Lon  string
	Now  [2]int
	/*
		W1   [2]int
		W2   [2]int
		W3   [2]int*/
	W [3][2]int
}

var weatherDB []Data
var weatherAPIKey string = "de13a6c0963d7352292091b6f234070b"
var weatherSite string = "https://api.openweathermap.org/data/2.5/"

//ReadWeatherDB return weatherdb
func ReadWeatherDB() []Data {
	//fmt.Println("weatherdb dump: ", weatherDB)
	return weatherDB
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

//Startup performs actions to be done on startup, start timers
func Startup() error {
	readToDB("weather")
	getCurrentTemp()
	go get6hrTemp()
	t1 := schedule(getCurrentTemp, 15*time.Minute)
	_ = t1
	t2 := schedule(get6hrTemp, 6*time.Hour)
	_ = t2
	return nil
}

//ReadToDB read cities in database
func readToDB(dbname string) {
	// open json file
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		fmt.Println("error reading weather db: ", err)
	}
	err = json.Unmarshal(jsonFile, &weatherDB)
	if err != nil {
		fmt.Println("error unmarshalling weather db: ", err)
		//fmt.Println("dump:", jsonFile)
	}
}

//GetCurrentTemp poll database entries for current temperature
func getCurrentTemp() {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	temp := make([][]string, 0)
	t2 := make([]string, 0)
	for i := 0; i < len(weatherDB); i++ {
		//fmt.Println("len is: ", len(t2), "temp len is:", len(temp), " weatherdb len is: ", len(weatherDB))
		if len(t2) >= 20 {
			//fmt.Println("appending t2 to temp")
			temp = append(temp, t2)
			t2 = nil
			t2 = make([]string, 0)
			t2 = append(t2, weatherDB[i].ID)
		} else {
			t2 = append(t2, weatherDB[i].ID)
		}
	}
	temp = append(temp, t2)
	t2 = nil
	//fmt.Println("temp length is:", len(temp))
	for i := 0; i < len(temp); i++ {
		fmt.Println("loading map temperature batch:", i+1, "/", len(temp))
		time.Sleep(5 * time.Second)
		var url = weatherSite + "group?id=" + strings.Join(temp[i], ",") + "&units=metric&appid=" + weatherAPIKey
		//fmt.Println("url is: ", url)
		response, err := netClient.Get(url)
		if err != nil {
			fmt.Println("Error getcurrenttemp()", err)
			continue
		}
		defer response.Body.Close()
		//var temp2 []weatherData
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("error decoding getcurrenttemp:", err)
			fmt.Println("dump:", response)
		} else {
			//fmt.Println("we made it this far:", jsonResponse)
			//fmt.Println("count is", jsonResponse["cnt"].(string)
			responseArr, ok := jsonResponse["list"].([]interface{})
			if !ok {
				message, ok := jsonResponse["message"].([]interface{})
				if !ok {
					fmt.Println("error decoding response from getcurrenttemp, unknown message")
				} else {
					fmt.Println("error decoding response from getcurrenttemp with message:", message)
				}
			} else {
				for j := 0; j < int(jsonResponse["cnt"].(float64)); j++ {
					temp2 := responseArr[j].(map[string]interface{})
					t3 := temp2["main"].(map[string]interface{})
					t4 := temp2["weather"].([]interface{})
					t5 := t4[0].(map[string]interface{})
					id := int(temp2["id"].(float64))
					nowtemp := int(math.Round(t3["temp"].(float64)))
					nowid := int(t5["id"].(float64))
					//fmt.Println("name: ", temp2["name"], "now: ", nowtemp, "weatherid:", nowid)
					for index := range weatherDB {
						t0, _ := strconv.Atoi(weatherDB[index].ID)
						//fmt.Println("comparing ", t0, " & ", id)
						//if strings.Compare(weatherDB[index].ID, id) == 0 {
						if t0 == id {
							//fmt.Println("found at index: ", index)
							weatherDB[index].Now[0] = nowtemp
							weatherDB[index].Now[1] = nowid
							break
						}
					}
				}
			}
		}
	}
	//fmt.Println("temp: ", weatherDB[0].Now.temp, "id: ", weatherDB[0].Now.id)
	//fmt.Println("end dump: ", weatherDB)
}

func get6hrTemp() {
	for i := range weatherDB {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		time.Sleep(5 * time.Second)
		//var url = weatherSite + "onecall?lat=" + weatherDB[i].Lat + "&lon=" + weatherDB[i].Lon + "&exclude=minutely,current&units=metric&appid=" + weatherAPIKey
		var url = weatherSite + "forecast?id=" + weatherDB[i].ID + "&appid=" + weatherAPIKey + "&units=metric&cnt=18"
		//fmt.Println("url is: ", url)
		response, err := netClient.Get(url)
		if err != nil {
			fmt.Println("err getting 6hr temp data:", err)
			continue
		}
		defer response.Body.Close()
		//var temp2 []weatherData
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("error decoding response:", err, "dump: ", response)
		} else {
			responseArr, ok := jsonResponse["list"].(map[string]interface{})
			if !ok {
				message, ok := jsonResponse["message"].([]interface{})
				if !ok {
					fmt.Println("error decoding response for 6 hour temp for index: ", i, " message dump: ", response.Body)
				} else {
					fmt.Println("error decoding response for 6 hour temp for index: ", i, " with message", message)
				}
			} else {
				nowHour := time.Now().Hour()
				h := 6 - (nowHour % 6)
				k := 0
				for j := h; j < 24; j = i + 6 {
					//get main temp
					main := responseArr["main"].([]interface{})
					t1 := main[j].(map[string]interface{})
					weatherDB[i].W[k][0] = int(t1["temp"].(float64))
					//get weather status
					weather := responseArr["weather"].([]interface{})
					t2 := weather[0].(map[string]interface{})
					weatherDB[i].W[k][1] = int(t2["id"].(float64))
					k++
				}
				fmt.Println("weather onecall index:", i, "w1:", weatherDB[i].W[0][0], ",", weatherDB[i].W[0][1], "w2:", weatherDB[i].W[1][0], ",", weatherDB[i].W[1][1], "w3:", weatherDB[i].W[2][0], ",", weatherDB[i].W[2][1])
			}
			/*
				responseArr, ok := jsonResponse["hourly"].([]interface{})
				if !ok {
					message, ok := jsonResponse["message"].([]interface{})
					if !ok {
						fmt.Println("error decoding response for 6 hour temp for index: ", i, " unknown message")
					} else {
						fmt.Println("error decoding response for 6 hour temp for index: ", i, " with message", message)
					}
				} else {
					nowHour := time.Now().Hour()
					h := 6 - (nowHour % 6)
					t1 := responseArr[h].(map[string]interface{})
					weatherDB[i].W1[0] = int(t1["temp"].(float64))
					tt1 := responseArr[6+h].(map[string]interface{})
					weatherDB[i].W2[0] = int(tt1["temp"].(float64))
					ttt1 := responseArr[12+h].(map[string]interface{})
					weatherDB[i].W3[0] = int(ttt1["temp"].(float64))
					t3 := t1["weather"].([]interface{})
					t4 := t3[0].(map[string]interface{})
					weatherDB[i].W1[1] = int(t4["id"].(float64))
					tt3 := tt1["weather"].([]interface{})
					tt4 := tt3[0].(map[string]interface{})
					weatherDB[i].W2[1] = int(tt4["id"].(float64))
					ttt3 := ttt1["weather"].([]interface{})
					ttt4 := ttt3[0].(map[string]interface{})
					weatherDB[i].W3[1] = int(ttt4["id"].(float64))
					fmt.Println("weather onecall index:", i, "w1:", weatherDB[i].W1[0], ",", weatherDB[i].W1[1], "w2:", weatherDB[i].W2[0], ",", weatherDB[i].W2[1], "w3:", weatherDB[i].W3[0], ",", weatherDB[i].W3[1])
				}
			*/
		}
	}
}
