package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// Selection is a re-export of goquery.Selection for use in scraper files
type Selection = goquery.Selection

// WithRetry retries fn up to maxRetries times with exponential backoff
func WithRetry(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt == maxRetries {
			break
		}
		delay := baseDelay * time.Duration(attempt+1)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

// RateLimiter provides a simple delay between operations
type RateLimiter struct {
	delay time.Duration
	last  time.Time
}

func NewRateLimiter(delay time.Duration) *RateLimiter {
	return &RateLimiter{delay: delay}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	elapsed := time.Since(rl.last)
	if elapsed < rl.delay {
		wait := rl.delay - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	rl.last = time.Now()
	return nil
}

var legoIDStartRegex = regexp.MustCompile(`^\s*(\d{4,6})\b`)
var legoIDEndRegex = regexp.MustCompile(`\b(\d{4,6})\s*$`)
var legoIDAfterLego = regexp.MustCompile(`(?i)(?:lego|ლეგო)\s+(\d{4,6})\b`)

// ExtractLEGOID extracts a LEGO set number (4-6 digits) from product name
// Checks: start of name, after "LEGO"/"ლეგო", end of name
func ExtractLEGOID(name string) string {
	name = strings.TrimSpace(name)
	clean := strings.ReplaceAll(name, "-", " ")

	// Check start of name: "76452 LEGO HARRY POTTER..."
	if m := legoIDStartRegex.FindStringSubmatch(clean); len(m) > 1 {
		return m[1]
	}

	// Check right after "LEGO" or "ლეგო": "LEGO 10713 Creative Suitcase"
	if m := legoIDAfterLego.FindStringSubmatch(clean); len(m) > 1 {
		return m[1]
	}

	// Check end of name: "...Creative Suitcase 10713" or "...Quidditch - 76452"
	if m := legoIDEndRegex.FindStringSubmatch(clean); len(m) > 1 {
		return m[1]
	}

	return ""
}

var legoIDAnywhereRegex = regexp.MustCompile(`\b(\d{5,6})\b`)

// ExtractLEGOIDAnywhere finds a 5-6 digit LEGO set number anywhere in the name
func ExtractLEGOIDAnywhere(name string) string {
	// First try the strict version
	if id := ExtractLEGOID(name); id != "" {
		return id
	}
	// Fallback: find first 5-6 digit number anywhere
	clean := strings.ReplaceAll(name, "-", " ")
	if m := legoIDAnywhereRegex.FindStringSubmatch(clean); len(m) > 1 {
		return m[1]
	}
	return ""
}

// IsLEGOProduct checks if a product name indicates it's a LEGO product
func isLEGOProduct(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "lego") || strings.Contains(lower, "ლეგო")
}

var priceRegex = regexp.MustCompile(`[\d]+[.,]?\d*`)

// ParsePrice extracts a float64 from a price string like "69.95 ₾" or "69,95"
func ParsePrice(s string) (*float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	match := priceRegex.FindString(s)
	if match == "" {
		return nil, fmt.Errorf("no number found in %q", s)
	}

	match = strings.ReplaceAll(match, ",", ".")
	val, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return nil, fmt.Errorf("parse float %q: %w", match, err)
	}

	return &val, nil
}

// CalcDiscountPercent calculates discount percentage from original and discounted prices
func CalcDiscountPercent(original, discounted float64) *float64 {
	if original <= 0 || discounted >= original {
		return nil
	}
	pct := (1 - discounted/original) * 100
	return &pct
}

// FetchPageHTML uses chromedp to navigate to a URL and return the rendered HTML
func FetchPageHTML(ctx context.Context, wsURL, targetURL string, logger *slog.Logger) (*goquery.Document, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, wsURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	timeoutCtx, cancel := context.WithTimeout(browserCtx, 45*time.Second)
	defer cancel()

	var html string
	logger.Info("chromedp navigating", "url", targetURL)
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(4*time.Second),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp navigate %s: %w", targetURL, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return doc, nil
}

// FetchPageHTMLLong is like FetchPageHTML but with configurable wait time for heavy pages
func FetchPageHTMLLong(ctx context.Context, wsURL, targetURL string, waitTime time.Duration, logger *slog.Logger) (*goquery.Document, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, wsURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var html string
	logger.Info("chromedp navigating (long)", "url", targetURL, "wait", waitTime)
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(waitTime),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp navigate %s: %w", targetURL, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return doc, nil
}

// FetchPageHTMLWithSearch navigates to a URL and types in a search box
func FetchPageHTMLWithSearch(ctx context.Context, wsURL, targetURL, searchSelector, query string, logger *slog.Logger) (*goquery.Document, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, wsURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	timeoutCtx, cancel := context.WithTimeout(browserCtx, 60*time.Second)
	defer cancel()

	var html string
	logger.Info("chromedp search", "url", targetURL, "query", query)
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3*time.Second),
		chromedp.SendKeys(searchSelector, query+"\n"),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp search %s: %w", targetURL, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return doc, nil
}
