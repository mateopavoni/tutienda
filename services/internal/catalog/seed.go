package catalog

import (
	"time"

	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// demoCatalog returns the seed products for one example store, keyed by slug. SKUs match the stock the
// inventory service seeds for the same slug. Unknown slug ⇒ no products (a store with an empty catalog
// still renders its homepage chrome). See tenant.DemoStores for the example storefronts.
func demoCatalog(slug string) []Product {
	switch slug {
	case tenant.DemoSlug:
		return demoProducts()
	case "casa-bruta":
		return casaBrutaProducts()
	case "papel-y-tinta":
		return papelYTintaProducts()
	case "cafe-noventa":
		return cafeNoventaProducts()
	default:
		return nil
	}
}

// demoAsset points at a real product photo shipped in web/static/demo/<store>/<file> — a relative path
// resolves against the storefront's own origin, so no media/object-store config is needed for seed data.
func demoAsset(store, file string) string {
	return "/demo/" + store + "/" + file
}

// demoProducts is the seed catalog. SKUs match the stock seeded by the inventory service.
func demoProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }

	// One scheduled drop in the demo store: the System Jacket releases 15 minutes after first boot, so the
	// storefront shows a live countdown, then flips to the live no-oversell sale + real-time stock counter.
	dropAt := base.Add(15 * time.Minute)

	sizes := func(prefix string, labels ...string) []Variant {
		out := make([]Variant, 0, len(labels))
		for _, l := range labels {
			out = append(out, Variant{SKU: prefix + "-" + skuPart(l), Label: "US " + l})
		}
		return out
	}
	asset := func(file string) string { return demoAsset(tenant.DemoSlug, file) }

	return []Product{
		{
			ID: "aero-strata-x1", Name: "Aero Strata_X1", Category: "footwear",
			Description: "Engineered for structural integrity and momentum. A hyper-ventilated architectural mesh upper integrated with a brutalist kinetic sole system. Manufactured without excess.",
			PriceCents:  28500, Currency: "USD", ImageURL: asset("aero-strata-x1.jpg"),
			Images:       []string{asset("aero-strata-x1.jpg")},
			Variants:     sizes("AX1", "7", "7.5", "8", "8.5", "9", "9.5", "10", "10.5"),
			VariantLabel: "Size",
			CreatedAt:    at(0),
		},
		{
			ID: "system-jacket-xt", Name: "System Jacket X-T", Category: "outerwear",
			Description: "Nylon ripstop shell with matte black hardware. Oversized technical fit built around movement and weather.",
			PriceCents:  89000, Currency: "USD", ImageURL: asset("system-jacket-xt.webp"),
			Images: []string{asset("system-jacket-xt.webp")},
			Variants: []Variant{
				{SKU: "SJX-S", Label: "S"}, {SKU: "SJX-M", Label: "M"},
				{SKU: "SJX-L", Label: "L"}, {SKU: "SJX-XL", Label: "XL"},
			},
			VariantLabel: "Size",
			CreatedAt:    at(1),
			DropAt:       &dropAt, // scheduled drop — see comment above
		},
		{
			ID: "void-parka", Name: "Void Parka", Category: "outerwear",
			Description: "Sealed weather protection in a severe, minimalist cut. Cold, clinical, uncompromisingly modern.",
			PriceCents:  120000, Currency: "USD", ImageURL: asset("void-parka.webp"),
			Variants: []Variant{
				{SKU: "VP-S", Label: "S"}, {SKU: "VP-M", Label: "M"}, {SKU: "VP-L", Label: "L"},
			},
			VariantLabel: "Size",
			CreatedAt:    at(2),
		},
		{
			ID: "utility-rig", Name: "Utility Rig", Category: "accessories",
			Description: "Modular carry system. Heavy matte black hardware on rigid geometric webbing.",
			PriceCents:  22000, Currency: "USD", ImageURL: asset("utility-rig.avif"),
			Variants:  []Variant{{SKU: "UR-OS", Label: "One Size"}},
			CreatedAt: at(3),
		},
		{
			ID: "aero-tread-01", Name: "Aero-Tread Unit 01", Category: "footwear",
			Description: "Aggressive geometric rubber outsole, minimalist upper. Industrial technical footwear.",
			PriceCents:  45000, Currency: "USD", ImageURL: asset("aero-tread-01.webp"),
			Variants:     sizes("ATU", "8", "9", "10"),
			VariantLabel: "Size",
			CreatedAt:    at(4),
		},
		{
			ID: "tactical-vest", Name: "Tactical Vest", Category: "apparel",
			Description: "Structured utility vest with geometric seams and technical fabric. Brutalist and intensely modern.",
			PriceCents:  45000, Currency: "USD", ImageURL: asset("tactical-vest.webp"),
			Variants: []Variant{
				{SKU: "TV-M", Label: "M"}, {SKU: "TV-L", Label: "L"}, {SKU: "TV-XL", Label: "XL"},
			},
			VariantLabel: "Size",
			CreatedAt:    at(5),
		},
		{
			ID: "utility-tote", Name: "Utility Tote", Category: "accessories",
			Description: "Heavy industrial canvas tote with an uncompromising brutalist structure.",
			PriceCents:  12000, Currency: "USD", ImageURL: asset("utility-tote.avif"),
			Variants:  []Variant{{SKU: "UT-OS", Label: "One Size"}},
			CreatedAt: at(6),
		},
	}
}

// casaBrutaProducts seeds CASA BRUTA (home/deco): concrete, steel and linen objects for the brutalist home.
func casaBrutaProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	one := func(sku string) []Variant { return []Variant{{SKU: sku, Label: "One Size"}} }
	asset := func(file string) string { return demoAsset("casa-bruta", file) }
	return []Product{
		{
			ID: "concrete-lamp-01", Name: "Concrete Lamp 01", Category: "lighting",
			Description: "A cast-concrete table lamp with a raw, unsealed finish. Heavy, monolithic, warm light from a cold body.",
			PriceCents:  16500, Currency: "USD", ImageURL: asset("concrete-lamp-01.webp"),
			Images:    []string{asset("concrete-lamp-01.webp")},
			Variants:  one("CB-LAMP"),
			CreatedAt: at(0),
		},
		{
			ID: "monolith-stool", Name: "Monolith Stool", Category: "furniture",
			Description: "A single block of a stool. Powder-coated steel, square geometry, no visible joinery.",
			PriceCents:  24000, Currency: "USD", ImageURL: asset("monolith-stool.webp"),
			Variants: []Variant{
				{SKU: "CB-STOOL-BLK", Label: "Black"}, {SKU: "CB-STOOL-GRY", Label: "Grey"},
			},
			VariantLabel: "Color",
			CreatedAt:    at(1),
		},
		{
			ID: "linen-throw", Name: "Linen Throw", Category: "textiles",
			Description: "Heavyweight washed linen with a raw selvedge edge. Architectural drape, undyed tones.",
			PriceCents:  9000, Currency: "USD", ImageURL: asset("linen-throw.webp"),
			Variants: []Variant{
				{SKU: "CB-THROW-SND", Label: "Sand"}, {SKU: "CB-THROW-SLT", Label: "Slate"},
			},
			VariantLabel: "Color",
			CreatedAt:    at(2),
		},
		{
			ID: "ceramic-vessel", Name: "Ceramic Vessel", Category: "objects",
			Description: "Hand-thrown matte stoneware. Severe silhouette, intentional weight in the hand.",
			PriceCents:  6800, Currency: "USD", ImageURL: asset("ceramic-vessel.webp"),
			Variants: []Variant{
				{SKU: "CB-VES-S", Label: "Small"}, {SKU: "CB-VES-M", Label: "Medium"}, {SKU: "CB-VES-L", Label: "Large"},
			},
			VariantLabel: "Size",
			CreatedAt:    at(3),
		},
		{
			ID: "steel-shelf", Name: "Steel Shelf", Category: "furniture",
			Description: "Folded raw steel wall shelf. Bracket and surface are one continuous plane.",
			PriceCents:  13500, Currency: "USD", ImageURL: asset("steel-shelf.webp"),
			Variants:  one("CB-SHELF"),
			CreatedAt: at(4),
		},
	}
}

// papelYTintaProducts seeds PAPEL & TINTA (bookshop): an editorial bookshop on form, type and ideas.
func papelYTintaProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	one := func(sku string) []Variant { return []Variant{{SKU: sku, Label: "Paperback"}} }
	asset := func(file string) string { return demoAsset("papel-y-tinta", file) }
	return []Product{
		{
			ID: "brutalist-architecture-vol-i", Name: "Brutalist Architecture, Vol. I", Category: "art",
			Description: "A photographic survey of concrete civic architecture. 240 pages, smyth-sewn, matte stock.",
			PriceCents:  5200, Currency: "USD", ImageURL: asset("brutalist-architecture-vol-i.jpg"),
			Images:    []string{asset("brutalist-architecture-vol-i.jpg")},
			Variants:  one("PT-ARCH"),
			CreatedAt: at(0),
		},
		{
			ID: "concrete-poems", Name: "Concrete Poems", Category: "poetry",
			Description: "Typographic poetry where the layout is the line. Letterpress cover, 96 pages.",
			PriceCents:  2400, Currency: "USD", ImageURL: asset("concrete-poems.jpg"),
			Variants:  one("PT-POEMS"),
			CreatedAt: at(1),
		},
		{
			ID: "essays-on-form", Name: "Essays on Form", Category: "essays",
			Description: "Twelve essays on design, structure and restraint. Pocket format, thread-bound.",
			PriceCents:  3100, Currency: "USD", ImageURL: asset("essays-on-form.webp"),
			Variants:  one("PT-ESSAYS"),
			CreatedAt: at(2),
		},
		{
			ID: "the-grid-novel", Name: "The Grid (Novel)", Category: "fiction",
			Description: "A novel about a city planned to the millimetre, and the one street that refuses the grid.",
			PriceCents:  2800, Currency: "USD", ImageURL: asset("the-grid-novel.webp"),
			Variants:  one("PT-GRID"),
			CreatedAt: at(3),
		},
		{
			ID: "type-specimen", Name: "Type Specimen", Category: "art",
			Description: "A specimen book of a single grotesque typeface across every weight. Oversized, 180 pages.",
			PriceCents:  4500, Currency: "USD", ImageURL: asset("type-specimen.webp"),
			Variants:  one("PT-TYPE"),
			CreatedAt: at(4),
		},
	}
}

// cafeNoventaProducts seeds CAFÉ NOVENTA (coffee): single-origin beans, brewing gear and merch.
func cafeNoventaProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	one := func(sku string) []Variant { return []Variant{{SKU: sku, Label: "One Size"}} }
	bag := func(prefix string) []Variant {
		return []Variant{{SKU: prefix + "-250", Label: "250 g"}, {SKU: prefix + "-1K", Label: "1 kg"}}
	}
	asset := func(file string) string { return demoAsset("cafe-noventa", file) }
	return []Product{
		{
			ID: "single-origin-ethiopia", Name: "Single Origin — Ethiopia", Category: "beans",
			Description: "Washed Yirgacheffe. Floral, citric, tea-like. Roasted light to keep the origin loud.",
			PriceCents:  1800, Currency: "USD", ImageURL: asset("single-origin-ethiopia.jpg"),
			Images:       []string{asset("single-origin-ethiopia.jpg")},
			Variants:     bag("CN-ETH"),
			VariantLabel: "Weight",
			CreatedAt:    at(0),
		},
		{
			ID: "espresso-blend-90", Name: "Espresso Blend 90", Category: "beans",
			Description: "A dark, cocoa-heavy blend built for espresso. Thick crema, low acidity, no apologies.",
			PriceCents:  1600, Currency: "USD", ImageURL: asset("espresso-blend-90.jpg"),
			Variants:     bag("CN-ESP"),
			VariantLabel: "Weight",
			CreatedAt:    at(1),
		},
		{
			ID: "cold-brew-concentrate", Name: "Cold Brew Concentrate", Category: "beans",
			Description: "Coarse-ground blend for an 18-hour cold extraction. Dilute to taste.",
			PriceCents:  1400, Currency: "USD", ImageURL: asset("cold-brew-concentrate.webp"),
			Variants:  one("CN-COLD"),
			CreatedAt: at(2),
		},
		{
			ID: "steel-dripper", Name: "Steel Dripper", Category: "equipment",
			Description: "A single-piece stainless pour-over cone. No paper filter, no waste.",
			PriceCents:  3800, Currency: "USD", ImageURL: asset("steel-dripper.jpeg"),
			Variants:  one("CN-DRIP"),
			CreatedAt: at(3),
		},
		{
			ID: "mono-mug", Name: "Mono Mug", Category: "merch",
			Description: "Matte black stoneware mug, 300 ml. Heavy base, severe handle.",
			PriceCents:  2200, Currency: "USD", ImageURL: asset("mono-mug.webp"),
			Variants:  one("CN-MUG"),
			CreatedAt: at(4),
		},
	}
}

// skuPart turns a size label into a SKU-safe fragment ("8.5" -> "85").
func skuPart(label string) string {
	out := make([]rune, 0, len(label))
	for _, r := range label {
		if r == '.' {
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
