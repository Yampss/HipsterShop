package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultGeminiModel = "gemini-2.5-flash"
	geminiAPIBase      = "https://generativelanguage.googleapis.com/v1beta/models"
)

type assistantChatRequest struct {
	ProductID string `json:"productId"`
	Message   string `json:"message"`
}

type assistantChatResponse struct {
	Reply string `json:"reply"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiPromptFeedback struct {
	BlockReason string `json:"blockReason"`
}

type geminiAPIErrorBody struct {
	Error struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

type geminiResponse struct {
	Candidates     []geminiCandidate     `json:"candidates"`
	PromptFeedback *geminiPromptFeedback `json:"promptFeedback,omitempty"`
}

const maxAssistantJSONBody = 16 << 10 // 16 KiB

func (fe *frontendServer) assistantChatHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	if !assistantEnabled {
		http.Error(w, "assistant is disabled", http.StatusNotFound)
		return
	}

	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		http.Error(w, "assistant is unavailable", http.StatusServiceUnavailable)
		return
	}

	var req assistantChatRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, maxAssistantJSONBody))
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.ProductID = strings.TrimSpace(req.ProductID)
	req.Message = strings.TrimSpace(req.Message)
	if req.ProductID == "" {
		http.Error(w, "productId is required", http.StatusBadRequest)
		return
	}

	product, err := fe.getProduct(r.Context(), req.ProductID)
	if err != nil {
		log.WithField("product_id", req.ProductID).WithField("error", err).Warn("assistant failed to load product")
		http.Error(w, "could not load product", http.StatusBadGateway)
		return
	}

	reply, err := callGeminiAssistant(r.Context(), apiKey, os.Getenv("GEMINI_MODEL"), product, req.Message)
	if err != nil {
		log.WithField("product_id", req.ProductID).WithField("error", err).Warn("assistant call failed")
		http.Error(w, "assistant request failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(assistantChatResponse{Reply: reply})
}

func callGeminiAssistant(ctx context.Context, apiKey, model string, product *Product, userMessage string) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultGeminiModel
	}

	question := userMessage
	if question == "" {
		question = "Give me a concise 2-3 sentence description of this product, including what it is best for."
	}

	productJSON, err := json.MarshalIndent(product, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal product context: %w", err)
	}

	systemInstruction := "You are a shopping assistant for Hipster Shop. Use only the provided product JSON as source of truth. " +
		"If a detail is missing, clearly say you do not know. Keep responses brief, practical, and trustworthy."

	userPrompt := fmt.Sprintf("Product JSON:\n%s\n\nCustomer question:\n%s", string(productJSON), question)
	payload := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemInstruction}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userPrompt}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.4,
			MaxOutputTokens: 220,
		},
	}

	rawBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal gemini payload: %w", err)
	}

	u, err := url.Parse(fmt.Sprintf("%s/%s:generateContent", geminiAPIBase, url.PathEscape(model)))
	if err != nil {
		return "", fmt.Errorf("parse gemini url: %w", err)
	}
	q := u.Query()
	q.Set("key", apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(rawBody))
	if err != nil {
		return "", fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send gemini request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read gemini response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr geminiAPIErrorBody
		msg := ""
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
			msg = apiErr.Error.Message
		}
		if msg == "" {
			msg = strings.TrimSpace(string(body))
			if len(msg) > 280 {
				msg = msg[:280] + "…"
			}
		}
		return "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, msg)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse gemini response: %w", err)
	}
	if len(parsed.Candidates) == 0 {
		if parsed.PromptFeedback != nil && strings.TrimSpace(parsed.PromptFeedback.BlockReason) != "" {
			return "", fmt.Errorf("gemini blocked response: %s", parsed.PromptFeedback.BlockReason)
		}
		return "", fmt.Errorf("gemini returned no candidates")
	}
	var b strings.Builder
	for _, part := range parsed.Candidates[0].Content.Parts {
		b.WriteString(part.Text)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", fmt.Errorf("gemini returned empty text")
	}
	return out, nil
}
