package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Config vars (set from main.go or env)
var OpenAIKey string

// Request/response structures for OpenAI
type ChatRequest struct {
	Model    string     `json:"model"`
	Messages []Message  `json:"messages"`
}

type Message struct {
	Role    string     `json:"role"`
	Content []Content  `json:"content"`
}

type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Call OpenAI with an image URL and return LaTeX string
func GenerateLatexFromImage(imageURL string) (string, error) {
	if OpenAIKey == "" {
		OpenAIKey = os.Getenv("OPEN_AI_API_KEY")
	}
	if OpenAIKey == "" {
		return "", fmt.Errorf("OPEN_AI_API_KEY not set")
	}

	reqBody := ChatRequest{
		Model: "gpt-5-2025-08-07", // you can also try "gpt-4.1"
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{Type: "text", Text: `
Convert the image into **Obsidian-flavoured Markdown**.
Rules:
- Keep all prose as normal Markdown (headings, lists, bold/italic). Do NOT put prose inside LaTeX.
- Put ONLY mathematical notation in LaTeX.
- Inline math: wrap with $ like $a^2+b^2=c^2$.
- Multiline or multi-step equations/derivations: wrap with $$ on their own lines:
$$
(a + bi) + (c + di) = (a+c) + (b+d)i;
$$
- No code fences, no backticks, no explanations—**return the final Markdown only**.
- Try to preserve the original structure and numbering from the image (problems, parts (a)/(b), bullets).
- Normalize math: use \frac{…}{…}, \cdot, \sum, \int, \mathbb{R}, \vec{x}, \ldots, etc.
- If any symbol/word is unclear, insert [illegible] rather than guessing.
`},
 
					{Type: "image_url", ImageURL: &ImageURL{URL: imageURL}},
				},
			},
		},
	}

	data, _ := json.Marshal(reqBody)
	url := "https://api.openai.com/v1/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("new request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OpenAIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI error response: %s", string(body))
		return "", fmt.Errorf("openai failed: %s", resp.Status)
	}

	var cr ChatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	latex := cr.Choices[0].Message.Content
	log.Printf("LLM produced LaTeX: %s", latex)
	return latex, nil
}
