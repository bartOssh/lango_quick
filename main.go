package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/pterm/pterm"
)

// FirebaseTranslations relation input to translations
type FirebaseTranslations struct {
	Input      string       `json:"input"`
	Translated map[string]string `json:"translated"`
}

var atomicRWTranslations struct {
	mu sync.RWMutex
	ct *map[string]map[string]string
}

func mapToLanguage(ft []FirebaseTranslations) *map[string]map[string]string {
	progressbar, err := pterm.DefaultProgressbar.WithTotal(len(ft)).Start()
	if err != nil {
		panic(err)
	}
	defer progressbar.Stop()
	languages := make(map[string]map[string]string)
	for i, group := range ft {
		progressbar.Title = fmt.Sprintf("Adding %v translation", i)
		pterm.Success.Printf("Translation %v added", i)
		progressbar.Increment()
		for ln, trans := range group.Translated {
			if l, ok := languages[ln]; ok {
				if _, oko := l[group.Input]; !oko {
					l[group.Input] = trans
					languages[ln] = l
				}
			} else {
				l := make(map[string]string)
				l[group.Input] = trans
				languages[ln] = l
			}
		}
	}
	return &languages
}

func getTranslations(ctx iris.Context) {
	lang := ctx.Params().GetString("ln")
  
	if lang == "" {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	
	atomicRWTranslations.mu.RLock()
	defer atomicRWTranslations.mu.RUnlock()
	translations, ok := (*atomicRWTranslations.ct)[lang]
	if !ok {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	ctx.JSON(translations)
}

func init() {

	header := pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgRed))
	pterm.DefaultCenter.Println(header.Sprint("Foodie Translations Micro-service"))

	_ = pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString("Lan_go quick")).Render()

	godotenv.Load()
	fBToken := os.Getenv("FIREBASE_FUNC_TOKEN")
	fBTranslationAddress := os.Getenv("FIREBASE_FUNC_ADDRESS")
	fBUserEmail := os.Getenv("FIREBASE_USER_EMAIL_ADDRESS")

	postBody, err := json.Marshal(map[string]string{
		"email": fBUserEmail,
		"token": fBToken,
	})
	if err != nil {
		log.Fatalf("cannot initialize post body, error: %s", err)
	}

	tr := &http.Transport{
		MaxIdleConns:       MaxIdleConn,
		IdleConnTimeout:    IdleConnTimeout * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	
	rawBody := bytes.NewBuffer(postBody)
	resp, err := client.Post(fBTranslationAddress, "application/json", rawBody)
	if err != nil {
		log.Fatalf("cannot fetch initial translation value, error: %s", err)
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("cannot read initial response from firebase translation end point %s", err)
	}

	translations := new([]FirebaseTranslations)
	err = json.Unmarshal(result, translations)
	if err != nil {
		log.Printf("error decoding response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Fatalf("cennot unmarshal firebase translation result %s for response %s", err, result)
	}

	atomicRWTranslations.mu = sync.RWMutex{}
	atomicRWTranslations.mu.Lock()
	defer atomicRWTranslations.mu.Unlock()
	atomicRWTranslations.ct = mapToLanguage(*translations)

	pterm.DefaultSection.Printf("properly initialized lango quick microservice with %v languages", len(*atomicRWTranslations.ct))
}

func main() {
	srvAddressAndPort := os.Getenv("SERVER_ADDRESS_AND_PORT")

	app := iris.New()
  	app.Handle("GET", "/translations/{ln:string}", getTranslations)
  	app.Listen(srvAddressAndPort)
}
