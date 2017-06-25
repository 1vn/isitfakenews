package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/mat"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

var (
	templatePaths = []string{
		"index.html",
	}
	templates *template.Template
	WordBank  map[string]*mat.Vector

	blacklistSet = []string{
		"www.americannews.com", "www.bigamericannews.com", "www.christwire.org", "www.civictribune.com", "www.clickhole.com", "www.creambmp.com", "www.dcgazette.com", "www.dailycurrant.com", "www.dcclothesline.com", "www.derfmagazine.com", "www.drudgereport.com.co", "www.duhprogressive.com", "www.empirenews.com", "www.enduringvision.com", "www.msnbc.co", "www.msnbc.website", "www.mediamass.net", "www.nationalreport.net", "www.newsbiscuit.com", "www.news-hound.com", "www.newsmutiny.com", "www.politicalears.com", "www.private-eye.co.uk", "www.realnewsrightnow.com", "www.rilenews.com", "www.sprotspickle.com", "www.thenewsnerd.com", "www.theuspatriot.com", "www.witscience.org", "www.theonion.com", "www.amplifyingglass.com", "www.duffleblog.com", "www.empiresports.co", "www.gomerblog.com", "www.huzlers.com", "www.itaglive.com", "www.newslo.com", "www.nahadaily.com", "www.rockcitytimes.com", "www.thelapine.ca", "www.thespoof.com", "www.weeklyworldnews.com", "www.worldnewsdailyreport.com", "www.21stcenturywire.com", "www.activistpost.com", "www.beforeitsnews.com", "www.bigpzone.com", "www.chronicle.su", "www.coasttocoastam.com", "www.consciouslifenews.com", "www.conservativeoutfitters.com", "www.countdowntozerotime.com", "www.counterpsyops.com", "www.dailybuzzlive.com", "www.disclose.tv", "www.fprnradio.com", "www.geoengineeringwatch.org", "www.globalresearch.ca", "www.govtslaves.info", "www.gulagbound.com", "www.jonesreport.com", "www.hangthebankers.com", "www.humansarefree.com", "www.infowars.com", "www.intellihub.com", "www.lewrockwell.com", "www.libertytalk.fm", "www.libertyvideos.org", "www.megynkelly.us", "www.naturalnews.com", "www.newswire-24.com", "www.nodisinfo.com", "www.nowtheendbegins.com", "www.pakalertpress.com", "www.politicalblindspot.com", "www.prisonplanet.com", "www.prisonplanet.tv", "www.realfarmacy.com", "www.redflagnews.com", "www.truthfrequencyradio.com", "www.thedailysheeple.com", "www.therundownlive.com", "www.unconfirmedsources.com", "www.veteranstoday.com", "www.wakingupwisconsin.com", "www.worldtruth.tv",
	}

	model     = []float64{-1.2737482873423125, 0.5142605867357818, 0.03370879411437541, 1.128443861590454, 1.1358368525010816, 0.35720347428853566, -1.1961518743190735, 0.2322227248512921, -0.4210460534911449, -0.03376353450943456, -1.3867871194847803, -0.8787988804117033, -0.8830888463820143, -0.6836304363488519, 0.2373738518714629, -0.6095781790421557, -1.5451400225785885, -1.3579258923320978, 0.08380092445168075, -0.3004120938915199, 1.0112704286991592, 1.610971235757192, -0.5639771214339668, 0.5832278847971472, 1.149164820993785, -1.0852161539572005, 0.9473528032342516, 0.2760889497210596, -0.34378759609904785, -1.2744406378919138, -1.3498792064156049, 0.221988954552692, -0.4693493678622002, -1.0411901706035622, -0.5776024532774324, -1.7042797044375582, -0.14961636750328172, -0.625535064861366, -0.7666057751699099, -0.21154871657483376, 0.046044142838342995, -0.37012732994482095, -0.7387381503691399, -0.1540698344108049, -0.7523041503763994, 0.25642665668049963, 0.10152643899105757, 0.2810367199385284, 1.456686148736361, -0.1313619368253374, -1.2845432190223274}
	modelVec  *mat.Vector
	linkCache map[string]*mat.Vector
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

	return false
}

func InferNewsHandler(w http.ResponseWriter, r *http.Request) {
	responseMap := make(map[string]interface{})
	urlStr := r.FormValue("url")

	urlStruct, err := url.Parse(urlStr)
	if err != nil {
		responseMap["result"] = true
		json.NewEncoder(w).Encode(responseMap)
	}

	for _, black := range blacklistSet {
		if black == urlStruct.Host {
			responseMap["result"] = true
			json.NewEncoder(w).Encode(responseMap)
		}
	}

	responseMap["result"] = inferNews(urlStruct.String())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseMap)
}

func Index(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index.html", nil)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	initTmpl()
	initWordBank()
	initModel()
	linkCache = make(map[string]*mat.Vector)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", Index)
	r.HandleFunc("/infer", InferNewsHandler)

	s := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))
	r.PathPrefix("/assets/").Handler(s)

	http.ListenAndServe(":5555", r)

}
