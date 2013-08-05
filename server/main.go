package main

import(
	"fmt"
	"log"
	"text/template"
	"io/ioutil"
	"strings"
	"strconv"
	"net/http"
	"container/list"
	"code.google.com/p/go.net/html"
	"encoding/json"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func FindChildren(n *html.Node, predicate func(*html.Node) bool) list.List{
	results := list.List{}
	if predicate(n){
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func get_random() []string {
	resp, err := http.Get("http://deadendthrills.com/category/random/")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	   log.Fatal(err)
	}
	body_string := string(body)
	doc, err := html.Parse(strings.NewReader(body_string))
	if err != nil {
		log.Fatal(err)
	}
	
	//posts := FindNodesByClass(doc, "post")
	
	posts := FindChildren(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "div"{
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "content" {
					return true
				}
			}
		}
		return false
	})
	
	imgs := make([]string, posts.Len())
	i := 0
	for e := posts.Front(); e != nil; e = e.Next() {
		post := e.Value.(*html.Node)
		links := FindChildren(post, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "a"
	    })
		link := links.Front().Value.(*html.Node)
		
		for _, a := range link.Attr {
			if a.Key == "href" {
				imgs[i] = a.Val
			}
		}
		i++
    }
	
	return imgs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func main() {

    port := 8081
    fmt.Println("Starting HTTP server at localhost:" + strconv.Itoa(port) + "...")
	
	http.Handle("/jsonp", http.HandlerFunc(func (c http.ResponseWriter, req *http.Request) {
		response := make(map[string]interface{})
		
		response["images"] = get_random()
		
		b, err := json.Marshal(response)
		if err != nil {
			c.Write([]byte("parseRequest({\"error\":\"An error occurred\"})"))
		} else {
			c.Write([]byte("parseRequest("+string(b)+")"))
		}
	}))
	http.Handle("/", http.HandlerFunc(func (c http.ResponseWriter, req *http.Request) {
		template.Must(template.ParseFiles("../client/index.html")).Execute(c, req.Host)
	}))
	err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), nil)
	if err != nil {
		fmt.Printf("ListenAndServe Error :" + err.Error())
	}
}
