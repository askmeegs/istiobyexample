package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const timeLayout = "2006-01-02T15:04:05.000Z"

type Article struct {
	Id              int32     `json:"id"`
	Title           string    `json:"title"`
	PublicationDate time.Time `json:"publication_date"`
	Byline          string    `json:"byline"`
	ArticleText     string    `json:"article_text"`
}

func main() {
	http.HandleFunc("/", BaseHandler)
	http.HandleFunc("/article/breaking-news/2020/astrophysics-discovery", GetBreakingNewsHandler)
	http.HandleFunc("/beta/blog/2020/new-engineering-blog", GetBlogHandler)
	http.HandleFunc("/article/1962/usa-chestnut-tree-blight", GetArticleHandler)
	http.ListenAndServe(":8080", nil)
}

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📰 This is the articles service. Sleeping:")
	time.Sleep(time.Second * 15)
	a := &Article{}
	j, _ := json.Marshal(a)
	w.Write(j)
}

func GetArticleHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📆 received article request: %s")
	str := "1965-11-12T11:45:26.371Z"

	t, err := time.Parse(timeLayout, str)
	if err != nil {
		log.Println(err)
	}
	article := &Article{
		Id:              12345,
		Title:           "The Chestnut Blight: How the U.S. Lost 4 Billion Beloved Trees",
		PublicationDate: t,
		Byline:          "Abigail Smith",
		ArticleText:     "The American chestnut (Castanea dentata) is a large, monoecious deciduous tree of the beech family native to eastern North America.[2] Before the species was devastated by the chestnut blight, a fungal disease, it was one of the most important forest trees throughout its range, and was considered the finest chestnut tree in the world.[3] It is estimated that between 3 and 4 billion American chestnut trees were destroyed in the first half of the 20th century by blight after its initial discovery in 1904.[4][5][6] Very few mature specimens of the tree exist within its historical range, although many small shoots of the former live trees remain. [Source: Wikipedia.com]",
	}
	a, err := json.Marshal(article)
	if err != nil {
		log.Println(err)
	}
	w.Write(a)
}

func GetBreakingNewsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔥 received breaking news request: %s")
	str := "2020-01-14T11:45:26.371Z"

	t, err := time.Parse(timeLayout, str)
	if err != nil {
		log.Println(err)
	}
	article := &Article{
		Id:              67890,
		Title:           "Wow! New Habitable Planet Discovered 9.5 Light Years Away",
		PublicationDate: t,
		Byline:          "Paul Rockport",
		ArticleText:     "The planet is large. Its atmosphere is high in nitrogen. It is twice the size of Earth.",
	}
	a, err := json.Marshal(article)
	if err != nil {
		log.Println(err)
	}
	w.Write(a)
}

func GetBlogHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🎆 received blog request: %s")
	str := "2020-01-07T11:45:26.371Z"

	t, err := time.Parse(timeLayout, str)
	if err != nil {
		log.Println(err)
	}
	article := &Article{
		Id:              91385,
		Title:           "Welcome to the new GothamNews Blog!",
		PublicationDate: t,
		Byline:          "The Engineering Team",
		ArticleText:     "This is our new blog.",
	}
	a, err := json.Marshal(article)
	if err != nil {
		log.Println(err)
	}
	w.Write(a)
}
