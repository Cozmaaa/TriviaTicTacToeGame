package api

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func OpenAIAPICall(correctAnswer, fullCorrectAnswer, userAnswer string, ch chan string) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	promptAndAnswer := fmt.Sprintf(`You will have to answer only with : true or false ,
                if the answer provided is the same as the users message you will say true 
                , else you will say false , the answer can be in any other languages as well, even korean . The user can write only the correct letter such as a,b,c,d ,
        but he can also only write the answer , like john , pacific ocean , both should be accepted
                : The answer is : %s. %s `, correctAnswer, fullCorrectAnswer)

	fmt.Println(promptAndAnswer)
	fmt.Println(userAnswer)

	OPENAI_API_KEY := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(option.WithAPIKey(OPENAI_API_KEY))
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(promptAndAnswer),
			openai.UserMessage(userAnswer),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println((chatCompletion.Choices[0].Message.Content))

	ch <- (chatCompletion.Choices[0].Message.Content)
}
