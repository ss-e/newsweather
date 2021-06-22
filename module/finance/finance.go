package finance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

//Item db, chartdata is [timestamp, open, high, low, close]
type Item struct {
	Name          string
	Ticker        string
	Value         float64
	Open          float64
	ChangePercent float64
	Chartdata     [][]float64
}

//StockDB contains all stock info
var StockDB []Item

//FxDB contains all foreign currency info
var FxDB []Item

//CryptoDB contains all crypto info
var CryptoDB []Item

var iexapikey = os.Getenv("IEX_APIKEY")
var iexsite = "https://cloud.iexapis.com/"
var cryptoapi = "https://api.cryptowat.ch/markets/binance/"

//ReadStockDB return weatherdb
func ReadStockDB() []Item {
	return StockDB
}

//ReadFxDB return weatherdb
func ReadFxDB() []Item {
	return FxDB
}

//ReadCryptoDB return weatherdb
func ReadCryptoDB() []Item {
	return CryptoDB
}

func getFxInfo() {
	//thisTime := time.Now()
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	ta := []string{}
	for i := range FxDB {
		ta = append(ta, FxDB[i].Ticker)
	}
	temp := strings.Join(ta, ",")
	req, err := http.NewRequest("GET", iexsite+"stable/fx/latest?symbols="+temp+"&token="+iexapikey, nil)
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	var jsonResponse []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		fmt.Println("error:", err)
		fmt.Println("dump:", response)
	} else {
		for i := range jsonResponse {
			ticker, ok := jsonResponse[i]["symbol"].(string)
			if ok {
				for j := range FxDB {
					if ticker == FxDB[j].Ticker {
						value, ok := jsonResponse[i]["rate"].(float64)
						fmt.Println("for: ", ticker, " got value: ", value)
						if ok {
							FxDB[j].Value = value
						} else {
							fmt.Println("error copying response")
						}
						break
					}
				}
			}
		}
	}
}

func getStockInfo() {
	for i := range StockDB {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", iexsite+"stable/stock/"+StockDB[i].Ticker+"/book?token="+iexapikey, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			fmt.Println("err getting stock data:", err)
			continue
		}
		defer response.Body.Close()
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("error:", err)
			fmt.Println("dump:", response)
		} else {
			tdb1, ok := jsonResponse["quote"].(map[string]interface{})
			if !ok {
				fmt.Println("Finance error response to getstockinfo for ticker item ", StockDB[i].Ticker)
			} else {
				value, ok := tdb1["latestPrice"].(float64)
				open, ok := tdb1["open"].(float64)
				changepercent, ok := tdb1["changePercent"].(float64)
				fmt.Println("for: ", StockDB[i].Ticker, " got value: ", value, " open: ", open, " changepercent: ", changepercent)
				if ok {
					StockDB[i].Value = value
					StockDB[i].Open = open
					StockDB[i].ChangePercent = changepercent * 100
				} else {
					fmt.Println("Finance error copying response in getstockinfo for item ", StockDB[i].Ticker)
				}
			}
		}
	}
}

func getStockChartData() {
	for i := range StockDB {
		fmt.Println("trying: ", StockDB[i].Ticker)
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", iexsite+"stable/stock/"+StockDB[i].Ticker+"/intraday-prices?chartInterval=5&token="+iexapikey, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			fmt.Println("err getting stock chart data:", err)
			continue
		}
		defer response.Body.Close()
		var jsonResponse []interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("error:", err)
			fmt.Println("dump:", response)
		} else {
			StockDB[i].Chartdata = nil
			for j := range jsonResponse {
				tdb1, ok := jsonResponse[j].(map[string]interface{})
				if !ok {
					fmt.Println("Finance error response to chartdatainfo for ticker item ", StockDB[i].Ticker)
				} else {
					//fmt.Println("tdb1 declared: ", tdb1)
					temp, err := time.Parse("2006-01-02 15:04", tdb1["date"].(string)+" "+tdb1["minute"].(string))
					if err != nil {
						//fmt.Println("for ", StockDB[i].Ticker, ", error parsing stock chart data, date ", err)
					}
					date := temp.Unix()
					open, err3 := tdb1["open"].(float64)
					if !err3 {
						//fmt.Println("for ", StockDB[i].Ticker, ", error parsing stock chart data, open ", err3)
					}
					high, err4 := tdb1["high"].(float64)
					if !err4 {
						//fmt.Println("for ", StockDB[i].Ticker, ", error parsing stock chart data, high ", err4)
					}
					low, err5 := tdb1["low"].(float64)
					if !err5 {
						//fmt.Println("for ", StockDB[i].Ticker, ", error parsing stock chart data, low ", err5)
					}
					close, err6 := tdb1["close"].(float64)
					if !err6 {
						//fmt.Println("for ", StockDB[i].Ticker, ", error parsing stock chart data, close ", err6)
					}
					if err == nil && err3 && err4 && err5 && err6 {
						//fmt.Println("inputting entry. [date:", float64(date)*1000, ",open:", open, ",high:", high, ",low:", low, ",close:", close)
						entry := []float64{float64(date) * 1000, open, high, low, close}
						StockDB[i].Chartdata = append(StockDB[i].Chartdata, entry)
					}
				}
			}
		}
	}
}

func getCryptoInfo() {
	for i := range CryptoDB {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", cryptoapi+CryptoDB[i].Ticker+"/summary", nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			fmt.Println("err getting crypto data:", err)
			continue
		}
		defer response.Body.Close()
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("error:", err)
			fmt.Println("dump:", response)
		} else {
			tdb1, ok := jsonResponse["result"].(map[string]interface{})
			if !ok {
				fmt.Println("Finance error response to getCryptoInfo for ticker item ", CryptoDB[i].Ticker)
				fmt.Println("resultdump:", jsonResponse)
			} else {
				tdb2 := tdb1["price"].(map[string]interface{})
				tdb3 := tdb2["change"].(map[string]interface{})
				value, ok := tdb2["last"].(float64)
				changepercent, ok := tdb3["percentage"].(float64)
				if ok {
					CryptoDB[i].Value = value
					CryptoDB[i].ChangePercent = changepercent * 100
					fmt.Println("for: ", CryptoDB[i].Ticker, " got value: ", CryptoDB[i].Value, " changepercent: ", CryptoDB[i].ChangePercent)
				} else {
					fmt.Println("error copying response")
				}
			}
		}
	}
}

func getCryptoChartData() {
	for i := range CryptoDB {
		t1 := time.Now()
		t2 := t1.Add(-24 * time.Hour)
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		thisTime := fmt.Sprintf("%v", t2.Unix())
		fmt.Println("trying url: ", cryptoapi+CryptoDB[i].Ticker+"/ohlc?periods=1800&after="+thisTime)
		req, err := http.NewRequest("GET", cryptoapi+CryptoDB[i].Ticker+"/ohlc?periods=1800&after="+thisTime, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			fmt.Println("err getting crypto chart data:", err)
			continue
		}
		defer response.Body.Close()
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("error:", err)
			fmt.Println("dump:", response)
		} else {
			CryptoDB[i].Chartdata = nil
			tdb1, ok := jsonResponse["result"].(map[string]interface{})
			if !ok {
				fmt.Println("Finance error response to crypto chartdata for ticker item ", CryptoDB[i].Ticker, ". error: ", ok)
				fmt.Println("resultdump:", jsonResponse)
			} else {
				tdb2, ok := tdb1["1800"].([]interface{})
				if !ok {
					fmt.Println("Finance error response to 1800 for ticker item ", CryptoDB[i].Ticker)
				} else {
					for j := range tdb2 {
						tdb3, ok := tdb2[j].([]interface{})
						if ok {
							date, ok2 := tdb3[0].(float64)
							open, ok3 := tdb3[1].(float64)
							high, ok4 := tdb3[2].(float64)
							low, ok5 := tdb3[3].(float64)
							close, ok6 := tdb3[4].(float64)
							if ok2 && ok3 && ok4 && ok5 && ok6 {
								//fmt.Println("inputting entry. [date:", float64(date)*1000, ",open:", open, ",high:", high, ",low:", low, ",close:", close)
								entry := []float64{float64(date) * 1000, open, high, low, close}
								CryptoDB[i].Chartdata = append(CryptoDB[i].Chartdata, entry)
							} else {
								fmt.Println("Finance error 2 parsing crypto chart data for ticker item ", CryptoDB[i].Ticker)
							}
						} else {
							fmt.Println("Finance error 1 parsing crypto chart data for ticker item ", CryptoDB[i].Ticker)
						}
					}
				}
			}
		}
	}
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

//ReadToDB read cities in database
func readToDB(dbname string, database *[]Item) {
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(jsonFile, &database)
	if err != nil {
		fmt.Println("error reading stock db: ", err)
	}
}

//Startup starts authentication and headline scheduling
func Startup() error {
	readToDB("stock", &StockDB)
	readToDB("crypto", &CryptoDB)
	getStockInfo()
	getStockChartData()
	getCryptoInfo()
	getCryptoChartData()
	t1 := schedule(getStockInfo, 3*time.Minute)
	_ = t1
	t2 := schedule(getStockChartData, 15*time.Minute)
	_ = t2
	t3 := schedule(getCryptoInfo, 5*time.Minute)
	_ = t3
	t4 := schedule(getCryptoChartData, 15*time.Minute)
	_ = t4
	return nil
}
