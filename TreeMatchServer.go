package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

// Question struct
type Question struct {
	ID         int      `json:"id" validate:"required"`
	Question   string   `json:"question" validate:"required"`
	Validation []string `json:"validation" validate:"required"`
}

// Questions struct
type Questions struct {
	Questions []Question `json:"questions"`
}

// Step struct
type Step struct {
	ID         int            `json:"id" validate:"required"`
	QuestionID int            `json:"question_id" validate:"required_without=ResultID"`
	Answers    map[string]int `json:"answers" validate:"required_without=ResultID"`
	ResultID   int            `json:"result_id" validate:"required_without_all=QuestionID Answers"`
}

// Steps struct
type Steps struct {
	Steps []Step `json:"steps"`
}

// Result struct
type Result struct {
	ID          int    `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

// Results struct
type Results struct {
	Results []Result `json:"results"`
}

// TemQuestion to display
type TemQuestion struct {
	StepID   int      `json:"step_id"`
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
}

// QuestionFormat struct
type QuestionFormat struct {
	Question TemQuestion `json:"question"`
}

// TemAnswer struct
type TemAnswer struct {
	StepID int    `json:"step_id" validate:"required"`
	Answer string `json:"answer" validate:"required"`
}

// Match struct
type Match struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// MatchFormat struct
type MatchFormat struct {
	MatchFormat Match `json:"match"`
}

// ErrMsg struct
type ErrMsg struct {
	ErrMsg string
}

// init
var questions Questions
var steps Steps
var results Results

var temQ QuestionFormat

func main() {

	// make database from JSON file ----------------------//

	// for validation
	validate := validator.New()

	// open, close file
	jsonFile, err := os.Open("questions.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("File opened.")
	defer jsonFile.Close()

	// read file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// get questions
	json.Unmarshal(byteValue, &questions)
	if len(questions.Questions) <= 0 {
		log.Fatal("There is no questions.")
	}
	for i := 0; i < len(questions.Questions)-1; i++ {
		err = validate.Struct(questions.Questions[i])
		if err != nil {
			s := "---- Suspect quesiont ID = " + strconv.Itoa(i+1) + " ----\n"
			log.Fatal(s + err.Error())
		}
	}

	// get steps
	json.Unmarshal(byteValue, &steps)
	if len(steps.Steps) <= 0 {
		log.Fatal("There is no steps.")
	}
	for i := 0; i < len(steps.Steps)-1; i++ {
		err = validate.Struct(steps.Steps[i])
		if err != nil {
			s := "---- Suspect step ID = " + strconv.Itoa(i+1) + " ----\n"
			log.Fatal(s + err.Error())
		}
	}

	// get results
	json.Unmarshal(byteValue, &results)
	if len(results.Results) <= 0 {
		log.Fatal("There is no results.")
	}
	for i := 0; i < len(results.Results)-1; i++ {
		err = validate.Struct(results.Results[i])
		if err != nil {
			s := "---- Suspect result ID = " + strconv.Itoa(i+1) + " ----\n"
			log.Fatal(s + err.Error())
		}
	}

	// make API -----------------------------------------//

	// initial Router
	r := mux.NewRouter()

	// Route Handles / Endpointers
	r.HandleFunc("/api/begin", begin).Methods("GET")
	r.HandleFunc("/api/answer", answer).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))

}

// get answer and return new question / match
// format:
// {
//		"step_id": 1,
//		"answer": "courtyard"
// }
func answer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get answer
	var answer TemAnswer
	_ = json.NewDecoder(r.Body).Decode(&answer)

	fmt.Println(answer.StepID)
	fmt.Println(answer)
	if answer.StepID <= 0 || answer.StepID > len(steps.Steps) {
		T := ErrMsg{ErrMsg: "Wrong step_id input, step_id can't be empty, only can be number, input again please."}
		json.NewEncoder(w).Encode(T)
		return
	}
	if len(answer.Answer) <= 0 {
		T := ErrMsg{ErrMsg: "Wrong answer input, answer can be only from previous answer, input again please."}
		json.NewEncoder(w).Encode(T)
		return
	}

	// find next question(or a tree match!)
	// get the step struct
	curStep := steps.Steps[answer.StepID-1]

	//var nextStepID int
	nextStepID := curStep.Answers[answer.Answer]
	if nextStepID <= 0 {
		T := ErrMsg{ErrMsg: "Wrong answer input, answer must be identical with one of previous answers choices, input again please."}
		json.NewEncoder(w).Encode(T)
		return
	}

	nextStep := steps.Steps[nextStepID-1]
	if nextStep.ResultID != 0 {
		m := Match{Name: results.Results[nextStep.ResultID-1].Name, Description: results.Results[nextStep.ResultID-1].Description}
		match := MatchFormat{MatchFormat: m}
		json.NewEncoder(w).Encode(match)
		return
	}

	nextQID := nextStep.QuestionID
	json.NewEncoder(w).Encode(questions.Questions[nextQID-1])
}

// begin with first step's question
func begin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	qID := steps.Steps[0].QuestionID

	q := TemQuestion{
		StepID:   1,
		Question: questions.Questions[qID-1].Question,
		Answers:  questions.Questions[qID-1].Validation,
	}

	temQ = QuestionFormat{Question: q}
	json.NewEncoder(w).Encode(temQ)
	return
}
