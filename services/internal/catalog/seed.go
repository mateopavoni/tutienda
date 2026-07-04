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

// demoImg builds a real, category-matched Unsplash photo URL (curated per product below — not a
// random seed) sized to the storefront's 900x1100 product-card aspect.
func demoImg(photoID string) string {
	return "https://images.unsplash.com/photo-" + photoID + "?w=900&h=1100&fit=crop&q=80"
}

// demoProducts is the seed catalog. SKUs match the stock seeded by the inventory service.
// Images are curated Unsplash photos matching each product's category (see demoImg).
func demoProducts() []Product {
	base := time.Now().UTC()
	at := func(i int) time.Time { return base.Add(time.Duration(-i) * time.Minute) }
	img := demoImg

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
			PriceCents:  28500, Currency: "USD", ImageURL: img("1544441892-794166f1e3be"),
			// A multi-image product so the storefront gallery is visible out of the box.
			Images: []string{
				img("1544441892-794166f1e3be"), img("1629097499121-290fd9a95dee"), img("1596505376266-21f87d226d7a"),
			},
			Variants:  sizes("AX1", "7", "7.5", "8", "8.5", "9", "9.5", "10", "10.5"),
			CreatedAt: at(0),
		},
		{
			ID: "system-jacket-xt", Name: "System Jacket X-T", Category: "outerwear",
			Description: "Nylon ripstop shell with matte black hardware. Oversized technical fit built around movement and weather.",
			PriceCents:  89000, Currency: "USD", ImageURL: img("1654719796836-62b889d4598d"),
			Images:      []string{img("1654719796836-62b889d4598d"), img("1614031679232-0dae776a72ee")},
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
			PriceCents:  120000, Currency: "USD", ImageURL: img("1733330808072-faa5ca09d4f1"),
			Variants: []Variant{
				{SKU: "VP-S", Label: "S"}, {SKU: "VP-M", Label: "M"}, {SKU: "VP-L", Label: "L"},
			},
			CreatedAt: at(2),
		},
		{
			ID: "utility-rig", Name: "Utility Rig", Category: "accessories",
			Description: "Modular carry system. Heavy matte black hardware on rigid geometric webbing.",
			PriceCents:  22000, Currency: "USD", ImageURL: img("1622560481979-f5b0174242a0"),
			Variants:    []Variant{{SKU: "UR-OS", Label: "One Size"}},
			CreatedAt:   at(3),
		},
		{
			ID: "aero-tread-01", Name: "Aero-Tread Unit 01", Category: "footwear",
			Description: "Aggressive geometric rubber outsole, minimalist upper. Industrial technical footwear.",
			PriceCents:  45000, Currency: "USD", ImageURL: img("1562424995-2efe650421dd"),
			Variants:  sizes("ATU", "8", "9", "10"),
			CreatedAt: at(4),
		},
		{
			ID: "tactical-vest", Name: "Tactical Vest", Category: "apparel",
			Description: "Structured utility vest with geometric seams and technical fabric. Brutalist and intensely modern.",
			PriceCents:  45000, Currency: "USD", ImageURL: img("1638104977265-8e8ea24aa2aa"),
			Variants: []Variant{
				{SKU: "TV-M", Label: "M"}, {SKU: "TV-L", Label: "L"}, {SKU: "TV-XL", Label: "XL"},
			},
			CreatedAt: at(5),
		},
		{
			ID: "utility-tote", Name: "Utility Tote", Category: "accessories",
			Description: "Heavy industrial canvas tote with an uncompromising brutalist structure.",
			PriceCents:  12000, Currency: "USD", ImageURL: img("1578237493287-8d4d2b03591a"),
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
			PriceCents:  16500, Currency: "USD", ImageURL: demoImg("1667312939978-64cf31718a6e"),
			Images:    []string{demoImg("1667312939978-64cf31718a6e"), demoImg("1579326882518-21eaa7893b02")},
			Variants:  one("CB-LAMP"),
			CreatedAt: at(0),
		},
		{
			ID: "monolith-stool", Name: "Monolith Stool", Category: "furniture",
			Description: "A single block of a stool. Powder-coated steel, square geometry, no visible joinery.",
			PriceCents:  24000, Currency: "USD", ImageURL: demoImg("1618520408841-61a1c5ee84d6"),
			Variants: []Variant{
				{SKU: "CB-STOOL-BLK", Label: "Black"}, {SKU: "CB-STOOL-GRY", Label: "Grey"},
			},
			CreatedAt: at(1),
		},
		{
			ID: "linen-throw", Name: "Linen Throw", Category: "textiles",
			Description: "Heavyweight washed linen with a raw selvedge edge. Architectural drape, undyed tones.",
			PriceCents:  9000, Currency: "USD", ImageURL: demoImg("1634665810235-011d663754e7"),
			Variants: []Variant{
				{SKU: "CB-THROW-SND", Label: "Sand"}, {SKU: "CB-THROW-SLT", Label: "Slate"},
			},
			CreatedAt: at(2),
		},
		{
			ID: "ceramic-vessel", Name: "Ceramic Vessel", Category: "objects",
			Description: "Hand-thrown matte stoneware. Severe silhouette, intentional weight in the hand.",
			PriceCents:  6800, Currency: "USD", ImageURL: demoImg("1631125915902-d8abe9225ff2"),
			Variants: []Variant{
				{SKU: "CB-VES-S", Label: "Small"}, {SKU: "CB-VES-M", Label: "Medium"}, {SKU: "CB-VES-L", Label: "Large"},
			},
			CreatedAt: at(3),
		},
		{
			ID: "steel-shelf", Name: "Steel Shelf", Category: "furniture",
			Description: "Folded raw steel wall shelf. Bracket and surface are one continuous plane.",
			PriceCents:  13500, Currency: "USD", ImageURL: demoImg("1593085260707-5377ba37f868"),
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
			PriceCents:  5200, Currency: "USD", ImageURL: demoImg("1527576539890-dfa815648363"),
			Images:    []string{demoImg("1527576539890-dfa815648363"), demoImg("1560131653-63257db002c8")},
			Variants:  one("PT-ARCH"),
			CreatedAt: at(0),
		},
		{
			ID: "concrete-poems", Name: "Concrete Poems", Category: "poetry",
			Description: "Typographic poetry where the layout is the line. Letterpress cover, 96 pages.",
			PriceCents:  2400, Currency: "USD", ImageURL: demoImg("1716892001657-722e6e14ae29"),
			Variants:    one("PT-POEMS"),
			CreatedAt:   at(1),
		},
		{
			ID: "essays-on-form", Name: "Essays on Form", Category: "essays",
			Description: "Twelve essays on design, structure and restraint. Pocket format, thread-bound.",
			PriceCents:  3100, Currency: "USD", ImageURL: demoImg("1497633762265-9d179a990aa6"),
			Variants:    one("PT-ESSAYS"),
			CreatedAt:   at(2),
		},
		{
			ID: "the-grid-novel", Name: "The Grid (Novel)", Category: "fiction",
			Description: "A novel about a city planned to the millimetre, and the one street that refuses the grid.",
			PriceCents:  2800, Currency: "USD", ImageURL: demoImg("1526050071463-2c476b162a4c"),
			Variants:    one("PT-GRID"),
			CreatedAt:   at(3),
		},
		{
			ID: "type-specimen", Name: "Type Specimen", Category: "art",
			Description: "A specimen book of a single grotesque typeface across every weight. Oversized, 180 pages.",
			PriceCents:  4500, Currency: "USD", ImageURL: demoImg("1609605348579-3123e3d40eb8"),
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
			PriceCents:  1800, Currency: "USD", ImageURL: demoImg("1710752213589-2dea3a0ef23b"),
			Images:    []string{demoImg("1710752213589-2dea3a0ef23b"), demoImg("1679917010073-ead631d41d91")},
			Variants:  bag("CN-ETH"),
			CreatedAt: at(0),
		},
		{
			ID: "espresso-blend-90", Name: "Espresso Blend 90", Category: "beans",
			Description: "A dark, cocoa-heavy blend built for espresso. Thick crema, low acidity, no apologies.",
			PriceCents:  1600, Currency: "USD", ImageURL: demoImg("1675306408031-a9aad9f23308"),
			Variants:    bag("CN-ESP"),
			CreatedAt:   at(1),
		},
		{
			ID: "cold-brew-concentrate", Name: "Cold Brew Concentrate", Category: "beans",
			Description: "Coarse-ground blend for an 18-hour cold extraction. Dilute to taste.",
			PriceCents:  1400, Currency: "USD", ImageURL: demoImg("1549652127-2e5e59e86a7a"),
			Variants:    one("CN-COLD"),
			CreatedAt:   at(2),
		},
		{
			ID: "steel-dripper", Name: "Steel Dripper", Category: "equipment",
			Description: "A single-piece stainless pour-over cone. No paper filter, no waste.",
			PriceCents:  3800, Currency: "USD", ImageURL: demoImg("1739867426623-4d79363cdcbc"),
			Variants:    one("CN-DRIP"),
			CreatedAt:   at(3),
		},
		{
			ID: "mono-mug", Name: "Mono Mug", Category: "merch",
			Description: "Matte black stoneware mug, 300 ml. Heavy base, severe handle.",
			PriceCents:  2200, Currency: "USD", ImageURL: demoImg("1520031473529-2c06dea61853"),
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
