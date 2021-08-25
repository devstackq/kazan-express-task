package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Joke struct {
	Categories []string
	Id         string `json:"id"`
	Value      string `json:"value"`
}

//create files by category name
func createFiles(categories []string) {
	for _, v := range categories {
		f, err := os.OpenFile(strings.Trim(v, "\"")+`.txt`, os.O_CREATE|os.O_RDONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
	}
}

func getDataFromUrl(endpoint string, typeRequest string) *Joke {
	resp, err := http.Get(`https://api.chucknorris.io/jokes/` + endpoint)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	joke := Joke{}
	if typeRequest == "category" {
		json.Unmarshal(body, &joke.Categories)
	} else if typeRequest == "random" {
		json.Unmarshal(body, &joke)
	}
	return &joke
}

//create file - category name
func getCategories() []string {
	body := getDataFromUrl("categories", "category")
	createFiles(body.Categories)
	return body.Categories
}

//read arg cli
func readArg() int {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}
	text = strings.TrimSuffix(text, "\n")
	n, err := strconv.Atoi(text)

	if err != nil {
		log.Println(err)
	}
	return n
}

func readFile(filename string) []string {

	file, err := os.OpenFile(filename+".txt", os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	// append each joke, by category
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		// map[] := strings.Split(scanner.Text(), ":")
	}
	return lines
}

func main() {

	var wg sync.WaitGroup
	categories := getCategories()
	n := readArg()
	wg.Add(len(categories)) //add wait group -> for count each call goroutine

	//each category handle own goroutine
	for i := 0; i < len(categories); i++ {

		go func(category string) {
			var wg2 sync.WaitGroup
			wg2.Add(n) //wg2 count equal number cli

			file, err := os.OpenFile(category+".txt", os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				log.Println(err)
			}

			go func() {
				//get N joke, each category
				for j := 0; j < n; j++ { //9, travel, etc..

					randomJoke := getDataFromUrl("random?category="+category, "random")
					jokesFromFile := readFile(category)

					b := []byte(randomJoke.Id + ":" + randomJoke.Value + "\n")

					if len(jokesFromFile) > 0 {
						uniq := true
						//compare if joke exist current file
						for k := 0; k < len(jokesFromFile); k++ {
							jokeId := strings.Split(jokesFromFile[k], ":")
							if randomJoke.Id == jokeId[0] {
								uniq = false
							}
						}

						if uniq {
							if _, err := file.Write(b); err != nil {
								log.Println(err)
							}
							// else n += 1, continue search?
						}
					} else {
						//first joke
						if _, err := file.Write(b); err != nil {
							log.Println(err)
						}
					}
					wg2.Done()
				}
			}()
			wg2.Wait()

			defer file.Close()
			wg.Done()
		}(categories[i])
	}
	wg.Wait()
	log.Println("done")
}
