package accounts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
)

// ErrInvalidCredentials is returned for a wrong email/password pair (deliberately not distinguishing
// which, to avoid leaking which emails exist).
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrInvalidInput is returned for malformed signup/store payloads.
var ErrInvalidInput = errors.New("invalid input")

// ErrStoreLimit is returned when a merchant already owns the maximum number of stores. It caps demo
// abuse (anyone can sign up and spin up stores); the seed bypasses it (inserts directly).
var ErrStoreLimit = errors.New("store limit reached")

var slugRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,38}[a-z0-9])$`)

// Service holds the merchant/store business logic and signs session tokens.
type Service struct {
	repo      *Repository
	issuer    *authx.Issuer
	log       *slog.Logger
	maxStores int // per-merchant cap on store creation (anti-abuse); <=0 means unlimited.
}

func NewService(repo *Repository, issuer *authx.Issuer, log *slog.Logger, maxStores int) *Service {
	return &Service{repo: repo, issuer: issuer, log: log, maxStores: maxStores}
}

// Signup creates a merchant and returns a merchant-scoped session token.
func (s *Service) Signup(ctx context.Context, email, password string) (string, *Merchant, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(email, "@") || len(password) < 8 {
		return "", nil, fmt.Errorf("%w: email required and password >= 8 chars", ErrInvalidInput)
	}
	hash, err := authx.HashPassword(password)
	if err != nil {
		return "", nil, err
	}
	m := &Merchant{ID: primitive.NewObjectID(), Email: email, PasswordHash: hash, Role: authx.RoleMerchant, CreatedAt: time.Now().UTC()}
	if err := s.repo.insertMerchant(ctx, m); err != nil {
		return "", nil, err // ErrDuplicate bubbles up as a 409
	}
	token, err := s.issuer.IssueMerchant(m.ID.Hex(), m.Role)
	if err != nil {
		return "", nil, err
	}
	return token, m, nil
}

// Login verifies credentials and returns a merchant-scoped session token.
func (s *Service) Login(ctx context.Context, email, password string) (string, *Merchant, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	m, err := s.repo.merchantByEmail(ctx, email)
	if errors.Is(err, ErrMerchantNotFound) {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if !authx.CheckPassword(m.PasswordHash, password) {
		return "", nil, ErrInvalidCredentials
	}
	token, err := s.issuer.IssueMerchant(m.ID.Hex(), m.Role)
	if err != nil {
		return "", nil, err
	}
	return token, m, nil
}

// CreateStore registers a new tenant owned by the merchant.
func (s *Service) CreateStore(ctx context.Context, ownerID, slug, displayName, theme, accentColor string) (*Store, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if !slugRe.MatchString(slug) {
		return nil, fmt.Errorf("%w: slug must be 3-40 lowercase alphanumeric/hyphen chars", ErrInvalidInput)
	}
	if displayName == "" {
		displayName = slug
	}
	if s.maxStores > 0 {
		owned, err := s.repo.storesByOwner(ctx, ownerID)
		if err != nil {
			return nil, err
		}
		if len(owned) >= s.maxStores {
			return nil, fmt.Errorf("%w: max %d stores per merchant", ErrStoreLimit, s.maxStores)
		}
	}
	store := &Store{
		ID:          primitive.NewObjectID().Hex(),
		Slug:        slug,
		OwnerID:     ownerID,
		DisplayName: displayName,
		Settings: Settings{
			// A blank accent ⇒ the storefront falls back to the chosen theme's own accent, so a new store
			// looks like its theme out of the box instead of every store sharing one hardcoded color.
			AccentColor: sanitizeHexAccent(accentColor),
			Currency:    "USD",
			Theme:       normalizeTheme(theme),
			// Seed a hero over the mandatory grid so the homepage looks built, not an empty catalog.
			Layout: []Section{
				{Type: "hero", Enabled: true, Heading: displayName},
				{Type: "catalog", Enabled: true},
			},
		},
		Plan:      string(plan.Free), // every store starts free; upgrades flip this tier
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.insertStore(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

// MyStores lists the stores owned by the merchant.
func (s *Service) MyStores(ctx context.Context, ownerID string) ([]Store, error) {
	stores, err := s.repo.storesByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if stores == nil {
		stores = []Store{}
	}
	return stores, nil
}

// StoreToken issues a store-scoped token after verifying the merchant owns the store. The signed
// StoreID claim is what the gateway turns into X-Tenant-ID, so ownership is enforced at issue time.
func (s *Service) StoreToken(ctx context.Context, ownerID, storeID string) (string, *Store, error) {
	store, err := s.repo.storeByID(ctx, storeID)
	if err != nil {
		return "", nil, err
	}
	if store.OwnerID != ownerID {
		return "", nil, ErrStoreNotFound // don't reveal existence of stores the caller doesn't own
	}
	// Sign the store's plan into the token so the gateway can gate paid features without a DB hit.
	token, err := s.issuer.IssueStore(ownerID, store.ID, store.Plan)
	if err != nil {
		return "", nil, err
	}
	return token, store, nil
}

// SetPlan changes a store's membership tier, scoped to its owner. The caller must pass a valid tier
// (validated here); the merchant re-fetches a store token afterwards to pick up the new plan claim.
// This is the self-serve upgrade path; a billing webhook or the super-admin would call the same method.
func (s *Service) SetPlan(ctx context.Context, ownerID, storeID, tier string) (*Store, error) {
	if !plan.Valid(tier) {
		return nil, fmt.Errorf("%w: unknown plan %q", ErrInvalidInput, tier)
	}
	return s.repo.updatePlan(ctx, storeID, ownerID, tier)
}

// UpdateStore patches a store's settings/displayName, scoped to the owner.
func (s *Service) UpdateStore(ctx context.Context, ownerID, storeID, displayName string, settings Settings) (*Store, error) {
	settings.Layout = sanitizeLayout(settings.Layout)
	settings.Pages = sanitizePages(settings.Pages)
	settings.Theme = normalizeTheme(settings.Theme)
	return s.repo.updateSettings(ctx, storeID, ownerID, displayName, settings)
}

// defaultTheme is the storefront template every store falls back to. knownThemes is the closed set a
// merchant may pick (mirrors THEMES in web/src/lib/theme.ts); anything else is a typo/tampered value.
const defaultTheme = "monolith"

var knownThemes = map[string]bool{
	"monolith": true,
	"boutique": true,
	"pop":      true,
	"terminal": true,
}

// normalizeTheme keeps a known theme, else falls back to the default — the same "invalid value ⇒ safe
// default" stance as an unknown plan tier becoming free. The storefront also defaults on render.
func normalizeTheme(t string) string {
	t = strings.TrimSpace(t)
	if knownThemes[t] {
		return t
	}
	return defaultTheme
}

var hexAccentRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// sanitizeHexAccent returns a valid #rrggbb accent (lowercased) or "" for anything else. A blank accent
// is meaningful: the storefront then uses the chosen theme's own accent instead of a forced color.
func sanitizeHexAccent(c string) string {
	c = strings.TrimSpace(c)
	if hexAccentRe.MatchString(c) {
		return strings.ToLower(c)
	}
	return ""
}

// knownSectionTypes is the closed set of storefront blocks a merchant may arrange. Anything else is a
// typo or a tampered request and gets dropped — the same defensive stance as an unknown plan tier
// falling back to free: a bad value never persists. Mirrors SECTION_META in web/src/lib/layout.ts.
var knownSectionTypes = map[string]bool{
	"announcement": true,
	"hero":         true,
	"feature":      true,
	"catalog":      true,
	"about":        true,
	"contact":      true,
	"gallery":      true,
	"faq":          true,
}

// sectionFieldMax caps each piece of section copy so a merchant can't store an unbounded blob in the
// store document (which the storefront SSR would then have to ship on every page load). maxSectionItems
// bounds the repeatable items (gallery images / FAQ entries) for the same reason.
const (
	sectionFieldMax = 2000
	maxSectionItems = 24

	// heroOverlayDefault/Min/Max bound the hero's overlay peak strength (a percentage, at the edge nearest
	// the text — see the gradient rendering in the storefront). Min/max keep the photo always at least
	// partly visible and the text always at least partly protected.
	heroOverlayDefault = 55
	heroOverlayMin     = 15
	heroOverlayMax     = 90
)

// knownOverlayColors is the closed set for a hero's overlay tint: "black" (default, white text — the
// classic photo-hero look) or "white" (dark text, for a merchant who wants an airy/bright hero). Anything
// else is a typo and falls back to "black" — same defensive stance as an unknown section/theme/plan.
var knownOverlayColors = map[string]bool{
	"black": true,
	"white": true,
}

// sanitizeLayout normalizes a layout on write: it keeps only known section types, drops duplicates,
// trims/caps the copy, and forces the mandatory product grid ("catalog") to stay enabled. The storefront
// also normalizes at render time (normalizeLayout), but sanitizing on write keeps the stored document
// clean and trustworthy regardless of which client wrote it.
func sanitizeLayout(raw []Section) []Section {
	if len(raw) == 0 {
		return nil
	}
	out := make([]Section, 0, len(raw))
	seen := map[string]bool{}
	for _, sec := range raw {
		if !knownSectionTypes[sec.Type] || seen[sec.Type] {
			continue
		}
		seen[sec.Type] = true
		sec.Heading = capRunes(strings.TrimSpace(sec.Heading), sectionFieldMax)
		sec.Body = capRunes(strings.TrimSpace(sec.Body), sectionFieldMax)
		sec.ImageURL = capRunes(strings.TrimSpace(sec.ImageURL), sectionFieldMax)
		sec.CTALabel = capRunes(strings.TrimSpace(sec.CTALabel), sectionFieldMax)
		sec.Items = sanitizeItems(sec.Items)
		if sec.Type == "catalog" {
			sec.Enabled = true // the grid is mandatory; it can be reordered but never disabled.
		}
		if sec.Type == "hero" {
			switch {
			case sec.Overlay <= 0:
				sec.Overlay = heroOverlayDefault
			case sec.Overlay < heroOverlayMin:
				sec.Overlay = heroOverlayMin
			case sec.Overlay > heroOverlayMax:
				sec.Overlay = heroOverlayMax
			}
			if !knownOverlayColors[sec.OverlayColor] {
				sec.OverlayColor = "black"
			}
		}
		out = append(out, sec)
	}
	return out
}

// sanitizeItems trims/caps each repeatable item's copy, drops fully-empty items (so a half-edited row
// never persists) and caps the count. Returns nil for an empty result so the stored document stays clean.
func sanitizeItems(raw []SectionItem) []SectionItem {
	if len(raw) == 0 {
		return nil
	}
	out := make([]SectionItem, 0, len(raw))
	for _, it := range raw {
		it.Heading = capRunes(strings.TrimSpace(it.Heading), sectionFieldMax)
		it.Body = capRunes(strings.TrimSpace(it.Body), sectionFieldMax)
		it.ImageURL = capRunes(strings.TrimSpace(it.ImageURL), sectionFieldMax)
		if it.Heading == "" && it.Body == "" && it.ImageURL == "" {
			continue
		}
		out = append(out, it)
		if len(out) >= maxSectionItems {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// knownPageTypes is the closed set of standalone storefront pages. Same defensive stance as
// knownSectionTypes: an unknown type is a typo/tampered request and gets dropped. Mirrors PAGE_TYPES in
// web/src/lib/pages.ts.
var knownPageTypes = map[string]bool{
	"about":   true,
	"faq":     true,
	"contact": true,
	"terms":   true,
}

// pageBodyMax caps a page body. Bigger than a section field (a terms/policy page is legitimately long)
// but still bounded so the store document can't grow without limit. The title reuses sectionFieldMax.
const pageBodyMax = 20000

// sanitizePages normalizes the pages on write: keeps only known types, drops duplicates, trims/caps the
// title and body, and drops any page with an empty body (an empty page is treated as absent — nothing to
// route or link). Write-time guarantee mirroring sanitizeLayout: the stored document is always clean.
func sanitizePages(raw []Page) []Page {
	if len(raw) == 0 {
		return nil
	}
	out := make([]Page, 0, len(raw))
	seen := map[string]bool{}
	for _, p := range raw {
		if !knownPageTypes[p.Type] || seen[p.Type] {
			continue
		}
		p.Title = capRunes(strings.TrimSpace(p.Title), sectionFieldMax)
		p.Body = capRunes(strings.TrimSpace(p.Body), pageBodyMax)
		if p.Body == "" {
			continue
		}
		seen[p.Type] = true
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// capRunes truncates s to at most n runes (not bytes, to avoid splitting a multi-byte character).
func capRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// StoreBySlug resolves a store from its public slug (used by the gateway and storefront).
func (s *Service) StoreBySlug(ctx context.Context, slug string) (*Store, error) {
	return s.repo.storeBySlug(ctx, strings.TrimSpace(strings.ToLower(slug)))
}

// --- Platform super-admin (cross-tenant). The gateway gates these on RoleAdmin; the methods below do
// not re-check ownership on purpose: an admin operates across every tenant. ---

// PlatformStats is the SaaS-wide summary the super-admin console shows.
type PlatformStats struct {
	Merchants    int64            `json:"merchants"`
	Stores       int64            `json:"stores"`
	StoresByPlan map[string]int64 `json:"storesByPlan"`
}

// AllStores lists every store across all tenants (super-admin only).
func (s *Service) AllStores(ctx context.Context) ([]Store, error) {
	stores, err := s.repo.allStores(ctx)
	if err != nil {
		return nil, err
	}
	if stores == nil {
		stores = []Store{}
	}
	return stores, nil
}

// AllMerchants lists every merchant across the platform (super-admin only).
func (s *Service) AllMerchants(ctx context.Context) ([]Merchant, error) {
	merchants, err := s.repo.allMerchants(ctx)
	if err != nil {
		return nil, err
	}
	if merchants == nil {
		merchants = []Merchant{}
	}
	return merchants, nil
}

// Stats returns platform-wide counts, including a per-plan store breakdown.
func (s *Service) Stats(ctx context.Context) (*PlatformStats, error) {
	merchants, err := s.repo.countMerchants(ctx)
	if err != nil {
		return nil, err
	}
	stores, err := s.repo.allStores(ctx)
	if err != nil {
		return nil, err
	}
	byPlan := map[string]int64{}
	for _, st := range stores {
		byPlan[string(plan.Normalize(st.Plan))]++
	}
	return &PlatformStats{Merchants: merchants, Stores: int64(len(stores)), StoresByPlan: byPlan}, nil
}

// SetPlanAsAdmin changes any store's plan, regardless of owner (super-admin only).
func (s *Service) SetPlanAsAdmin(ctx context.Context, storeID, tier string) (*Store, error) {
	if !plan.Valid(tier) {
		return nil, fmt.Errorf("%w: unknown plan %q", ErrInvalidInput, tier)
	}
	return s.repo.updatePlanByID(ctx, storeID, tier)
}
