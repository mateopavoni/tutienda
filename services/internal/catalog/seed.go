package catalog

import "time"

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
