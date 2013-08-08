package main

import (
	"code.google.com/p/go.net/html"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func FindChild(n *html.Node, predicate func(*html.Node) bool) *html.Node {
	if predicate(n) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := FindChild(c, predicate)
		if result != nil {
			return result
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func FindChildren(n *html.Node, predicate func(*html.Node) bool) list.List {
	results := list.List{}
	if predicate(n) {
		results.PushBack(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		inner_results := FindChildren(c, predicate)
		if inner_results.Len() > 0 {
			results.PushBackList(&inner_results)
		}
	}

	return results
}

func UrlToNode(baseUrl string) (*html.Node, error) {

	resp, err := http.Get(baseUrl)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body_string := string(body)
	doc, err := html.Parse(strings.NewReader(body_string))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func FetchImage(directory string) string {
	baseUrl := directory + "large/"

	//fmt.Println("Looking in", baseUrl)

	doc, err := UrlToNode(baseUrl)
	if err != nil {
		log.Fatal(err)
	}

	links := FindChildren(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					return true
				}
			}
		}
		return false
	})

	// shuffle links
	idx := 0
	randIdx := rand.Int() % links.Len()
	for link := links.Front(); link != nil; link = link.Next() {
		if idx == randIdx {
			linkNode := link.Value.(*html.Node)
			for _, a := range linkNode.Attr {
				if a.Key == "href" {
					return baseUrl + a.Val
				}
			}
		}
		idx++
	}

	return ""
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func GetRandomImages() []string {

	baseUrl := "http://deadendthrills.com/imagestore/DET3/"
	doc, err := UrlToNode(baseUrl)
	if err != nil {
		log.Fatal(err)
	}

	// parse	
	links := FindChildren(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					return true
				}
			}
		}
		return false
	})

	// shuffle links
	idx := 0
	linkSlice := make([]*html.Node, links.Len())
	rand.Seed(time.Now().UTC().UnixNano())
	randIdxs := rand.Perm(links.Len())
	for link := links.Front(); link != nil; link = link.Next() {
		linkSlice[randIdxs[idx]] = link.Value.(*html.Node)
		idx++
	}

	// get images
	i := int(math.Min(float64(len(linkSlice)-1), 10)) // number of images
	imgs := make([]string, i)
	for _, linkNode := range linkSlice {
		for _, a := range linkNode.Attr {
			if a.Key == "href" {

				if strings.Contains(a.Val, "imagestore") {
					continue
				}
				i--
				imgs[i] = FetchImage(baseUrl + a.Val)
				break
			}
		}
		if i <= 0 {
			break
		}
	}

	return imgs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func main() {

	args := os.Args
	port := 8080
	test := false

	for i := 0; i < len(args); i++ {
		if args[i] == "--test" {
			test = true
		}
		if args[i] == "--port" {
			i++
			nport, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Println("Invalid port")
				return
			}
			port = nport
		}
	}

	// test mode
	if test {
		imgs := GetRandomImages()
		fmt.Println(imgs)
		return
	}

	fmt.Println("Starting HTTP server at localhost:" + strconv.Itoa(port) + "...")

	http.Handle("/jsonp", http.HandlerFunc(func(c http.ResponseWriter, req *http.Request) {
		response := make(map[string]interface{})

		response["images"] = GetRandomImages()

		b, err := json.Marshal(response)
		if err != nil {
			c.Write([]byte("parseRequest({\"error\":\"An error occurred\"})"))
		} else {
			c.Write([]byte("parseRequest(" + string(b) + ")"))
		}
	}))
	http.Handle("/", http.HandlerFunc(func(c http.ResponseWriter, req *http.Request) {
		template.Must(template.ParseFiles("../client/index.html")).Execute(c, req.Host)
	}))
	err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), nil)
	if err != nil {
		fmt.Printf("ListenAndServe Error :" + err.Error())
	}
}
