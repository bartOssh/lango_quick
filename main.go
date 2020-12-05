package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Translations to popular languages
type Translations struct {
	Ru string `json:"ru"`
	En string `json:"en"`
	Pl string `json:"pl"`
	Es string `json:"es"`
	Fr string `json:"fr"`
	De string `json:"de"`
}

// CachedTranslation relation input to translations
type CachedTranslation struct {
	Input      string       `json:"input"`
	Translated Translations `json:"translated"`
}

var atomicRWTranslations struct {
	mu sync.RWMutex
	ct *[]CachedTranslation
}

var languageMap sync.Map

func getByInput(input string) *CachedTranslation {
	if t, ok := languageMap.Load(input); ok {
		return &CachedTranslation{
			Input:      input,
			Translated: t.(Translations),
		}
	}
	return nil
}

func sendTranslations(w http.ResponseWriter, r *http.Request) {
	atomicRWTranslations.mu.RLock()
	defer atomicRWTranslations.mu.RUnlock()
	if lang, ok := r.URL.Query()["input"]; ok {
		ct := getByInput(lang[0])
		json.NewEncoder(w).Encode(ct)
		return
	}
	json.NewEncoder(w).Encode(atomicRWTranslations.ct)
}

func main() {
	srvAddressAndPort := os.Getenv("SERVER_ADDRESS_AND_PORT")
	router := mux.NewRouter()
	router.HandleFunc("/translations", sendTranslations).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         srvAddressAndPort,
		WriteTimeout: WriteTimeoutS * time.Second,
		ReadTimeout:  ReadTimeoutS * time.Second,
	}

	go func() {
		log.Printf("server started on %s", srvAddressAndPort)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

	wait := time.Duration(ShoutDownTimeoutS) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down server")
	os.Exit(0)
}

func init() {
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
	rawBody := bytes.NewBuffer(postBody)

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(fBTranslationAddress, "application/json", rawBody)
	if err != nil {
		log.Fatalf("cannot fetch initial translation value, error: %s", err)
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("cannot read initial response from firebase translation end point %s", err)
	}
	ct := new([]CachedTranslation)
	err = json.Unmarshal(result, ct)
	if err != nil {
		log.Printf("error decoding response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("response: %q", result)
		log.Fatalf("cennot unmarshal firebase translation result %s", err)
	}
	atomicRWTranslations.mu = sync.RWMutex{}
	atomicRWTranslations.mu.Lock()
	defer atomicRWTranslations.mu.Unlock()
	atomicRWTranslations.ct = ct
	for _, t := range *ct {
		languageMap.Store(t.Input, t.Translated)
	}
	log.Printf("properly initialized lango quick microservice with translations map of size %v translations", len(*atomicRWTranslations.ct))

}
