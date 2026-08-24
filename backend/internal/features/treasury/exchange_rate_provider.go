package treasury

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// LiveRateTimeout bounds every outbound HTTP call so a down or
	// slow external API can never hang the desktop UI.
	LiveRateTimeout = 4 * time.Second

	// LiveRateBudget bounds the whole multi-provider resolution.
	LiveRateBudget = 10 * time.Second

	// FallbackUSDPEN is returned when every source fails, so forms
	// can still render and the user keeps a sane editable default.
	FallbackUSDPEN = "3.75"

	sunatAPIURL  = "https://api.apis.net.pe/v1/tipo-cambio-sunat?currency=pen"
	erAPIURLTmpl = "https://open.er-api.com/v6/latest/%s"
)

// Rate sources recorded alongside persisted snapshots.
const (
	SourceSunatAPI = "apis.net.pe"
	SourceERAPI    = "open.er-api.com"
	SourceFallback = "fallback"
)

// liveRateProvider fetches reference exchange rates from public APIs.
type liveRateProvider struct {
	client *http.Client
}

func newLiveRateProvider() *liveRateProvider {
	return &liveRateProvider{client: &http.Client{Timeout: LiveRateTimeout}}
}

// Fetch resolves from->to against the configured providers in order
// and returns (rate, source). It fails only when every applicable
// source fails; callers are expected to degrade gracefully.
func (p *liveRateProvider) Fetch(ctx context.Context, from, to string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, LiveRateBudget)
	defer cancel()

	fromU, toU := strings.ToUpper(from), strings.ToUpper(to)
	var lastErr error

	if fromU == "USD" && toU == "PEN" {
		rate, err := p.fetchSunat(ctx)
		if err == nil {
			return rate, SourceSunatAPI, nil
		}
		lastErr = err
	}

	rate, err := p.fetchOpenERAPI(ctx, fromU, toU)
	if err == nil {
		return rate, SourceERAPI, nil
	}
	if lastErr == nil {
		lastErr = err
	}
	return "", "", lastErr
}

// fetchSunat queries the SUNAT daily rate (venta) via apis.net.pe.
func (p *liveRateProvider) fetchSunat(ctx context.Context) (string, error) {
	body, err := p.getJSON(ctx, sunatAPIURL)
	if err != nil {
		return "", err
	}
	var payload struct {
		Venta json.RawMessage `json:"venta"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	raw := strings.Trim(string(payload.Venta), `"`)
	if raw == "" || raw == "null" {
		return "", fmt.Errorf("sunat api: empty venta")
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 {
		return "", fmt.Errorf("sunat api: invalid venta %q", raw)
	}
	return raw, nil
}

// fetchOpenERAPI queries open.er-api.com, which supports any base
// currency and therefore acts as the generic fallback source.
func (p *liveRateProvider) fetchOpenERAPI(ctx context.Context, from, to string) (string, error) {
	target := fmt.Sprintf(erAPIURLTmpl, url.PathEscape(from))
	body, err := p.getJSON(ctx, target)
	if err != nil {
		return "", err
	}
	var payload struct {
		Result string             `json:"result"`
		Rates  map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.Result != "" && !strings.EqualFold(payload.Result, "success") {
		return "", fmt.Errorf("er-api: result %q", payload.Result)
	}
	rate, ok := payload.Rates[to]
	if !ok || rate <= 0 {
		return "", fmt.Errorf("er-api: missing rate for %s", to)
	}
	return strconv.FormatFloat(rate, 'f', -1, 64), nil
}

func (p *liveRateProvider) getJSON(ctx context.Context, target string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: status %d", target, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64<<10))
}
