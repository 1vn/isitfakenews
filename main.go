package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/mat"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type OnlineVec struct {
	Vector *mat.Vector
	Fake   bool
}

var (
	UPDATE_SAMPLE_SIZE = 5
	intercept          = -1.06837132
	templatePaths      = []string{
		"index.html",
	}
	templates *template.Template
	WordBank  map[string]*mat.Vector

	blacklistSet = []string{
		"www.thebeaverton.com", "www.americannews.com", "www.bigamericannews.com", "www.christwire.org", "www.civictribune.com", "www.clickhole.com", "www.creambmp.com", "www.dcgazette.com", "www.dailycurrant.com", "www.dcclothesline.com", "www.derfmagazine.com", "www.drudgereport.com.co", "www.duhprogressive.com", "www.empirenews.com", "www.enduringvision.com", "www.msnbc.co", "www.msnbc.website", "www.mediamass.net", "www.nationalreport.net", "www.newsbiscuit.com", "www.news-hound.com", "www.newsmutiny.com", "www.politicalears.com", "www.private-eye.co.uk", "www.realnewsrightnow.com", "www.rilenews.com", "www.sprotspickle.com", "www.thenewsnerd.com", "www.theuspatriot.com", "www.witscience.org", "www.theonion.com", "www.amplifyingglass.com", "www.duffleblog.com", "www.empiresports.co", "www.gomerblog.com", "www.huzlers.com", "www.itaglive.com", "www.newslo.com", "www.nahadaily.com", "www.rockcitytimes.com", "www.thelapine.ca", "www.thespoof.com", "www.weeklyworldnews.com", "www.worldnewsdailyreport.com", "www.21stcenturywire.com", "www.activistpost.com", "www.beforeitsnews.com", "www.bigpzone.com", "www.chronicle.su", "www.coasttocoastam.com", "www.consciouslifenews.com", "www.conservativeoutfitters.com", "www.countdowntozerotime.com", "www.counterpsyops.com", "www.dailybuzzlive.com", "www.disclose.tv", "www.fprnradio.com", "www.geoengineeringwatch.org", "www.globalresearch.ca", "www.govtslaves.info", "www.gulagbound.com", "www.jonesreport.com", "www.hangthebankers.com", "www.humansarefree.com", "www.infowars.com", "www.intellihub.com", "www.lewrockwell.com", "www.libertytalk.fm", "www.libertyvideos.org", "www.megynkelly.us", "www.naturalnews.com", "www.newswire-24.com", "www.nodisinfo.com", "www.nowtheendbegins.com", "www.pakalertpress.com", "www.politicalblindspot.com", "www.prisonplanet.com", "www.prisonplanet.tv", "www.realfarmacy.com", "www.redflagnews.com", "www.truthfrequencyradio.com", "www.thedailysheeple.com", "www.therundownlive.com", "www.unconfirmedsources.com", "www.veteranstoday.com", "www.wakingupwisconsin.com", "www.worldtruth.tv",
	}

	model = []float64{
		-0.368792056532667, 0.5617918799819819, -0.8969301616906722, 0.615050743057359, 0.021108663784505582, 0.32901341450560834, -0.972283536803463, 0.21897015968450517, 0.5304807137037654, 0.0224418724399407, 0.9666131708689365, 0.5487397413234459, -0.2655609115069282, 0.16102965897440627, -0.2114920997975223, 0.036386778917172416, -0.16986404381406836, 0.29868152743115123, -0.7307668625289594, -0.3733521831709352, 0.6553497173019711, -0.4756550771767227, -0.6622093034239738, 0.5662172187635592, -1.1104679688139603, 1.379177950084947, 0.6484418125267746, -0.42206803769583395, -0.15065507228734834, 0.040955042175854, -0.08859817073559602, 0.16466860027422697, 0.7361759665144547, 0.903249887055352, 1.1338205580264482, 0.14668452719599281, 0.6142722670120557, 0.47725961287903346, 0.9699850721658022, -0.00583461748861685, 0.2858714288589395, -0.6228942481045906, -0.5051683821645134, 0.1701091805423817, 0.07643289649563592, 0.5287422999463888, 0.40564931450988123, -0.41551099331410557, -0.28746131294393173, 0.3942938233039454, -0.6986426039654011,
	}
	modelVec  *mat.Vector
	linkCache map[string]*mat.Vector

	sampleVectors []*OnlineVec
)

func initWordBank() {
	WordBank = make(map[string]*mat.Vector)
	f, err := os.Open("wordbank.txt")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		floatArray := make([]float64, 0)
		for _, val := range words[1:] {
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				panic(err)
			}
			floatArray = append(floatArray, floatVal)
		}
		floatArray = append(floatArray, 0)
		WordBank[words[0]] = mat.NewVector(51, floatArray)
	}

}

func initSamples() {
	sampleVectors = make([]*OnlineVec, 0)
	f, err := os.Open("samples.txt")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		floatArray := make([]float64, 0)
		for _, val := range words[:51] {
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				panic(err)
			}
			floatArray = append(floatArray, floatVal)
		}

		var fake bool
		switch words[51] {
		case "0":
			fake = true
		case "1":
			fake = false
		}

		sampleVectors = append(sampleVectors, &OnlineVec{
			Fake:   fake,
			Vector: mat.NewVector(51, floatArray),
		})
	}

}

func sigmoid(in float64) float64 {
	val := 1 / (1 + math.Exp(-in))
	return val
}

func initModel() {
	modelVec = mat.NewVector(51, model)
}

func initTmpl() {
	templates = template.Must(template.ParseFiles(templatePaths...))
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, renderArgs map[string]interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, renderArgs)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getStory(url string) string {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Print(err)
		return ""
	}

	//find most "p"
	largestLength := 0
	largestIndex := 0
	doc.Find("body").Children().Each(func(i int, s *goquery.Selection) {
		length := s.Find("p").Length()
		if length > largestLength {
			largestIndex = i
			largestLength = length
		}
	})

	var articleBuffer bytes.Buffer

	doc.Find("body").Children().Each(func(i int, s *goquery.Selection) {
		if i == largestIndex {
			s.Find("p").Each(func(i int, s *goquery.Selection) {
				articleBuffer.WriteString(fmt.Sprintf("%s ", s.Text()))
			})
		}
	})

	return articleBuffer.String()
}

func vectorizeStory(story string) *mat.Vector {
	storyWords := strings.Split(story, " ")

	underlying := make([]float64, 51)
	sum := mat.NewVector(51, underlying)
	for _, word := range storyWords {
		trimWord := strings.Trim(word, `;:,./)(*&^%$#@?<>!1234567890\\|"{}[]`)
		cleanWord := strings.ToLower(trimWord)
		if vec, ok := WordBank[cleanWord]; ok {
			sum.AddVec(sum, vec)
		} else {
			notFound := make([]float64, 51)
			notFound[50] = 1
			notFoundVec := mat.NewVector(51, notFound)
			sum.AddVec(sum, notFoundVec)
		}
	}

	for idx, val := range underlying {
		underlying[idx] = val / float64(len(storyWords))
	}

	return sum
}

func inferNews(url string) bool {
	story := getStory(url)
	var storyVec *mat.Vector
	if vec, ok := linkCache[url]; ok {
		storyVec = vec
	} else {
		storyVec = vectorizeStory(story)
		linkCache[url] = storyVec
	}

	sig := predictVector(storyVec)
	predict := sig < 0.5

	log.Println(sig)
	return predict
}

func predictVector(vec *mat.Vector) float64 {
	total := float64(0)
	for idx, val := range modelVec.RawVector().Data {
		total += val + vec.RawVector().Data[idx]
	}

	total += intercept

	sig := sigmoid(total)
	return sig
}

func Infer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseMap := make(map[string]interface{})
	urlStr := r.FormValue("url")

	urlStruct, err := url.Parse(urlStr)
	if err != nil {
		responseMap["result"] = true
		json.NewEncoder(w).Encode(responseMap)
		return
	}

	for _, black := range blacklistSet {
		if black == urlStruct.Host {
			responseMap["result"] = true
			json.NewEncoder(w).Encode(responseMap)
			return
		}
	}

	var inference bool
	inference = inferNews(urlStruct.String())

	responseMap["result"] = inference

	json.NewEncoder(w).Encode(responseMap)
}

func Index(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index.html", nil)
}

func Correct(w http.ResponseWriter, r *http.Request) {
	urlStr := r.PostFormValue("url")
	correctStr := r.PostFormValue("correct")
	correct, err := strconv.ParseBool(correctStr)
	if err != nil {
		log.Println(err)
		return
	}

	if _, ok := linkCache[urlStr]; !ok {
		return
	}

	correctVector := &OnlineVec{
		Vector: linkCache[urlStr],
		Fake:   inferNews(urlStr),
	}

	if !correct {
		correctVector.Fake = !correctVector.Fake
	}

	randomSamples := make([]*OnlineVec, 0)
	for i := 1; i <= UPDATE_SAMPLE_SIZE; i++ {
		randIndex := rand.Intn(len(sampleVectors))
		sampleVec := sampleVectors[randIndex]
		randomSamples = append(randomSamples, sampleVec)
	}

	randomSamples = append(randomSamples, correctVector)
	sampleVectors = append(sampleVectors, correctVector)

	temp := make([]float64, 52)
	alpha := float64(0.5)
	for idx, _ := range temp {
		if idx == 51 {
			break
		}

		oldCoeff := model[idx]

		gradSum := float64(0)
		for _, sample := range randomSamples {
			pred := predictVector(sample.Vector)
			label := float64(0)
			if !sample.Fake {
				label = 1
			}

			diff := pred - label

			xj := sample.Vector.RawVector().Data[idx]

			diff *= xj

			gradSum += diff
		}

		temp[idx] = oldCoeff - alpha*gradSum
	}

	//deal with 52nd feature
	gradSum := float64(0)
	for _, sample := range randomSamples {
		pred := predictVector(sample.Vector)
		label := float64(0)
		if !sample.Fake {
			label = 1
		}

		diff := pred - label

		xj := float64(1)

		diff *= xj

		gradSum += diff
	}

	temp[51] = intercept - alpha*gradSum
	for idx, _ := range model {
		model[idx] = temp[idx]
	}

	intercept = temp[51]

}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	initTmpl()
	initWordBank()
	initModel()
	initSamples()
	linkCache = make(map[string]*mat.Vector)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", Index)
	r.HandleFunc("/infer", Infer)
	r.HandleFunc("/correct", Correct)

	s := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))
	r.PathPrefix("/assets/").Handler(s)

	http.ListenAndServe(":5555", r)

}
