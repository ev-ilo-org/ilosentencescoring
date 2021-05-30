package main

import (
	//"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/k3a/html2text"
	"github.com/kennygrant/sanitize"
	"gopkg.in/jdkato/prose.v2"
)

const AUTOWEIGHT = true
const NO_AUTOWEIGHT = false

type ReferenceSentences struct {
	Reference string
	Text      string
}

type SentenceService struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
}

type SentenceServiceResponse struct {
	Similarty float64 `json:"similarty"`
}

const NLP_SERVER = "http://localhost:8083/sim"

type SpacyLemmatizerResult []struct {
	Label string `json:"label"`
}

type SpacyLemCall struct {
	Text  string `json:"text"`
	Model string `json:"model"`
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// Clean all HTML / Tags and newlines
func cleantext(tobecleaned string) string {

	var cleanedtext = ""
	cleanedtext = sanitize.Accents(tobecleaned)
	cleanedtext = sanitize.HTML(cleanedtext) // remove HTML Tags

	cleanedtext = html2text.HTML2Text(cleanedtext)

	cleanedtext = strings.Replace(cleanedtext, "\n", " ", -1)
	cleanedtext = strings.Replace(cleanedtext, "|", ". ", -1)
	cleanedtext = strings.Replace(cleanedtext, ".", " ", -1)
	cleanedtext = strings.Replace(cleanedtext, ",", " ", -1)

	//cleantest - spaces etc
	cleanedtext = standardizeSpaces(cleanedtext)
	cleanedtext = strings.Replace(cleanedtext, " - ", ". ", -1)
	cleanedtext = strings.Replace(cleanedtext, `"`, "", -1)

	//cleanedtext = stopwords.CleanString(cleanedtext, "en", true)
	cleanedtext = strings.ToLower(cleanedtext)

	// remove stop words
	cleanedtext = stopwords.CleanString(cleanedtext, "en", true)
	return cleanedtext
}

func SpacyLemmatizerSentence(rawword string) string {

	// url := "http://tika.eastvillagescl.com:8083/lem"
	url := "http://tika.eastvillagescl.com:8083/lem" // spacy server in Wales
	var callpayload SpacyLemCall
	var spacylem SpacyLemmatizerResult

	if len(rawword) == 0 {
		return rawword
	}
	callpayload.Model = "en_core_web_md" // "en_core_web_lg"

	callpayload.Text = rawword

	bodybytes, _ := json.Marshal(callpayload)
	payload := strings.NewReader(string(bodybytes))

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	err := json.Unmarshal(body, &spacylem)
	if err != nil {
		return rawword
	}
	// Iterate over the doc's tokens:
	var lemmatext string
	for _, Lemmatizedwords := range spacylem {
		lemmatext = lemmatext + " " + Lemmatizedwords.Label
		//return Lemmatizedwords.Label

	}

	return lemmatext

}

func process_pos(text string) string {

	var newsentence string
	doc, err := prose.NewDocument(text)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over the doc's tokens:
	for _, tok := range doc.Tokens() {
		fmt.Println(tok.Text, tok.Tag, tok.Label)

		// Go NNP B-GPE
		// is VBZ O
		// an DT O
		// ...
		if tok.Tag == "NN" || tok.Tag == "NNS" || tok.Tag == "VB" {
			newsentence = fmt.Sprintf("%v %v", newsentence, tok.Text)
		}
	}

	return newsentence
}

func sentence_similarity(reference string, text string, autoweight bool) float64 {

	url := NLP_SERVER

	var callpayload SentenceService
	var callresponse SentenceServiceResponse

	//callpayload.Reference = SpacyLemmatizerSentence(strings.ToLower(cleantext(reference)))
	//callpayload.Text = SpacyLemmatizerSentence(strings.ToLower(cleantext(text)))

	callpayload.Reference = strings.ToLower(cleantext(reference))
	callpayload.Text = strings.ToLower(cleantext(text))

	bodybytes, _ := json.Marshal(callpayload)
	payload := strings.NewReader(string(bodybytes))

	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return callresponse.Similarty
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return callresponse.Similarty
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return callresponse.Similarty
	}

	//fmt.Println("Input Text ", " ", inputtext)
	err_unmarhsal := json.Unmarshal(body, &callresponse)
	if err_unmarhsal != nil {
		fmt.Printf("Error = %v\n", err_unmarhsal)
	}
	/*
		//	fmt.Printf("Error = %v\n", err)

			fmt.Printf("Sentiment %v %v %v\n",
				callresponse.SentimentPolarity,
				callresponse.SentimentSubjectivity,
				callresponse.DetectedLanguage)
	*/

	if autoweight {
		weight := setintersection_members(reference, text) + 1 // +1 ensure that score remains even if no match direct intersection
		callresponse.Similarty = callresponse.Similarty * float64(weight)
	}

	return callresponse.Similarty
}

func read_references() []string {
	var refsentences []string
	csvfile, err := os.Open("reference_sentences.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)
	//r := csv.NewReader(bufio.NewReader(csvfile))
	var count int
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		count++
		refsentences = append(refsentences, record[1])
	}

	for _, sent := range refsentences {
		fmt.Printf("Sentence = %v\n", sent)
	}

	return refsentences

}

func process_sentences(filename string, references []string) {

	// Three maps for each col / save re-computing
	col0map := make(map[string]float64)
	col1map := make(map[string]float64)
	//col2map := make(map[string]float64)

	csvfile, err := os.Open(filename)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Write CSV Header
	var records [][]string
	var record []string

	record = append(record, "Line")
	record = append(record, "Seq")
	record = append(record, "Answer")
	for _, sent := range references {
		record = append(record, sent)
		fmt.Printf("Sentence = %v\n", sent)
	}
	records = append(records, record)

	// Parse the file
	r := csv.NewReader(csvfile)
	//r := csv.NewReader(bufio.NewReader(csvfile))
	var count int
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		count++

		// Loop through Each Reference Sentence -

		//
		col0_text := record[0]
		col1_text := record[1]
		col2_text := record[2]

		{
			var resultrecord []string
			countstr := fmt.Sprintf("%v", count)
			resultrecord = append(resultrecord, countstr)
			resultrecord = append(resultrecord, "1")
			resultrecord = append(resultrecord, col0_text)

			for _, sent := range references {
				mapkey := fmt.Sprintf("%v+%v", sent, col0_text)
				// check if in the map already
				var result float64
				cachedresult, found := col0map[mapkey]
				if found {
					result = cachedresult
					fmt.Printf("Result cached = %v / %v\n", mapkey, result)
				} else {
					result = sentence_similarity(sent, col0_text, NO_AUTOWEIGHT)
					col0map[mapkey] = result
					fmt.Printf("adding cached = %v / %v\n", mapkey, result)
				}

				resultstr := fmt.Sprintf("%.7f", result)
				fmt.Printf("%v Sent %s /  %v = %v\n", count, sent, col0_text, resultstr)
				resultrecord = append(resultrecord, resultstr)
			}
			records = append(records, resultrecord)
		}

		{
			var resultrecord []string
			countstr := fmt.Sprintf("%v", count)
			resultrecord = append(resultrecord, countstr)
			resultrecord = append(resultrecord, "2")
			resultrecord = append(resultrecord, col1_text)

			for _, sent := range references {
				mapkey := fmt.Sprintf("%v+%v", sent, col0_text)
				// check if in the map already
				var result float64
				cachedresult, found := col1map[mapkey]
				if found {
					result = cachedresult
					fmt.Printf("Result cached = %v / %v\n", mapkey, result)
				} else {
					result = sentence_similarity(sent, col1_text, NO_AUTOWEIGHT)
					col1map[mapkey] = result
					fmt.Printf("adding cached = %v / %v\n", mapkey, result)
				}
				resultstr := fmt.Sprintf("%.7f", result)
				fmt.Printf("%v Sent %s /  %v = %v\n", count, sent, col1_text, resultstr)
				resultrecord = append(resultrecord, resultstr)
			}
			records = append(records, resultrecord)
		}

		{ // 3rd Col
			var resultrecord []string
			countstr := fmt.Sprintf("%v", count)
			resultrecord = append(resultrecord, countstr)
			resultrecord = append(resultrecord, "3")
			resultrecord = append(resultrecord, col2_text)

			for _, sent := range references {
				result := sentence_similarity(sent, col2_text, NO_AUTOWEIGHT)
				resultstr := fmt.Sprintf("%.7f", result)
				fmt.Printf("%v Sent %s /  %v = %v\n", count, sent, col2_text, resultstr)
				resultrecord = append(resultrecord, resultstr)
			}
			records = append(records, resultrecord)
		}
		/*
			if count > 10 {
				break
			}*/
	}

	// Write Result to CSV
	f, err := os.Create("result.csv")
	defer f.Close()

	if err != nil {

		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

}

func setintersection_members(reference string, text string) (intersection_members int) {

	cleanreference := process_pos(SpacyLemmatizerSentence(cleantext(reference)))

	cleantext := process_pos(SpacyLemmatizerSentence(cleantext(text)))

	words := strings.Fields(cleanreference)

	fmt.Printf("Clean ref = %v\n", cleanreference)
	fmt.Printf("Clean txt = %v\n", cleantext)

	for _, word := range words {
		fmt.Printf("Word = %v Occurs = %v\n", word, strings.Count(cleantext, word))
		intersection_members = intersection_members + strings.Count(cleantext, word) // Check how many occurences

	}
	return intersection_members
}

func main() {
	// Open the file

	//fmt.Printf("Q = %v\n", process_pos(SpacyLemmatizerSentence("Monitor construction or the operations")))
	//return
	//process_sentences("input.csv", read_references())
	//process_sentences("fullinput.csv", read_references())
	process_sentences("DWA_cat.csv", read_references())

	return

	fmt.Printf("Interactions Members = %v\n", setintersection_members("strange", "world. the world is full of stranger worlds"))
	return

	fmt.Printf("A = %v\n", process_pos(SpacyLemmatizerSentence("position construction forms or molds")))
	fmt.Printf("Q = %v\n", process_pos(SpacyLemmatizerSentence("read form")))
	fmt.Printf("Q = %v\n", process_pos(SpacyLemmatizerSentence("operate or work with any heavy machines or industrial equipment machines equipment in factories construction sites warehouses repair shops or machine shops industrial kitchens some farming tractors harvesters milking machine")))
	return

}

/*
Clean
Tolower
Lemma

Process similarity = produce Score

Build two sets (from Question and Answer Sentences) - NOUN only
Check Intersection number

Update Score with (overlap Member * Score ) = amplified score



*/
