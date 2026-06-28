package tenant

// Demo identifies the seeded showcase store ("SYSTEM ARCHIVE"). Every service seeds its demo data
// under this fixed tenant id so the docker-compose smoke test and the original storefront keep working
// after the multi-tenant pivot. The storefront resolves it by DemoSlug.
const (
	DemoID   = "000000000000000000000001"
	DemoSlug = "system-archive"
	DemoName = "SYSTEM ARCHIVE"
)

// DemoStore is one seeded EXAMPLE store. The whole DemoStores list is sample data for the
// showcase/portfolio demo — these are NOT real merchants. Each service seeds its slice under the
// fixed ID (accounts: the store + the shared demo owner, catalog: products, inventory: stock) so every
// storefront resolves by Slug out of the box. Remove or override these in any real deployment.
type DemoStore struct {
	ID      string
	Slug    string
	Name    string
	Accent  string // storefront accent color
	Plan    string // membership tier (see platform/plan): free | pro | scale
	Theme   string // storefront visual theme (see web/src/lib/theme.ts): monolith | boutique | pop | terminal
	Tagline string // hero copy seeded into the storefront homepage
}

// DemoStores are the example storefronts seeded for the demo, one per rubro so the dashboard and the
// landing showcase a real multi-tenant spread (apparel, home/deco, bookshop, coffee). The first entry
// is the original SYSTEM ARCHIVE showcase (kept verbatim, scale tier, the loadtest target). All four
// are owned by the same demo merchant so a single login shows every store in /app.
var DemoStores = []DemoStore{
	{DemoID, DemoSlug, DemoName, "#1b03ea", "scale", "monolith", "Engineered garments. Manufactured without excess."},
	{"000000000000000000000002", "casa-bruta", "CASA BRUTA", "#b5651d", "pro", "boutique", "Objects for the brutalist home — concrete, steel, linen."},
	{"000000000000000000000003", "papel-y-tinta", "PAPEL & TINTA", "#1f6f4a", "free", "terminal", "An editorial bookshop. Form, type, and ideas on paper."},
	{"000000000000000000000004", "cafe-noventa", "CAFÉ NOVENTA", "#6f4e37", "pro", "pop", "Single-origin coffee, roasted to a severe standard."},
}
