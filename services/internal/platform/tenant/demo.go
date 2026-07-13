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
	ID        string
	Slug      string
	Name      string
	Accent    string // storefront accent color
	Plan      string // membership tier (see platform/plan): free | pro | scale
	Theme     string // storefront visual theme (see web/src/lib/theme.ts): monolith | boutique | pop | terminal
	Tagline   string // hero copy seeded into the storefront homepage
	HeroImage string // full-bleed hero background photo (Unsplash direct link)
	About     string // "about" section body seeded below the grid
}

// DemoStores are the example storefronts seeded for the demo, one per rubro so the dashboard and the
// landing showcase a real multi-tenant spread (apparel, home/deco, bookshop, coffee). The first entry
// is the original SYSTEM ARCHIVE showcase (kept verbatim, scale tier, the loadtest target). All four
// are owned by the same demo merchant so a single login shows every store in /app.
var DemoStores = []DemoStore{
	{
		DemoID, DemoSlug, DemoName, "#5433eb", "scale", "monolith",
		"Engineered garments. Manufactured without excess.",
		"/demo/system-archive/hero.jpg",
		"SYSTEM ARCHIVE designs for the long run: technical fabrics, minimal patterning, no seasonal noise. Every piece is built to outlast the trend it launched in.",
	},
	{
		"000000000000000000000002", "casa-bruta", "CASA BRUTA", "#b5651d", "pro", "boutique",
		"Objects for the brutalist home — concrete, steel, linen.",
		"https://images.unsplash.com/photo-1586871608370-4adee64d1794?w=1920&h=1080&fit=crop&q=80",
		"CASA BRUTA sources and makes objects for the honest home: cast concrete, raw steel, undyed linen. Nothing decorative that doesn't also work.",
	},
	{
		"000000000000000000000003", "papel-y-tinta", "PAPEL & TINTA", "#1f6f4a", "free", "terminal",
		"An editorial bookshop. Form, type, and ideas on paper.",
		"https://images.unsplash.com/photo-1543248939-4296e1fea89b?w=1920&h=1080&fit=crop&q=80",
		"PAPEL & TINTA is a small shop for people who still read the colophon: independent presses, typographic essays, and fiction we'd defend at a dinner party.",
	},
	{
		"000000000000000000000004", "cafe-noventa", "CAFÉ NOVENTA", "#6f4e37", "pro", "pop",
		"Single-origin coffee, roasted to a severe standard.",
		"https://images.unsplash.com/photo-1600093463592-8e36ae95ef56?w=1920&h=1080&fit=crop&q=80",
		"CAFÉ NOVENTA roasts small batches of single-origin coffee, every lot cupped and logged before it ships. No blends, no house roast — just the bean, done right.",
	},
}
