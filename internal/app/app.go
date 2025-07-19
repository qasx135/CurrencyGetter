package app

import (
	"SomeTask/internal/models"
	"bytes"
	"encoding/xml"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type App struct {
	mu           *sync.Mutex
	wg           *sync.WaitGroup
	timeForBegin time.Time
}

func NewApp() *App {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	timeForBegin := time.Now()
	return &App{mu: mu, wg: wg, timeForBegin: timeForBegin}
}

const (
	UrlTemplate = "http://www.cbr.ru/scripts/XML_daily.asp?date_req=%s"
)

func (app *App) Run() {
	var allResults []models.ComputeStruct
	for i := 0; i < 99; i++ {
		app.wg.Add(1)
		go func(day int) {
			defer app.wg.Done()
			dateForReq := app.timeForBegin.AddDate(0, 0, -day)
			resFromReq, err := GetValutes(dateForReq)
			if err != nil {
				log.Println("error getting valutes from daily date", err)
				return
			}
			app.mu.Lock()
			allResults = append(allResults, resFromReq...)
			app.mu.Unlock()
		}(i)
	}
	app.wg.Wait()
	maxE, minE, avrg := ProcessData(allResults)
	printResult(maxE, minE, avrg)
}

func GetValutes(date time.Time) ([]models.ComputeStruct, error) {
	url := fmt.Sprintf(UrlTemplate, date.Format("02/01/2006"))
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %v", err)
	}
	// ЦБ не пропускает простые запросы
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyApp/1.0)")
	req.Header.Set("Accept", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Println("close body failed:", err)
		}
	}(resp.Body)

	//Проверка самых обычных состояний
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("page not found")
		case http.StatusForbidden:
			return nil, fmt.Errorf("forbidden")
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("unauthorized")
		}
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading page: %s", err)
		return nil, fmt.Errorf("error reading page: %s", err)
	}
	decoder := charmap.Windows1251.NewDecoder()
	utf8Body, err := io.ReadAll(transform.NewReader(bytes.NewReader(body), decoder))
	if err != nil {
		return nil, fmt.Errorf("encoding conversion failed: %v", err)
	}

	// ЦБ отдает инфу в кодировке windows-1251, а мне нужно UTF-8
	utf8Str := string(utf8Body)
	utf8Str = strings.Replace(utf8Str, `<?xml version="1.0" encoding="windows-1251"?>`, `<?xml version="1.0" encoding="UTF-8"?>`, 1)
	var valCurs models.ValCurs
	err = xml.Unmarshal([]byte(utf8Str), &valCurs)
	if err != nil {
		log.Printf("error parsing XML: %s", err)
		return nil, fmt.Errorf("error parsing XML: %s", err)
	}
	var computeData []models.ComputeStruct
	for _, valute := range valCurs.Valutes {
		valueStr := strings.ReplaceAll(valute.Value, ",", ".")

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			log.Printf("error parsing value to float: %s", err)
			return nil, fmt.Errorf("error parsing value to float: %s", err)
		}
		// Не весь номинал представлен в единственном экземпляре, поэтому придется получить номинал и и потом получиьт текущий курс
		nominal, err := strconv.ParseInt(valute.Nominal, 10, 32)
		if err != nil {
			log.Printf("error parsing Nominal to int: %s", err)
			return nil, fmt.Errorf("error parsing Nominal to int: %s", err)
		}
		goodCurrency := value / float64(nominal)
		computeStruct := models.ComputeStruct{
			Date:         date,
			NumCode:      valute.NumCode,
			Name:         valute.Name,
			RealCurrency: goodCurrency,
		}
		if err = computeStruct.Validate(); err != nil {
			log.Printf("error validating compute struct: %v, with err %s", computeStruct, err)
		} else {
			computeData = append(computeData, computeStruct)
		}
	}
	return computeData, nil
}

func ProcessData(data []models.ComputeStruct) (models.ComputeStruct, models.ComputeStruct, float64) {
	if len(data) == 0 {
		return models.ComputeStruct{}, models.ComputeStruct{}, 0
	}
	var maxExchRate models.ComputeStruct
	if data[0].NumCode == "960" {
		maxExchRate = data[1]
	} else {
		maxExchRate = data[0]
	}
	minExchRate := data[0]
	averageExchRate := 0.0
	for _, val := range data {
		// Проверка на СДР - не считается отдельной валютой
		if val.RealCurrency >= maxExchRate.RealCurrency && val.NumCode != "960" {
			maxExchRate = val
		}
		if val.RealCurrency <= minExchRate.RealCurrency {
			minExchRate = val
		}
		averageExchRate += val.RealCurrency
	}
	averageExchRate /= float64(len(data))
	return maxExchRate, minExchRate, averageExchRate

}

func printResult(maxExchRat, minExchRat models.ComputeStruct, averageExchRate float64) {
	fmt.Printf("Максимальные значения: \n")
	fmt.Printf("-Максимальный курс единицы валюты к одному рублю: %.6f \n", maxExchRat.RealCurrency)
	fmt.Printf("-Название единицы валюты: %s\n", maxExchRat.Name)
	fmt.Printf("-Дата, когда курс был максимальным: %s \n", maxExchRat.Date.Format("02.01.2006"))

	fmt.Printf("\nМинимальные значения: \n")
	fmt.Printf("-Минимальный курс единицы валюты к одному рублю: %.6f\n", minExchRat.RealCurrency)
	fmt.Printf("-Название единицы валюты: %s\n", minExchRat.Name)
	fmt.Printf("-Дата, когда курс был минимальным: %s\n", minExchRat.Date.Format("02.01.2006"))

	fmt.Printf("\nСреднее значение курса единиц валют к одному рублю: %.6f\n", averageExchRate)
}
