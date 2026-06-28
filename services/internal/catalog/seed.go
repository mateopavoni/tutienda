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

// demoImg is the shared placeholder image source: a stable seeded picsum URL the storefront renders
// grayscale per the design system and reveals on hover.
func demoImg(seed string) string { return "https://picsum.photos/seed/" + seed + "/900/1100" }

// demoProducts is the seed catalog. SKUs match the stock seeded by the inventory service.
// Images use a stable seeded placeholder source; the storefront renders them grayscale per the
// design system and reveals colour on hover.
func demoProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	img := func(seed string) string { return "https://picsum.photos/seed/" + seed + "/900/1100" }

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

	return []Product{
		{
			ID: "aero-strata-x1", Name: "Aero Strata_X1", Category: "footwear",
			Description: "Engineered for structural integrity and momentum. A hyper-ventilated architectural mesh upper integrated with a brutalist kinetic sole system. Manufactured without excess.",
			PriceCents:  28500, Currency: "USD", ImageURL: img("aero-strata-x1"),
			// A multi-image product so the storefront gallery is visible out of the box.
			Images:    []string{img("aero-strata-x1"), img("aero-strata-x1-b"), img("aero-strata-x1-c")},
			Variants:  sizes("AX1", "7", "7.5", "8", "8.5", "9", "9.5", "10", "10.5"),
			CreatedAt: at(0),
		},
		{
			ID: "system-jacket-xt", Name: "System Jacket X-T", Category: "outerwear",
			Description: "Nylon ripstop shell with matte black hardware. Oversized technical fit built around movement and weather.",
			PriceCents:  89000, Currency: "USD", ImageURL: img("system-jacket-xt"),
			Images:      []string{img("system-jacket-xt"), img("system-jacket-xt-b")},
			Variants: []Variant{
				{SKU: "SJX-S", Label: "S"}, {SKU: "SJX-M", Label: "M"},
				{SKU: "SJX-L", Label: "L"}, {SKU: "SJX-XL", Label: "XL"},
			},
			CreatedAt: at(1),
			DropAt:    &dropAt, // scheduled drop — see comment above
		},
		{
			ID: "void-parka", Name: "Void Parka", Category: "outerwear",
			Description: "Sealed weather protection in a severe, minimalist cut. Cold, clinical, uncompromisingly modern.",
			PriceCents:  120000, Currency: "USD", ImageURL: img("void-parka"),
			Variants: []Variant{
				{SKU: "VP-S", Label: "S"}, {SKU: "VP-M", Label: "M"}, {SKU: "VP-L", Label: "L"},
			},
			CreatedAt: at(2),
		},
		{
			ID: "utility-rig", Name: "Utility Rig", Category: "accessories",
			Description: "Modular carry system. Heavy matte black hardware on rigid geometric webbing.",
			PriceCents:  22000, Currency: "USD", ImageURL: img("utility-rig"),
			Variants:    []Variant{{SKU: "UR-OS", Label: "One Size"}},
			CreatedAt:   at(3),
		},
		{
			ID: "aero-tread-01", Name: "Aero-Tread Unit 01", Category: "footwear",
			Description: "Aggressive geometric rubber outsole, minimalist upper. Industrial technical footwear.",
			PriceCents:  45000, Currency: "USD", ImageURL: img("aero-tread-01"),
			Variants:  sizes("ATU", "8", "9", "10"),
			CreatedAt: at(4),
		},
		{
			ID: "tactical-vest", Name: "Tactical Vest", Category: "apparel",
			Description: "Structured utility vest with geometric seams and technical fabric. Brutalist and intensely modern.",
			PriceCents:  45000, Currency: "USD", ImageURL: img("tactical-vest"),
			Variants: []Variant{
				{SKU: "TV-M", Label: "M"}, {SKU: "TV-L", Label: "L"}, {SKU: "TV-XL", Label: "XL"},
			},
			CreatedAt: at(5),
		},
		{
			ID: "utility-tote", Name: "Utility Tote", Category: "accessories",
			Description: "Heavy industrial canvas tote with an uncompromising brutalist structure.",
			PriceCents:  12000, Currency: "USD", ImageURL: img("utility-tote"),
			Variants:    []Variant{{SKU: "UT-OS", Label: "One Size"}},
			CreatedAt:   at(6),
		},
	}
}

// casaBrutaProducts seeds CASA BRUTA (home/deco): concrete, steel and linen objects for the brutalist home.
func casaBrutaProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	one := func(sku string) []Variant { return []Variant{{SKU: sku, Label: "One Size"}} }
	return []Product{
		{
			ID: "concrete-lamp-01", Name: "Concrete Lamp 01", Category: "lighting",
			Description: "A cast-concrete table lamp with a raw, unsealed finish. Heavy, monolithic, warm light from a cold body.",
			PriceCents:  16500, Currency: "USD", ImageURL: demoImg("concrete-lamp-01"),
			Images:    []string{demoImg("concrete-lamp-01"), demoImg("concrete-lamp-01-b")},
			Variants:  one("CB-LAMP"),
			CreatedAt: at(0),
		},
		{
			ID: "monolith-stool", Name: "Monolith Stool", Category: "furniture",
			Description: "A single block of a stool. Powder-coated steel, square geometry, no visible joinery.",
			PriceCents:  24000, Currency: "USD", ImageURL: demoImg("monolith-stool"),
			Variants: []Variant{
				{SKU: "CB-STOOL-BLK", Label: "Black"}, {SKU: "CB-STOOL-GRY", Label: "Grey"},
			},
			CreatedAt: at(1),
		},
		{
			ID: "linen-throw", Name: "Linen Throw", Category: "textiles",
			Description: "Heavyweight washed linen with a raw selvedge edge. Architectural drape, undyed tones.",
			PriceCents:  9000, Currency: "USD", ImageURL: demoImg("linen-throw"),
			Variants: []Variant{
				{SKU: "CB-THROW-SND", Label: "Sand"}, {SKU: "CB-THROW-SLT", Label: "Slate"},
			},
			CreatedAt: at(2),
		},
		{
			ID: "ceramic-vessel", Name: "Ceramic Vessel", Category: "objects",
			Description: "Hand-thrown matte stoneware. Severe silhouette, intentional weight in the hand.",
			PriceCents:  6800, Currency: "USD", ImageURL: demoImg("ceramic-vessel"),
			Variants: []Variant{
				{SKU: "CB-VES-S", Label: "Small"}, {SKU: "CB-VES-M", Label: "Medium"}, {SKU: "CB-VES-L", Label: "Large"},
			},
			CreatedAt: at(3),
		},
		{
			ID: "steel-shelf", Name: "Steel Shelf", Category: "furniture",
			Description: "Folded raw steel wall shelf. Bracket and surface are one continuous plane.",
			PriceCents:  13500, Currency: "USD", ImageURL: demoImg("steel-shelf"),
			Variants:    one("CB-SHELF"),
			CreatedAt:   at(4),
		},
	}
}

// papelYTintaProducts seeds PAPEL & TINTA (bookshop): an editorial bookshop on form, type and ideas.
func papelYTintaProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	one := func(sku string) []Variant { return []Variant{{SKU: sku, Label: "Paperback"}} }
	return []Product{
		{
			ID: "brutalist-architecture-vol-i", Name: "Brutalist Architecture, Vol. I", Category: "art",
			Description: "A photographic survey of concrete civic architecture. 240 pages, smyth-sewn, matte stock.",
			PriceCents:  5200, Currency: "USD", ImageURL: demoImg("pt-arch"),
			Images:    []string{demoImg("pt-arch"), demoImg("pt-arch-b")},
			Variants:  one("PT-ARCH"),
			CreatedAt: at(0),
		},
		{
			ID: "concrete-poems", Name: "Concrete Poems", Category: "poetry",
			Description: "Typographic poetry where the layout is the line. Letterpress cover, 96 pages.",
			PriceCents:  2400, Currency: "USD", ImageURL: demoImg("pt-poems"),
			Variants:    one("PT-POEMS"),
			CreatedAt:   at(1),
		},
		{
			ID: "essays-on-form", Name: "Essays on Form", Category: "essays",
			Description: "Twelve essays on design, structure and restraint. Pocket format, thread-bound.",
			PriceCents:  3100, Currency: "USD", ImageURL: demoImg("pt-essays"),
			Variants:    one("PT-ESSAYS"),
			CreatedAt:   at(2),
		},
		{
			ID: "the-grid-novel", Name: "The Grid (Novel)", Category: "fiction",
			Description: "A novel about a city planned to the millimetre, and the one street that refuses the grid.",
			PriceCents:  2800, Currency: "USD", ImageURL: demoImg("pt-grid"),
			Variants:    one("PT-GRID"),
			CreatedAt:   at(3),
		},
		{
			ID: "type-specimen", Name: "Type Specimen", Category: "art",
			Description: "A specimen book of a single grotesque typeface across every weight. Oversized, 180 pages.",
			PriceCents:  4500, Currency: "USD", ImageURL: demoImg("pt-type"),
			Variants:    one("PT-TYPE"),
			CreatedAt:   at(4),
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
	return []Product{
		{
			ID: "single-origin-ethiopia", Name: "Single Origin — Ethiopia", Category: "beans",
			Description: "Washed Yirgacheffe. Floral, citric, tea-like. Roasted light to keep the origin loud.",
			PriceCents:  1800, Currency: "USD", ImageURL: demoImg("cn-ethiopia"),
			Images:    []string{demoImg("cn-ethiopia"), demoImg("cn-ethiopia-b")},
			Variants:  bag("CN-ETH"),
			CreatedAt: at(0),
		},
		{
			ID: "espresso-blend-90", Name: "Espresso Blend 90", Category: "beans",
			Description: "A dark, cocoa-heavy blend built for espresso. Thick crema, low acidity, no apologies.",
			PriceCents:  1600, Currency: "USD", ImageURL: demoImg("cn-espresso"),
			Variants:    bag("CN-ESP"),
			CreatedAt:   at(1),
		},
		{
			ID: "cold-brew-concentrate", Name: "Cold Brew Concentrate", Category: "beans",
			Description: "Coarse-ground blend for an 18-hour cold extraction. Dilute to taste.",
			PriceCents:  1400, Currency: "USD", ImageURL: demoImg("cn-coldbrew"),
			Variants:    one("CN-COLD"),
			CreatedAt:   at(2),
		},
		{
			ID: "steel-dripper", Name: "Steel Dripper", Category: "equipment",
			Description: "A single-piece stainless pour-over cone. No paper filter, no waste.",
			PriceCents:  3800, Currency: "USD", ImageURL: demoImg("cn-dripper"),
			Variants:    one("CN-DRIP"),
			CreatedAt:   at(3),
		},
		{
			ID: "mono-mug", Name: "Mono Mug", Category: "merch",
			Description: "Matte black stoneware mug, 300 ml. Heavy base, severe handle.",
			PriceCents:  2200, Currency: "USD", ImageURL: demoImg("cn-mug"),
			Variants:    one("CN-MUG"),
			CreatedAt:   at(4),
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
