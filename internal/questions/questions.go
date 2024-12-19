package questions

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
)

type Question struct {
	Question string `json:"question"`
	A        string `json:"A"`
	B        string `json:"B"`
	C        string `json:"C"`
	D        string `json:"D"`
	Answer   string `json:"answer"`
}

func LoadQuestionsJSON() []Question {
	dat, err := os.Open("multipleAnswer.json")
	if err != nil {
		panic(err)
	}
	var questionArray []Question
	byteValue, _ := io.ReadAll(dat)
	errMarshal := json.Unmarshal(byteValue, &questionArray)
	if errMarshal != nil {
		fmt.Println("Erorare marshall")
	}
	return questionArray
}

func PickRandomQuestion(questionArray []Question) Question {
	n := len(questionArray)

	return questionArray[rand.IntN(n)]
}
