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

// firebaseTranslations relation input to translations
type firebaseTranslations struct {
	Input      string       `json:"input"`
	Translated map[string]string `json:"translated"`
}

// requestTranslations allows to decode slice of translations from json request
type requestTranslations struct {
	Translate []string `json:"translate"`
}

type langoClient struct {
	inner *http.Client
	fBToken, fBTranslationAddress, fBUserEmail string
}

// serverAddressAndPort is a default address and port,
// if SERVER_ADDRESS_AND_PORT env variable is provided then variable is overwritten
var srvAddressAndPort = "0.0.0.0:8000"

var sleepTime time.Duration = 60

var atomicRWTranslations struct {
	mu sync.RWMutex
	ct *map[string]map[string]string
}

// global for this service http client
var client langoClient


// mapToLanguage iterate through input : translations slice and maps translations per language key like so:
// {"language_key" : {"translation in english as a key" : "translation to language of the language_key"}}
func mapToLanguage(ft []firebaseTranslations) *map[string]map[string]string {
	progressbar, err := pterm.DefaultProgressbar.WithTotal(len(ft)).Start()
	if err != nil {
		panic(err)
	}
	defer progressbar.Stop()
	languages := make(map[string]map[string]string)
	for i, group := range ft {
		pterm.DefaultCenter.Println(pterm.Success.Sprintf("Translation %v added", i))
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

func renderEndpoints(ct *map[string]map[string]string, addr string) {

	data := pterm.TableData{
		{"language code", "endpoint"},
	}

	for k := range *ct {
		data = append(data, []string{k, fmt.Sprintf("%s/get/%s", addr, k)})
	}


	s, err := pterm.DefaultTable.WithHasHeader().WithData(data).Srender()
	if err != nil {
		log.Fatalf("cannot log endpoints table, error: %s", err)
	}
	header := pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgRed))
	pterm.DefaultCenter.Println(header.Sprint("Translation endpoints list:"))
	pterm.DefaultCenter.Println(s)
}

func getLanguage(ctx iris.Context) {
	lang := ctx.Params().GetString("ln")
  
	if lang == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}
	
	atomicRWTranslations.mu.RLock()
	defer atomicRWTranslations.mu.RUnlock()
	translations, ok := (*atomicRWTranslations.ct)[lang]
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}
	ctx.JSON(translations)
}

func createTranslations(ctx iris.Context) {
	t := &requestTranslations{}
	err := ctx.ReadJSON(t)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
	}
	pterm.DefaultCenter.Println(pterm.Info.Sprintf("Sending request to remote firebase functions to create %v translations", len(t.Translate)))
	if ok := client.postCreateToRemote(t.Translate); !ok {
		pterm.DefaultCenter.Println("Error received from remote firebase function, couldn't create translations")
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}
	pterm.DefaultCenter.Println(pterm.Success.Sprintf("Created %v translations", len(t.Translate)))
	go func() {
		time.Sleep(sleepTime*time.Second)
		client.readTranslationsFromRemote()
	}()
	ctx.StatusCode(iris.StatusAccepted)
}

func (c langoClient) readTranslationsFromRemote() {
	postBody, err := json.Marshal(map[string]string{
		"email": c.fBUserEmail,
		"token": c.fBToken,
	})
	if err != nil {
		log.Fatalf("cannot initialize post body, error: %s", err)
	}
	rawBody := bytes.NewBuffer(postBody)
	resp, err := c.inner.Post(c.fBTranslationAddress + "/translations", "application/json", rawBody)
	if err != nil {
		log.Fatalf("error when posting translation request to remote firebase function, error: %s", err)
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("cannot read initial response from firebase translation end point %s", err)
	}

	translations := new([]firebaseTranslations)
	err = json.Unmarshal(result, translations)
	if err != nil {
		log.Printf("error decoding response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Fatalf("cennot unmarshal firebase translation result %s for response %s", err, result)
	}
	atomicRWTranslations.mu.Lock()
	defer atomicRWTranslations.mu.Unlock()
	atomicRWTranslations.ct = mapToLanguage(*translations)
	pterm.DefaultCenter.Println(pterm.DefaultBasicText.Sprintf("properly initialized lango quick microservice with %v languages", len(*atomicRWTranslations.ct)))
	renderEndpoints(atomicRWTranslations.ct, srvAddressAndPort)
}

func (c langoClient) postCreateToRemote(data []string) (ok bool) {
	payload := struct{
		Email string `json:"email"`
		Token string `json:"token"`
		Translations []string `json:"translations"`
	}{
		Email: c.fBUserEmail,
		Token: c.fBToken,
		Translations: data,	
	}
	postBody, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("cannot initialize post body, error: %s", err)
	}
	rawBody := bytes.NewBuffer(postBody)
	resp, err := c.inner.Post(c.fBTranslationAddress + "/create", "application/json", rawBody)
	if err != nil {
		log.Fatalf("error when posting create request to remote firebase function, error: %s", err)
	}
	if resp.StatusCode !=  http.StatusCreated {
		return
	}
	ok = true
	return
}

func init() {

	s, err := pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString("Lan.go quick")).Srender()
	if err != nil {
		log.Fatalf("cannot log in BIG TEXT, error: %s", err)
	}
	pterm.DefaultCenter.Println(s)

	godotenv.Load()
	tr := &http.Transport{
		MaxIdleConns:       MaxIdleConn,
		IdleConnTimeout:    IdleConnTimeout * time.Second,
		DisableCompression: true,
	}
	client = langoClient{inner: &http.Client{Transport: tr}}

	client.fBToken = os.Getenv("FIREBASE_FUNC_TOKEN")
	client.fBTranslationAddress = os.Getenv("FIREBASE_FUNC_ADDRESS")
	client.fBUserEmail = os.Getenv("FIREBASE_USER_EMAIL_ADDRESS")
	srvAddressAndPort = os.Getenv("SERVER_ADDRESS_AND_PORT")

	atomicRWTranslations.mu = sync.RWMutex{}

	client.readTranslationsFromRemote()
}

func main() {
	app := iris.New()
  	app.Get("/get/{ln:string}", getLanguage)
	app.Post("/create", createTranslations)
  	app.Listen(srvAddressAndPort)
}
