package proposal

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"

	"github.com/goverland-labs/inbox-storage/internal/metrics"
)

type AIClient struct {
	apiKey string
}

func NewAIClient(apiKey string) *AIClient {
	return &AIClient{
		apiKey: apiKey,
	}
}

func (c *AIClient) GetSummaryByProposalLink(ctx context.Context, link string) (string, error) {
	return c.do(ctx, fmt.Sprintf("I need the brief digest up to 70 words with important points for the %s", link))
}

func (c *AIClient) GetSummaryByDiscussionLink(ctx context.Context, link string) (string, error) {
	return c.do(ctx, fmt.Sprintf(`Your task is to make a brief but objective summarization. Highlight the most important parts, like whether funding is requested and whether the participants have a consensus in the comments. The summarization text should be brief—ideally within 100-200 words (or less, if possible). It should be in the most straightforward language possible (without losing meaning). Here is the link: %s`, link))
}

func (c *AIClient) GetSummaryByDescription(ctx context.Context, description string) (string, error) {
	return c.do(ctx, fmt.Sprintf(`Your task is to make a brief but objective summarization. Highlight the most important parts, like whether funding is requested and whether the participants have a consensus in the comments. The summarization text should be brief—ideally within 100-200 words (or less, if possible). It should be in the most straightforward language possible (without losing meaning). Below is the text of the proposal. 
%s`, description))
}

// do make request to the ChatGPT with provided string request
func (c *AIClient) do(ctx context.Context, req string) (string, error) {
	var err error
	defer func(start time.Time) {
		metrics.CollectRequestsMetric("open_ai", "create_chat_completion", err, start)
	}(time.Now())

	log.Info().Msg(fmt.Sprintf("Getting summary for: %s", req))

	client := openai.NewClient(c.apiKey)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai.CreateChatCompletion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai.CreateChatCompletion: no choices found")
	}

	return resp.Choices[0].Message.Content, nil
}
