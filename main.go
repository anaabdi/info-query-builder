package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

var (
	prodBaseURL, stgBaseURL *string
)

func main() {
	port := flag.String("port", "3222", "openend port for serving request")
	prodBaseURL = flag.String("prodURL", "http://localhost/images", "base url for accessing images in production")
	stgBaseURL = flag.String("stgURL", "http://localhost/images", "base url for accessing images in staging")
	flag.Parse()

	router := chi.NewRouter()

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I am coming from info query builder"))
	})

	router.Post("/query/generate", generatePromoQuery)
	router.Post("/query/update", updateQueryHandler)

	log.Printf("Starting to serve request to info query builder in port %s\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), router))
}

func parseReq(r *http.Request, body interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		return json.NewDecoder(r.Body).Decode(&body)
	}

	return errors.New("no supported type")
}

type updateQuery struct {
	CreatedAt    string `json:"created_at"`
	PrevImageURL string `json:"prev_image_url"`
	query
}

type query struct {
	InfoType  string   `json:"info_type"`
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	StartTime string   `json:"start_time"`
	EndDate   string   `json:"end_date"`
	Cities    []string `json:"cities"`
	PromoCode string   `json:"promocode"`
}

func (q query) imageBaseURL(env string) string {
	if env == "prod" {
		return fmt.Sprintf("%s/images/%s.png", *prodBaseURL, q.filename())
	}

	return fmt.Sprintf("%s/images/%s.png", *stgBaseURL, q.filename())
}

const (
	formatDateQuery = "2006-01-02 15:04:05.999"
	layoutDate      = "2006-01-02"
)

var (
	allCities = []string{"Bali", "Bandung", "Cilegon", "Jakarta", "Lombok", "Makassar", "Manado", "Medan", "Palembang", "Semarang", "Surabaya"}
	allEnvs   = []string{"stg", "prod"}

	queryInsert = `INSERT INTO infos (info_type, image_url, title, message, start_date, end_date, city, promocode, created_at, updated_at, deleted_at, is_active) 
	VALUES('%s', '%s', '%s', '%s', '%s', '%s',  '%s',  '%s', now(), now(), NULL, true);`

	queryUpdate = `UPDATE infos SET title = '%s', message = '%s', start_date = '%s', end_date = '%s', updated_at = now()) 
	WHERE promocode = '%s' and image_url = '%s' and created_at::date = '%s' ;`
)

func (q query) buildQuery(env, createdAt, prevImageURL string) {
	startDate, _ := time.Parse(layoutDate, q.StartTime)
	endDate, _ := time.Parse(layoutDate, q.EndDate)
	endDate = endDate.Add(86399 * time.Second)

	endDateStr := endDate.Format(formatDateQuery)
	startDateStr := startDate.Format(formatDateQuery)

	isBuildQueryUpdate := createdAt != ""

	if len(q.Cities) == 0 {
		q.Cities = allCities
	}

	now := time.Now()
	newlines := "\n\n"

	for _, env := range allEnvs {
		for _, city := range q.Cities {
			imageBaseURL := q.imageBaseURL(env)

			var stmt string

			if isBuildQueryUpdate {
				stmt = fmt.Sprintf(queryUpdate, q.Title, q.Message, startDateStr, endDateStr, q.PromoCode, imageBaseURL, createdAt)
			} else {
				stmt = fmt.Sprintf(queryInsert,
					q.InfoType, imageBaseURL, q.Title, q.Message, startDateStr, endDateStr, city, q.PromoCode)
			}

			filepath := filepath(now, env)
			fileName := fmt.Sprintf("%s.sql", q.filename())
			writeToFile(filepath, fileName, []byte(stmt+newlines))
		}
	}
}

func filepath(t time.Time, env string) string {
	return fmt.Sprintf("%s/%s/", t.Format("20060102"), env)
}

func (q query) filename() string {
	filename := q.PromoCode
	if q.PromoCode == "" {
		filename = strings.Replace(q.Title, " ", "", -1)
		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			fmt.Println(err)
		}

		filename = reg.ReplaceAllString(filename, "")
	}

	return filename
}

func updateQueryHandler(w http.ResponseWriter, r *http.Request) {
	env := getEnv(r)

	var q updateQuery
	if err := parseReq(r, &q); err != nil {
		log.Printf("err: %v", err)
		return
	}

	if cdate, _ := time.Parse(layoutDate, q.CreatedAt); cdate.IsZero() {
		log.Printf("err: %v", errors.New("invalid created_at"))
		return
	}

	if q.PrevImageURL == "" {
		log.Printf("err: %v", errors.New("missing prev image url"))
		return
	}

	fmt.Printf("updating query for \n %v \n", q)

	q.buildQuery(env, q.CreatedAt, q.PrevImageURL)
	w.WriteHeader(http.StatusCreated)
}

func getEnv(r *http.Request) string {
	env := r.URL.Query().Get("env")
	if env != "" {
		return env
	}
	return "stg"
}

func generatePromoQuery(w http.ResponseWriter, r *http.Request) {
	env := getEnv(r)

	var q query

	if err := parseReq(r, &q); err != nil {
		log.Printf("err: %v", err)
		return
	}

	fmt.Printf("generating query for \n %v \n", q)

	q.buildQuery(env, "", "")
	w.WriteHeader(http.StatusCreated)
}

func writeToFile(filepath, filename string, data []byte) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}

	file := filepath + filename

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}
}
