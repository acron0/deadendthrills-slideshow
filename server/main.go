package main

import(
	"os"
	"bytes"
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

func FindChild(n *html.Node, predicate func(*html.Node) bool) *html.Node{
	if predicate(n){
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
	resp, err := http.Get("http://deadendthrills.com/random/#")
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
	
	noscript := FindChild(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "noscript"{
			return true
		}
		return false
	})

	nodeData := bytes.NewBufferString(noscript.FirstChild.Data)
	node, perr := html.Parse(nodeData)
	if perr != nil {
		fmt.Println(perr.Error())
		return nil
	}

	links := FindChildren(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a"
	})
	
	imgs := make([]string, links.Len())
	i := 0
	for link := links.Front(); link != nil; link = link.Next() {		
		linkNode := link.Value.(*html.Node)
		for _, a := range linkNode.Attr {
			if a.Key == "href" {
				imgs[i] = a.Val
				break
			}
		}
		i++
    }
	
	return imgs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func main() {

	args := os.Args
    port := 8080
    test := false


    for i := 0; i < len(args); i++{
    	if(args[i] == "--test"){
    		test = true
    	}
    	if(args[i] == "--port"){
    		i++
    		nport, err := strconv.Atoi(args[i])
    		if(err != nil){
    			fmt.Println("Invalid port")
    			return
    		}
    		port = nport
    	}
    }

	// test mode
    if(test){
    	 imgs := get_random()
    	 fmt.Println(imgs)
    	 return
    }

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
