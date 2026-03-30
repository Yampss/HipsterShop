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

	"github.com/GoogleCloudPlatform/microservices-demo/src/frontend/money"
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

type assistantContextProduct struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       string                 `json:"price"`
	Categories  []string               `json:"categories,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

type assistantContextPack struct {
	PrimaryProduct  assistantContextProduct   `json:"primaryProduct"`
	RelatedProducts []assistantContextProduct `json:"relatedProducts,omitempty"`
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

	relatedProducts, recErr := fe.getRecommendations(r.Context(), sessionID(r), []string{req.ProductID})
	if recErr != nil {
		log.WithField("product_id", req.ProductID).WithField("error", recErr).Debug("assistant recommendations unavailable")
		relatedProducts = nil
	}

	reply, err := callGeminiAssistant(r.Context(), apiKey, os.Getenv("GEMINI_MODEL"), product, relatedProducts, req.Message)
	if err != nil {
		log.WithField("product_id", req.ProductID).WithField("error", err).Warn("assistant call failed")
		http.Error(w, "assistant request failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(assistantChatResponse{Reply: reply})
}

func callGeminiAssistant(ctx context.Context, apiKey, model string, product *Product, relatedProducts []*Product, userMessage string) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultGeminiModel
	}

	question := userMessage
	if question == "" {
		question = "Give me a practical quick-buy summary of this product and who it is best for."
	}

	contextPack := buildAssistantContextPack(product, relatedProducts)
	contextJSON, err := json.MarshalIndent(contextPack, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal assistant context: %w", err)
	}

	systemInstruction := "You are a friendly, knowledgeable shopping assistant for Hipster Shop. " +
		"You are given a context JSON with product details including materials, colors, sizing, care, and more. " +
		"Use the context JSON as your primary source of truth. " +
		"You may also use reasonable general product knowledge to give helpful answers when the context provides enough context to infer it — for example, if a material is listed, you can comment on its feel or durability. " +
		"Answer questions about fit, use case, climate suitability, gift suitability, comparisons to related products, and care — being genuinely helpful. " +
		"If something is truly unknown (e.g. a color not listed), say what IS available and offer to help with something else. " +
		"Do not make up specifications like exact wattage, dimensions, or weight unless they are in the context JSON. " +
		"Keep responses concise, warm, and conversational. Plain text only — no markdown, no bullets, no bold markers."

	userPrompt := fmt.Sprintf("Context JSON:\n%s\n\nCustomer question:\n%s", string(contextJSON), question)
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
			Temperature:     0.5,
			MaxOutputTokens: 350,
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
	return normalizeAssistantReply(out), nil
}

func buildAssistantContextPack(product *Product, relatedProducts []*Product) assistantContextPack {
	pack := assistantContextPack{PrimaryProduct: toAssistantContextProduct(product)}
	if len(relatedProducts) > 0 {
		pack.RelatedProducts = make([]assistantContextProduct, 0, len(relatedProducts))
		for _, rp := range relatedProducts {
			if rp == nil || (product != nil && rp.Id == product.Id) {
				continue
			}
			pack.RelatedProducts = append(pack.RelatedProducts, toAssistantContextProduct(rp))
			if len(pack.RelatedProducts) == 3 {
				break
			}
		}
	}
	return pack
}

func toAssistantContextProduct(p *Product) assistantContextProduct {
	if p == nil {
		return assistantContextProduct{}
	}
	return assistantContextProduct{
		ID:          p.Id,
		Name:        p.Name,
		Description: p.Description,
		Price:       formatProductPrice(p.PriceUsd),
		Categories:  p.Categories,
		Details:     p.Details,
	}
}

func formatProductPrice(m *money.Money) string {
	if m == nil {
		return "unknown"
	}
	amount := float64(m.Units) + float64(m.Nanos)/1e9
	return fmt.Sprintf("%s %.2f", strings.TrimSpace(m.CurrencyCode), amount)
}

func normalizeAssistantReply(reply string) string {
	reply = strings.ReplaceAll(reply, "\r\n", "\n")
	reply = strings.ReplaceAll(reply, "**", "")
	reply = strings.ReplaceAll(reply, "__", "")
	reply = strings.ReplaceAll(reply, "`", "")
	lines := strings.Split(reply, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			if blank {
				continue
			}
			blank = true
			out = append(out, "")
			continue
		}
		blank = false
		out = append(out, line)
	}
	clean := strings.TrimSpace(strings.Join(out, "\n"))
	if len(clean) > 1100 {
		clean = strings.TrimSpace(clean[:1100]) + "..."
	}
	return clean
}
