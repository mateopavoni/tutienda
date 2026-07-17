<script lang="ts">
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { browser } from '$app/environment';
	import { t } from '$lib/i18n';
	import { session } from '$lib/admin/session';
	import {
		listStores,
		createStore,
		selectStore,
		setStoreDisabled,
		deleteStore,
		updateStore,
		changePlan,
		listProducts,
		createProduct,
		updateProduct,
		deleteProduct,
		setStock,
		listStock,
		uploadImage,
		type Store
	} from '$lib/admin/api';
	import type { ApiError, Product } from '$lib/types';
	import { PLANS, TIERS, hasFeature, productLimit, normalizeTier, UNLIMITED } from '$lib/plan';
	import { THEMES, normalizeTheme } from '$lib/theme';
	import { theme } from '$lib/stores/theme';
	import {
		normalizeLayout,
		enabledSections,
		SECTION_META,
		HERO_OVERLAY_MIN,
		HERO_OVERLAY_MAX,
		HERO_OVERLAY_DEFAULT,
		type Section,
		type SectionType,
		type ContentField
	} from '$lib/layout';
	import { PAGE_TYPES, PAGE_META, enabledPages, type Page, type PageType } from '$lib/pages';
	import StorefrontSections from '$lib/components/StorefrontSections.svelte';
	import { safeAccent, safeTitleColor } from '$lib/colorGuard';
	import { ExternalLink, Lock, ArrowUp, ArrowDown, X } from 'lucide-svelte';

	// Mirrors the server's MAX_STORES_PER_MERCHANT default (anti-abuse cap). The server is authoritative
	// (409 on create); this only gates the UI so the button disables before the round-trip.
	const MAX_STORES = 2;

	let stores = $state<Store[]>([]);
	let active = $state(session.activeStore());
	let products = $state<Product[]>([]);
	let error = $state('');
	let notice = $state('');
	// True until the initial onMount fetch settles — the dashboard is pure client-side (ssr disabled),
	// so without this the store grid renders empty for a beat instead of showing it's loading.
	let loading = $state(true);

	// The active store's full record (with its plan) lives in `stores`; `active` only holds id/slug/name.
	const activeStore = $derived(stores.find((s) => s.id === active?.id));
	const tier = $derived(normalizeTier(activeStore?.plan));
	const canDrops = $derived(hasFeature(tier, 'drops'));
	const canLogo = $derived(hasFeature(tier, 'customLogo'));
	const canSections = $derived(hasFeature(tier, 'sectionsPro'));
	const canPages = $derived(hasFeature(tier, 'staticPages'));
	// Mirrors proSectionTypes in services/internal/accounts/service.go — the accounts write-time
	// sanitizer is the real gate; this only dims the controls before a merchant tries and gets stripped.
	const PRO_SECTION_TYPES = new Set(['feature', 'contact', 'gallery', 'faq']);
	const limit = $derived(productLimit(tier));
	const atProductLimit = $derived(limit !== UNLIMITED && products.length >= limit);

	// New-store form
	let newSlug = $state('');
	let newName = $state('');

	// New-product form
	let pName = $state('');
	let pCategory = $state('apparel');
	let pPrice = $state(0);
	let pImages = $state<string[]>([]); // gallery for the new product; first is the primary image
	// One or more sellable variants (size/color); each carries its own SKU and starting stock. Inventory
	// keys on SKU, so blank/duplicate SKUs are filtered on submit (and again server-side in normalizeVariants).
	let pVariants = $state<{ label: string; sku: string; qty: number }[]>([{ label: '', sku: '', qty: 0 }]);
	// What the variant rows above actually mean for this product (Size/Color/Model/Storage/custom/none) —
	// sent as Product.variantLabel so the storefront selector says "Storage" instead of a hardcoded "Size".
	// "unico" collapses the editor to a single implicit variant with no selector on the storefront.
	let pVariantType = $state<'talle' | 'color' | 'modelo' | 'almacenamiento' | 'personalizado' | 'unico'>('talle');
	let pVariantLabelCustom = $state(''); // free text, only used when pVariantType === 'personalizado'
	$effect(() => {
		if (pVariantType === 'unico' && pVariants.length !== 1) pVariants = [{ label: '', sku: '', qty: 0 }];
	});
	let pDrop = $state(''); // datetime-local; empty = on sale now, a future value schedules a drop
	let uploading = $state(false); // true while a product image is being uploaded to the object store
	// Non-null while the form above is editing an existing product instead of creating a new one — the
	// form/fields are fully shared, only the submit target (create vs. update) and its reset differ.
	let editingId = $state<string | null>(null);

	// Current on-hand per SKU across the active store's products, for the inline per-variant stock editor.
	let stockMap = $state<Record<string, number>>({});

	// Settings form
	let setName = $state('');
	let setAccent = $state('#5433eb');
	let setTitleColor = $state(''); // blank = design system default (text-text), same as accent's blank case
	let setTheme = $state('monolith');
	let setLogo = $state('');
	// Storefront layout: the normalized, full list of sections (every known block exactly once, in order).
	// The merchant toggles, reorders and edits copy here; only enabled sections render on the storefront.
	let setLayout = $state<Section[]>(normalizeLayout());
	// Standalone pages, one editable row per known type (blank ones are dropped on save / server-side).
	let setPages = $state<Record<PageType, { slug: string; title: string; body: string }>>(emptyPages());

	// Live preview of the layout editor: the same component the storefront renders, fed straight from the
	// unsaved `setLayout` state, so a merchant sees a change before hitting save. Real store slug/products
	// so the preview looks (and links) like the real thing.
	const previewBase = $derived('/store/' + (activeStore?.slug ?? ''));
	const previewSections = $derived(enabledSections(setLayout));
	const previewFeatured = $derived(products.slice(0, 6));
	// The preview opens in a full-size drawer (see the modal at the bottom of this file) instead of a
	// cramped inline box — a merchant couldn't judge how a hero/layout actually looks in a 70vh strip.
	let previewOpen = $state(false);
	// Resolved light/dark scheme, same pattern as store/[slug]/+layout.svelte, so the preview's dark:
	// classes (the hero overlay, etc.) match whatever mode is actually on screen.
	let prefersDark = $state(false);
	$effect(() => {
		if (!browser) return;
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const update = () => (prefersDark = mq.matches);
		update();
		mq.addEventListener('change', update);
		return () => mq.removeEventListener('change', update);
	});
	const previewIsDark = $derived($theme === 'dark' || ($theme === 'system' && prefersDark));
	// Fed live from the color pickers below (setAccent/setTitleColor), same contrast guard as the real
	// storefront — this is what fixes the preview always showing the design-system default purple/text
	// color regardless of what the merchant just picked.
	const previewAccent = $derived(safeAccent(setAccent, previewIsDark));
	const previewTitleColor = $derived(safeTitleColor(setTitleColor, previewIsDark));

	function emptyPages(): Record<PageType, { slug: string; title: string; body: string }> {
		return Object.fromEntries(PAGE_TYPES.map((t) => [t, { slug: '', title: '', body: '' }])) as Record<
			PageType,
			{ slug: string; title: string; body: string }
		>;
	}

	// loadStoreSettings hydrates the settings form (including layout + pages) from a store record.
	function loadStoreSettings(s: Store) {
		setName = s.displayName;
		setAccent = s.settings.accentColor ?? '#5433eb';
		setTitleColor = s.settings.titleColor ?? '';
		setTheme = normalizeTheme(s.settings.theme);
		setLogo = s.settings.logoUrl ?? '';
		setLayout = normalizeLayout(s.settings.layout);
		const p = emptyPages();
		for (const pg of s.settings.pages ?? []) {
			if (pg.type in p) p[pg.type] = { slug: pg.slug ?? '', title: pg.title ?? '', body: pg.body ?? '' };
		}
		// A store with no Terms yet starts from a generic ecommerce boilerplate (client-side only — nothing
		// is persisted until the merchant hits Save) instead of a blank textarea, so there's something
		// complete to edit down rather than write from scratch.
		if (!p.terms.body.trim()) p.terms.body = $t('app.settings.termsDefaultBody');
		setPages = p;
	}

	// Whether Terms is actually live on the storefront right now — from the last-saved store, not the
	// (possibly still-unsaved, possibly boilerplate-prefilled) editor state above.
	const termsPublished = $derived(enabledPages(activeStore?.settings?.pages).length > 0);

	// A section can move up/down within the list; the mandatory grid stays in the flow but can be reordered.
	function moveSection(i: number, dir: -1 | 1) {
		const j = i + dir;
		if (j < 0 || j >= setLayout.length) return;
		const next = [...setLayout];
		[next[i], next[j]] = [next[j], next[i]];
		setLayout = next;
	}

	function toggleSection(i: number) {
		const next = [...setLayout];
		next[i] = { ...next[i], enabled: !next[i].enabled };
		setLayout = next;
	}

	function editSection(i: number, field: ContentField, value: string) {
		const next = [...setLayout];
		next[i] = { ...next[i], [field]: value };
		setLayout = next;
	}

	// Hero-only: gradient peak strength (0-100). Tint color now follows the visitor's light/dark mode
	// automatically (see StorefrontSections.svelte), so it's no longer a merchant choice.
	function editSectionOverlay(i: number, overlay: number) {
		const next = [...setLayout];
		next[i] = { ...next[i], overlay };
		setLayout = next;
	}

	// Section image upload: same object-store flow as product images, but writes the returned URL into the
	// section's imageUrl field. uploadingSection holds the index being uploaded so only that row shows a
	// spinner and disables its picker.
	let uploadingSection = $state<number | null>(null);
	async function onPickSectionImage(i: number, e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		uploadingSection = i;
		try {
			editSection(i, 'imageUrl', await uploadImage(file));
			ok($t('app.toast.imageUploaded'));
		} catch (err) {
			fail(err);
		} finally {
			uploadingSection = null;
			input.value = '';
		}
	}

	// Logo upload: same object-store flow as product/section images, into setLogo instead of a pasted URL.
	let uploadingLogo = $state(false);
	async function onPickLogo(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		uploadingLogo = true;
		try {
			setLogo = await uploadImage(file);
			ok($t('app.toast.imageUploaded'));
		} catch (err) {
			fail(err);
		} finally {
			uploadingLogo = false;
			input.value = '';
		}
	}

	const fieldLabels = $derived<Record<ContentField, string>>({
		heading: $t('app.fields.heading'),
		body: $t('app.fields.body'),
		imageUrl: $t('app.fields.imageUrl'),
		ctaLabel: $t('app.fields.ctaLabel')
	});

	// --- Repeatable items (gallery images / FAQ Q&A) inside an item-based section ---
	type ItemField = 'heading' | 'body' | 'imageUrl';

	function editItem(si: number, ii: number, field: ItemField, value: string) {
		const next = [...setLayout];
		const items = [...(next[si].items ?? [])];
		items[ii] = { ...items[ii], [field]: value };
		next[si] = { ...next[si], items };
		setLayout = next;
	}
	function addItem(si: number) {
		const next = [...setLayout];
		next[si] = { ...next[si], items: [...(next[si].items ?? []), {}] };
		setLayout = next;
	}
	function removeItem(si: number, ii: number) {
		const next = [...setLayout];
		next[si] = { ...next[si], items: (next[si].items ?? []).filter((_, k) => k !== ii) };
		setLayout = next;
	}

	// Per-item image upload (gallery). uploadingItem holds "si:ii" so only that row shows the spinner.
	let uploadingItem = $state<string | null>(null);
	async function onPickItemImage(si: number, ii: number, e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		uploadingItem = `${si}:${ii}`;
		try {
			editItem(si, ii, 'imageUrl', await uploadImage(file));
			ok($t('app.toast.imageUploaded'));
		} catch (err) {
			fail(err);
		} finally {
			uploadingItem = null;
			input.value = '';
		}
	}

	// notice/error render in one banner at the top of the page; on a long form (settings/layout editor)
	// the user is scrolled well past it and never sees the confirmation. Auto-dismiss both so a stale
	// message doesn't linger past its relevance.
	let noticeTimer: ReturnType<typeof setTimeout> | undefined;
	function fail(err: unknown) {
		clearTimeout(noticeTimer);
		error = (err as ApiError)?.error ?? get(t)('app.toast.requestFailed');
		notice = '';
		noticeTimer = setTimeout(() => (error = ''), 5000);
	}
	function ok(msg: string) {
		clearTimeout(noticeTimer);
		notice = msg;
		error = '';
		noticeTimer = setTimeout(() => (notice = ''), 4000);
	}

	onMount(async () => {
		try {
			stores = await listStores();
		} catch (err) {
			fail(err);
		}
		try {
			if (active && session.storeToken()) {
				await refreshProducts();
				const s = stores.find((x) => x.id === active?.id);
				if (s) loadStoreSettings(s);
			}
		} finally {
			loading = false;
		}
	});

	async function refreshProducts() {
		try {
			products = await listProducts();
			await refreshStock();
		} catch (err) {
			fail(err);
		}
	}

	// refreshStock loads on-hand for every variant SKU across the store's products, for the inline editor.
	async function refreshStock() {
		const skus = products.flatMap((p) => p.variants.map((v) => v.sku));
		if (skus.length === 0) {
			stockMap = {};
			return;
		}
		try {
			const stocks = await listStock(skus);
			const m: Record<string, number> = {};
			for (const s of stocks) m[s.sku] = s.available;
			stockMap = m;
		} catch {
			/* non-fatal: the inline stock just shows as unknown */
		}
	}

	// doSetVariantStock saves a single variant's on-hand from the inline editor and reflects it locally.
	async function doSetVariantStock(sku: string, available: number) {
		try {
			await setStock(sku, available);
			stockMap = { ...stockMap, [sku]: available };
			ok($t('app.toast.stockSet', { sku }));
		} catch (err) {
			fail(err);
		}
	}

	async function doCreateStore() {
		try {
			const s = await createStore(newSlug, newName);
			stores = [...stores, s];
			newSlug = '';
			newName = '';
			ok($t('app.toast.storeCreated'));
		} catch (err) {
			fail(err);
		}
	}

	async function doSelectStore(store: Store) {
		try {
			await selectStore(store);
			active = session.activeStore();
			loadStoreSettings(store);
			await refreshProducts();
			ok($t('app.toast.managing', { name: store.displayName }));
		} catch (err) {
			fail(err);
		}
	}

	async function doToggleStoreStatus(store: Store) {
		try {
			const updated = await setStoreDisabled(store.id, !store.disabled);
			stores = stores.map((s) => (s.id === updated.id ? updated : s));
			ok(updated.disabled ? $t('app.toast.storeDisabled') : $t('app.toast.storeEnabled'));
		} catch (err) {
			fail(err);
		}
	}

	// Delete is destructive, so it goes through the in-page confirm dialog below instead of a native
	// browser confirm() — those get silently blocked/misbehave in some embedded/mobile contexts.
	let confirmDeleteTarget = $state<Store | null>(null);

	function doDeleteStore(store: Store) {
		if (!browser) return;
		confirmDeleteTarget = store;
	}

	async function confirmDeleteStore() {
		const store = confirmDeleteTarget;
		confirmDeleteTarget = null;
		if (!store) return;
		try {
			await deleteStore(store.id);
			stores = stores.filter((s) => s.id !== store.id);
			if (active?.id === store.id) {
				session.clearStore();
				active = null;
			}
			ok($t('app.toast.storeDeleted'));
		} catch (err) {
			fail(err);
		}
	}

	async function doChangePlan(toTier: string) {
		if (!activeStore) return;
		try {
			const updated = await changePlan(activeStore.id, toTier);
			stores = stores.map((s) => (s.id === updated.id ? updated : s));
			// The plan is signed into the store token, so re-mint it for the gateway to honor the new tier.
			await selectStore(updated);
			active = session.activeStore();
			ok($t('app.toast.planChanged', { plan: PLANS[normalizeTier(toTier)].label }));
		} catch (err) {
			fail(err);
		}
	}

	function addVariant() {
		pVariants = [...pVariants, { label: '', sku: '', qty: 0 }];
	}
	function removeVariant(i: number) {
		if (pVariants.length > 1) pVariants = pVariants.filter((_, k) => k !== i);
	}

	function resetProductForm() {
		pName = '';
		pPrice = 0;
		pImages = [];
		pVariants = [{ label: '', sku: '', qty: 0 }];
		pVariantType = 'talle';
		pVariantLabelCustom = '';
		pDrop = '';
		editingId = null;
	}

	// startEditProduct loads an existing product into the same create form (below), which switches to
	// update mode for as long as editingId is set. Reuses every field/snippet instead of a second form.
	function startEditProduct(p: Product) {
		editingId = p.id;
		pName = p.name;
		pCategory = p.category;
		pPrice = p.priceCents / 100;
		pImages = p.images?.length ? [...p.images] : p.imageUrl ? [p.imageUrl] : [];
		pVariants = p.variants.length
			? p.variants.map((v) => ({ label: v.label, sku: v.sku, qty: stockMap[v.sku] ?? 0 }))
			: [{ label: '', sku: '', qty: 0 }];
		const preset = (['talle', 'color', 'modelo', 'almacenamiento'] as const).find(
			(k) => $t('app.products.variantType.' + k) === p.variantLabel
		);
		if (preset) {
			pVariantType = preset;
			pVariantLabelCustom = '';
		} else if (!p.variantLabel && p.variants.length <= 1) {
			pVariantType = 'unico';
			pVariantLabelCustom = '';
		} else {
			pVariantType = 'personalizado';
			pVariantLabelCustom = p.variantLabel ?? '';
		}
		pDrop = p.dropAt ? p.dropAt.slice(0, 16) : '';
		if (browser) window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	function cancelEditProduct() {
		resetProductForm();
	}

	async function doSubmitProduct(e: SubmitEvent) {
		e.preventDefault();
		// Keep any row the merchant actually touched (label, SKU or a starting quantity) — the SKU itself
		// can be left blank: the server auto-generates one from the product name + label (normalizeVariants),
		// which is why we seed stock from the *returned* product below instead of the typed SKU.
		// "Único" always keeps exactly the one row the effect above forces — no label/SKU typed, both
		// auto-generate server-side — so it's exempt from the "did the merchant touch this row" filter.
		const rows =
			pVariantType === 'unico'
				? pVariants
				: pVariants.filter((v) => v.sku.trim() || v.label.trim() || v.qty > 0);
		const variantLabel =
			pVariantType === 'unico'
				? ''
				: pVariantType === 'personalizado'
					? pVariantLabelCustom.trim()
					: $t('app.products.variantType.' + pVariantType);
		const payload = {
			name: pName,
			category: pCategory,
			priceCents: Math.round(pPrice * 100),
			currency: 'USD',
			imageUrl: pImages[0] ?? '',
			images: pImages,
			variants: rows.map((v) => ({ sku: v.sku.trim(), label: v.label.trim() })),
			variantLabel,
			// A future drop date schedules a drop (storefront countdown + server-side checkout gate).
			dropAt: pDrop ? new Date(pDrop).toISOString() : undefined
		};
		try {
			if (editingId) {
				// Editing doesn't touch stock — that's the inline per-variant editor in the list below, and
				// re-running setStock here from whatever qty happened to be on screen at edit-open time
				// would silently clobber real inventory with a stale number.
				await updateProduct(editingId, payload);
				resetProductForm();
				await refreshProducts();
				ok($t('app.toast.productUpdated'));
				return;
			}
			const created = await createProduct(payload);
			// Seed starting stock per variant, paired positionally with what came back (inventory is a
			// separate service keyed by SKU, and normalizeVariants preserves row order).
			for (let i = 0; i < rows.length && i < created.variants.length; i++) {
				if (rows[i].qty > 0) await setStock(created.variants[i].sku, Math.max(0, Math.round(rows[i].qty)));
			}
			resetProductForm();
			await refreshProducts();
			ok($t('app.toast.productCreated'));
		} catch (err) {
			fail(err);
		}
	}

	// onPickProductImage uploads the chosen file(s) to the object store and appends the returned public
	// link(s) to the new product's gallery, so the merchant builds a multi-image product by uploading photos
	// instead of hosting and pasting URLs. Supports selecting several files at once.
	async function onPickProductImage(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const files = Array.from(input.files ?? []);
		if (files.length === 0) return;
		uploading = true;
		try {
			for (const file of files) {
				pImages = [...pImages, await uploadImage(file)];
			}
			ok(files.length > 1 ? $t('app.toast.imagesUploaded', { n: files.length }) : $t('app.toast.imageUploaded'));
		} catch (err) {
			fail(err);
		} finally {
			uploading = false;
			input.value = ''; // allow re-picking the same file after an error
		}
	}

	function removeProductImage(i: number) {
		pImages = pImages.filter((_, k) => k !== i);
	}

	async function doDeleteProduct(id: string) {
		try {
			await deleteProduct(id);
			products = products.filter((p) => p.id !== id);
			ok($t('app.toast.productDeleted'));
		} catch (err) {
			fail(err);
		}
	}

	async function doSaveSettings(e: SubmitEvent) {
		e.preventDefault();
		if (!active) return;
		try {
			// Only ship pages that have body content; the server drops empty ones too, this just keeps the payload tidy.
			const pages: Page[] = PAGE_TYPES.filter((t) => setPages[t].body.trim()).map((t) => ({
				type: t,
				slug: setPages[t].slug.trim(),
				title: setPages[t].title.trim(),
				body: setPages[t].body.trim()
			}));
			const updated = await updateStore(active.id, setName, {
				accentColor: setAccent,
				titleColor: setTitleColor,
				theme: setTheme,
				logoUrl: setLogo,
				currency: 'USD',
				layout: setLayout,
				pages
			});
			stores = stores.map((s) => (s.id === updated.id ? updated : s));
			loadStoreSettings(updated);
			ok($t('app.toast.settingsSaved'));
			if (browser) window.scrollTo({ top: 0, behavior: 'smooth' });
		} catch (err) {
			fail(err);
		}
	}

	const inputClass =
		'brutal-border min-w-0 bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';
</script>

<!-- Shared by the 3 image-upload spots below (product images, section image, item image) — same accept
     list/file-input styling/uploading-disabled behavior, only the change handler differs per caller. -->
{#snippet imageThumb(url: string, size: string = 'h-10 w-10')}
	<img src={url} alt="" class="brutal-border {size} object-cover" />
{/snippet}

{#snippet imageFileInput(onchange: (e: Event) => void, disabled: boolean, multiple: boolean = false)}
	<input
		type="file"
		accept="image/png,image/jpeg,image/webp,image/avif,image/gif"
		{multiple}
		{onchange}
		{disabled}
		class="mt-1 font-mono text-metadata-sm text-text-muted file:mr-3 file:brutal-border file:bg-surface file:px-3 file:py-1 file:font-mono file:text-metadata-sm file:text-text"
	/>
{/snippet}

<div class="mb-8 flex flex-wrap items-center justify-between gap-4">
	<div class="flex items-center gap-3">
		<h1 class="font-sans uppercase leading-none tracking-tighter text-headline-lg">{$t('app.title')}</h1>
		{#if activeStore}
			<span class="brutal-border bg-accent px-2 py-1 font-mono text-metadata-sm uppercase tracking-[0.1em] text-on-accent">
				{PLANS[tier].label}
			</span>
		{/if}
	</div>
	{#if active}
		<a href="/store/{active.slug}" target="_blank" rel="noopener" class="btn-ghost flex items-center gap-2">
			{$t('app.viewStore')}
			<ExternalLink size={14} />
		</a>
	{/if}
</div>

{#if error}
	<p class="mb-6 brutal-border border-error px-4 py-3 font-mono text-metadata-sm text-error">{error}</p>
{/if}
{#if notice}
	<p class="mb-6 brutal-border px-4 py-3 font-mono text-metadata-sm text-text-muted">{notice}</p>
{/if}

<!-- Stores -->
<section class="mb-12">
	<h2 class="mb-4 font-sans uppercase leading-none tracking-tight text-headline-md">{$t('app.stores.title')}</h2>
	<div class="grid gap-px brutal-border bg-border sm:grid-cols-2 lg:grid-cols-3">
		{#if loading}
			{#each { length: 3 } as _}
				<div class="flex flex-col gap-3 bg-bg p-4">
					<div class="skeleton-block h-5 w-2/3"></div>
					<div class="skeleton-block h-3 w-1/3"></div>
					<div class="skeleton-block mt-2 h-10 w-full"></div>
				</div>
			{/each}
		{:else}
			{#each stores as store (store.id)}
				<div class="flex flex-col gap-2 bg-bg p-4">
					<span class="flex items-center gap-2 font-sans text-body-lg">
						{store.displayName}
						{#if store.disabled}
							<span class="brutal-border px-2 py-0.5 font-mono text-metadata-sm uppercase text-error">
								{$t('app.stores.disabled')}
							</span>
						{/if}
					</span>
					<span class={labelClass}>{store.slug}</span>
					{#if active?.id === store.id}
						<span class="btn-accent mt-2 cursor-default text-center">{$t('app.stores.active')}</span>
					{:else}
						<button onclick={() => doSelectStore(store)} class="btn-ghost mt-2">
							{$t('app.stores.manage')}
						</button>
					{/if}
					<button onclick={() => doToggleStoreStatus(store)} class="btn-ghost">
						{store.disabled ? $t('app.stores.enable') : $t('app.stores.disable')}
					</button>
					<button onclick={() => doDeleteStore(store)} class="btn-ghost text-error">
						{$t('app.stores.delete')}
					</button>
				</div>
			{:else}
				<p class="bg-bg p-4 font-mono text-metadata-sm text-text-muted">{$t('app.stores.empty')}</p>
			{/each}
		{/if}
	</div>

	{#if stores.length >= MAX_STORES}
		<p class="mt-6 brutal-border px-4 py-3 font-mono text-metadata-sm text-text-muted">
			{$t('app.stores.limitReached', { max: String(MAX_STORES) })}
		</p>
	{:else}
		<div class="mt-6 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('app.stores.slug')}</span>
				<input bind:value={newSlug} placeholder="acme" class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('app.stores.displayName')}</span>
				<input bind:value={newName} placeholder="ACME" class={inputClass} />
			</label>
			<button onclick={doCreateStore} class="btn-primary">{$t('app.stores.create')}</button>
		</div>
	{/if}
</section>

{#if active && session.storeToken()}
	<!-- Membership -->
	<section class="mb-12">
		<h2 class="mb-4 font-sans uppercase leading-none tracking-tight text-headline-md">{$t('app.membership.title')}</h2>
		<div class="grid gap-px brutal-border bg-border md:grid-cols-3">
			{#each TIERS as pt (pt)}
				<div class="flex flex-col gap-3 bg-bg p-4 {pt === tier ? 'ring-2 ring-inset ring-accent' : ''}">
					<div class="flex items-baseline justify-between">
						<span class="font-sans text-body-lg">{PLANS[pt].label}</span>
						<span class="font-mono text-metadata-sm text-text-muted">{PLANS[pt].price}{$t('app.membership.perMonth')}</span>
					</div>
					<ul class="flex flex-col gap-1 font-mono text-metadata-sm text-text-muted">
						<li>{PLANS[pt].productLimit === UNLIMITED ? $t('app.membership.unlimited') : PLANS[pt].productLimit} {$t('app.membership.products')}</li>
						<li class={PLANS[pt].features.includes('drops') ? 'text-text' : ''}>
							{PLANS[pt].features.includes('drops') ? '✓' : '·'} {$t('app.membership.timedDrops')}
						</li>
						<li class={PLANS[pt].features.includes('customLogo') ? 'text-text' : ''}>
							{PLANS[pt].features.includes('customLogo') ? '✓' : '·'} {$t('app.membership.customLogo')}
						</li>
						<li class={PLANS[pt].features.includes('removeBranding') ? 'text-text' : ''}>
							{PLANS[pt].features.includes('removeBranding') ? '✓' : '·'} {$t('app.membership.removeBranding')}
						</li>
					</ul>
					{#if pt === tier}
						<span class="mt-auto font-mono text-metadata-sm uppercase tracking-[0.1em] text-accent">{$t('app.membership.current')}</span>
					{:else}
						<button onclick={() => doChangePlan(pt)} class="mt-auto {TIERS.indexOf(pt) > TIERS.indexOf(tier) ? 'btn-accent' : 'btn-ghost'}">
							{TIERS.indexOf(pt) > TIERS.indexOf(tier) ? $t('app.membership.upgrade') : $t('app.membership.downgrade')}
						</button>
					{/if}
				</div>
			{/each}
		</div>
	</section>

	<!-- Products -->
	<section class="mb-12">
		<h2 class="mb-4 font-sans uppercase leading-none tracking-tight text-headline-md">
			{$t('app.products.title')} — {active.displayName}
			<span class="ml-2 font-mono text-metadata-sm text-text-muted">
				{products.length}{limit === UNLIMITED ? '' : '/' + limit}
			</span>
		</h2>

		<div class="mb-6 flex flex-col gap-px brutal-border bg-border">
			{#each products as product (product.id)}
				<div class="flex flex-col gap-3 bg-bg p-4">
					<div class="flex items-start justify-between gap-4">
						<div class="flex items-center gap-3">
							{#if product.imageUrl}
								<img src={product.imageUrl} alt="" class="brutal-border h-10 w-10 object-cover" />
							{/if}
							<div class="flex flex-col">
								<span class="font-sans text-body-md">
									{product.name}
									{#if product.images && product.images.length > 1}
										<span class="ml-2 font-mono text-metadata-sm text-text-muted">{$t('app.products.imgs', { n: product.images.length })}</span>
									{/if}
								</span>
								<span class={labelClass}>
									{product.category} · {(product.priceCents / 100).toFixed(2)} {product.currency}
								</span>
							</div>
						</div>
						<div class="flex items-center gap-4">
							<button
								onclick={() => startEditProduct(product)}
								class="font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted hover:text-text"
							>
								{$t('app.products.edit')}
							</button>
							<button
								onclick={() => doDeleteProduct(product.id)}
								class="font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted hover:text-error"
							>
								{$t('app.products.delete')}
							</button>
						</div>
					</div>
					{#if product.variants.length}
						<div class="flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-border pt-3">
							<span class={labelClass}>{$t('app.products.stock')}</span>
							{#each product.variants as v (v.sku)}
								<label class="flex items-center gap-2">
									<span class="font-mono text-metadata-sm text-text-muted">{v.label || v.sku}</span>
									<input
										type="number"
										min="0"
										value={stockMap[v.sku] ?? 0}
										onchange={(e) =>
											doSetVariantStock(v.sku, Math.max(0, parseInt(e.currentTarget.value, 10) || 0))}
										class="{inputClass} w-20 py-1"
									/>
								</label>
							{/each}
						</div>
					{/if}
				</div>
			{:else}
				<p class="bg-bg p-4 font-mono text-metadata-sm text-text-muted">{$t('app.products.empty')}</p>
			{/each}
		</div>

		<form onsubmit={doSubmitProduct} class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('app.products.name')}</span>
				<input bind:value={pName} required class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('app.products.category')}</span>
				<input bind:value={pCategory} class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('app.products.price')}</span>
				<input type="number" min="0" step="0.01" bind:value={pPrice} class="{inputClass} w-32" />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('app.products.variantTypeLabel')}</span>
				<select bind:value={pVariantType} class={inputClass}>
					<option value="talle">{$t('app.products.variantType.talle')}</option>
					<option value="color">{$t('app.products.variantType.color')}</option>
					<option value="modelo">{$t('app.products.variantType.modelo')}</option>
					<option value="almacenamiento">{$t('app.products.variantType.almacenamiento')}</option>
					<option value="personalizado">{$t('app.products.variantTypeCustom')}</option>
					<option value="unico">{$t('app.products.variantTypeUnico')}</option>
				</select>
				{#if pVariantType === 'personalizado'}
					<input
						bind:value={pVariantLabelCustom}
						placeholder={$t('app.products.variantTypeCustomPh')}
						class="{inputClass} mt-1"
					/>
				{/if}
			</label>
			{#if pVariantType === 'unico'}
				<label class="flex flex-col gap-1">
					<span class={labelClass}>{$t('app.products.unicoQty')}</span>
					<input type="number" min="0" bind:value={pVariants[0].qty} class="{inputClass} w-32" />
				</label>
			{:else}
				<div class="flex w-full flex-col gap-2">
					<span class={labelClass}>{$t('app.products.variantsLabel')}</span>
					{#each pVariants as v, i (i)}
						<div class="flex flex-wrap items-end gap-2">
							<input bind:value={v.label} placeholder={$t('app.products.variantLabelPh')} class={inputClass} />
							<input bind:value={v.sku} placeholder={$t('app.products.variantSkuPh')} class={inputClass} />
							<input type="number" min="0" bind:value={v.qty} placeholder={$t('app.products.variantStockPh')} class="{inputClass} w-24" />
							{#if pVariants.length > 1}
								<button
									type="button"
									onclick={() => removeVariant(i)}
									aria-label={$t('app.products.removeVariant')}
									class="flex h-12 w-12 items-center justify-center brutal-border bg-bg font-mono text-text hover:text-error"
								>
									×
								</button>
							{/if}
						</div>
					{/each}
					<button type="button" onclick={addVariant} class="self-start font-mono text-metadata-sm text-accent hover:underline">
						{$t('app.products.addVariant')}
					</button>
				</div>
			{/if}
			<label class="flex flex-col gap-1">
				<span class="{labelClass} flex items-center gap-2">
					{$t('app.products.images')}
					{#if uploading}<span class="text-accent">{$t('app.products.uploading')}</span>{/if}
				</span>
				{#if pImages.length}
					<div class="grid grid-cols-4 gap-2 sm:grid-cols-6">
						{#each pImages as img, i (img)}
							<div class="relative">
								{@render imageThumb(img, 'h-12 w-12')}
								{#if i === 0}
									<span class="absolute -bottom-2 left-0 brutal-border bg-bg px-1 font-mono text-[0.6rem] uppercase text-text-muted">{$t('app.products.main')}</span>
								{/if}
								<button
									type="button"
									onclick={() => removeProductImage(i)}
									aria-label={$t('app.products.removeImage')}
									class="absolute -right-2 -top-2 flex h-5 w-5 items-center justify-center brutal-border bg-bg font-mono text-metadata-sm text-text hover:text-error"
								>
									×
								</button>
							</div>
						{/each}
					</div>
				{/if}
				{@render imageFileInput(onPickProductImage, uploading, true)}
				<span class="font-mono text-metadata-sm text-text-muted">{$t('app.products.imageHint')}</span>
			</label>
			<label class="flex flex-col gap-1">
				<span class="{labelClass} flex items-center gap-1">
					{$t('app.products.dropDate')} {#if !canDrops}<Lock size={11} /> {$t('app.pro')}{:else}{$t('app.products.optional')}{/if}
				</span>
				<input
					type="datetime-local"
					bind:value={pDrop}
					disabled={!canDrops}
					title={canDrops ? '' : $t('app.products.dropLocked')}
					class="{inputClass} disabled:cursor-not-allowed disabled:opacity-40"
				/>
			</label>
			<button
				type="submit"
				disabled={!editingId && atProductLimit}
				class="btn-primary disabled:cursor-not-allowed disabled:opacity-40"
			>
				{editingId ? $t('app.products.save') : $t('app.products.add')}
			</button>
			{#if editingId}
				<button type="button" onclick={cancelEditProduct} class="btn-ghost">
					{$t('app.products.cancelEdit')}
				</button>
			{/if}
		</form>
		{#if !editingId && atProductLimit}
			<p class="mt-3 font-mono text-metadata-sm text-text-muted">
				{$t('app.products.limitHit', { plan: PLANS[tier].label, limit })}
			</p>
		{/if}
	</section>

	<!-- Settings -->
	<section class="mb-12">
		<h2 class="mb-4 font-sans uppercase leading-none tracking-tight text-headline-md">{$t('app.settings.title')}</h2>
		<p class="mb-6 font-mono text-metadata-sm text-text-muted">{$t('app.settings.contentHint')}</p>
		<form onsubmit={doSaveSettings} class="flex flex-col gap-8">
			<div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
				<label class="flex flex-col gap-1">
					<span class={labelClass}>{$t('app.settings.displayName')}</span>
					<input bind:value={setName} class={inputClass} />
				</label>
				<label class="flex flex-col gap-1">
					<span class={labelClass}>{$t('app.settings.accent')}</span>
					<input type="color" bind:value={setAccent} class="brutal-border h-12 w-16 bg-surface" />
				</label>
				<label class="flex flex-col gap-1">
					<span class={labelClass}>{$t('app.settings.titleColor')}</span>
					<div class="flex items-center gap-2">
						<input
							type="color"
							value={setTitleColor || '#111111'}
							oninput={(e) => (setTitleColor = e.currentTarget.value)}
							class="brutal-border h-12 w-16 bg-surface"
						/>
						{#if setTitleColor}
							<button
								type="button"
								onclick={() => (setTitleColor = '')}
								class="font-mono text-metadata-sm text-text-muted hover:text-text hover:underline"
							>
								{$t('app.settings.titleColorReset')}
							</button>
						{/if}
					</div>
				</label>
				<label class="flex flex-col gap-1">
					<span class={labelClass}>{$t('app.settings.theme')}</span>
					<select bind:value={setTheme} class="{inputClass} h-12" title={THEMES.find((th) => th.id === setTheme)?.description}>
						{#each THEMES as th}
							<option value={th.id}>{th.label}</option>
						{/each}
					</select>
				</label>
				<label class="flex flex-col gap-1">
					<span class="{labelClass} flex items-center gap-1">
						{$t('app.settings.logoUrl')} {#if !canLogo}<Lock size={11} /> {$t('app.pro')}{/if}
						{#if uploadingLogo}<span class="text-accent">{$t('app.products.uploading')}</span>{/if}
					</span>
					<div class="flex items-center gap-3">
						{#if setLogo}{@render imageThumb(setLogo, 'h-12 w-12')}{/if}
						{@render imageFileInput(onPickLogo, !canLogo || uploadingLogo)}
					</div>
					<input
						bind:value={setLogo}
						disabled={!canLogo}
						title={canLogo ? '' : $t('app.settings.logoLocked')}
						placeholder={$t('app.settings.logoUrlPh')}
						class="{inputClass} mt-1 disabled:cursor-not-allowed disabled:opacity-40"
					/>
				</label>
			</div>

			<!-- Storefront layout: which sections appear, in what order, and their copy. The grid is fixed
			     (always on, can be reordered); the rest are optional. Chrome and design tokens never change. -->
			<div>
				<div class="mb-1 flex items-baseline gap-3">
					<h3 class="font-sans text-body-lg">{$t('app.settings.layoutTitle')}</h3>
					<span class="font-mono text-metadata-sm text-text-muted">{$t('app.settings.layoutHint')}</span>
				</div>
				<p class="mb-4 font-mono text-metadata-sm text-text-muted">
					{$t('app.settings.layoutDesc')}
				</p>
				<div class="flex flex-col gap-px brutal-border bg-border">
					{#each setLayout as section, i (section.type)}
						{@const meta = SECTION_META[section.type as SectionType]}
						{@const locked = PRO_SECTION_TYPES.has(section.type) && !canSections}
						<div class="flex flex-col gap-3 bg-bg p-4 {section.enabled && !locked ? '' : 'opacity-60'}">
							<div class="flex flex-wrap items-center gap-3">
								<div class="flex flex-col">
									<button
										type="button"
										onclick={() => moveSection(i, -1)}
										disabled={i === 0}
										aria-label={$t('app.settings.moveUp')}
										class="text-text-muted hover:text-text disabled:opacity-20"
									>
										<ArrowUp size={14} />
									</button>
									<button
										type="button"
										onclick={() => moveSection(i, 1)}
										disabled={i === setLayout.length - 1}
										aria-label={$t('app.settings.moveDown')}
										class="text-text-muted hover:text-text disabled:opacity-20"
									>
										<ArrowDown size={14} />
									</button>
								</div>
								<div class="flex flex-1 flex-col">
									<span class="font-sans text-body-md flex items-center gap-1">
										{meta.label} {#if locked}<Lock size={11} /> {$t('app.pro')}{/if}
									</span>
									<span class={labelClass}>{meta.hint}</span>
								</div>
								{#if meta.fixed}
									<span class="font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">{$t('app.settings.alwaysOn')}</span>
								{:else}
									<button
										type="button"
										onclick={() => toggleSection(i)}
										disabled={locked}
										title={locked ? $t('app.settings.sectionsLocked') : ''}
										class="font-mono text-metadata-sm uppercase tracking-[0.1em] disabled:cursor-not-allowed disabled:opacity-40 {section.enabled
											? 'text-accent'
											: 'text-text-muted hover:text-text'}"
									>
										{section.enabled ? $t('app.settings.shown') : $t('app.settings.hidden')}
									</button>
								{/if}
							</div>

							{#if section.enabled && !locked && meta.fields.length}
								<div class="flex flex-wrap gap-3 border-t border-border pt-3">
									{#each meta.fields as field (field)}
										<label class="flex flex-1 flex-col gap-1" style="min-width: 12rem">
											<span class="{labelClass} flex items-center gap-2">
												{fieldLabels[field] ?? field}
												{#if field === 'imageUrl' && uploadingSection === i}<span class="text-accent">{$t('app.products.uploading')}</span>{/if}
											</span>
											{#if field === 'body'}
												<textarea
													value={section[field] ?? ''}
													oninput={(e) => editSection(i, field, e.currentTarget.value)}
													rows="2"
													class={inputClass}
												></textarea>
											{:else if field === 'imageUrl'}
												<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
													{#if section.imageUrl}
														{@render imageThumb(section.imageUrl)}
													{/if}
													<input
														value={section[field] ?? ''}
														oninput={(e) => editSection(i, field, e.currentTarget.value)}
														placeholder={$t('app.settings.imageUrlPh')}
														class="{inputClass} flex-1"
													/>
												</div>
												{@render imageFileInput((e) => onPickSectionImage(i, e), uploadingSection === i)}
												<span class="font-mono text-metadata-sm text-text-muted">
													{section.type === 'hero' ? $t('app.settings.heroImageHint') : $t('app.settings.sectionImageHint')}
												</span>
											{:else}
												<input
													value={section[field] ?? ''}
													oninput={(e) => editSection(i, field, e.currentTarget.value)}
													class={inputClass}
												/>
											{/if}
										</label>
									{/each}
								</div>
							{/if}

							{#if section.enabled && meta.overlay && section.imageUrl}
								<!-- Overlay tint (white/black) now follows the visitor's light/dark mode automatically
								     (see StorefrontSections.svelte) — only the strength is still a merchant choice. -->
								<div class="flex flex-wrap items-center gap-4 border-t border-border pt-3">
									<label class="flex flex-1 flex-col gap-1" style="min-width: 12rem">
										<span class={labelClass}>{$t('app.fields.overlay')} ({section.overlay ?? HERO_OVERLAY_DEFAULT}%)</span>
										<input
											type="range"
											min={HERO_OVERLAY_MIN}
											max={HERO_OVERLAY_MAX}
											value={section.overlay ?? HERO_OVERLAY_DEFAULT}
											oninput={(e) => editSectionOverlay(i, Number(e.currentTarget.value))}
										/>
									</label>
								</div>
							{/if}

							{#if section.enabled && meta.item}
								{@const im = meta.item}
								<div class="flex flex-col gap-3 border-t border-border pt-3">
									<span class={labelClass}>{im.label}s</span>
									{#each section.items ?? [] as item, ii (ii)}
										<div class="flex flex-col items-stretch gap-3 brutal-border bg-surface p-3 sm:flex-row sm:flex-wrap sm:items-start">
											{#if im.kind === 'image'}
												<div class="flex items-center gap-3">
													{#if item.imageUrl}
														{@render imageThumb(item.imageUrl)}
													{/if}
													{@render imageFileInput((e) => onPickItemImage(i, ii, e), uploadingItem === i + ':' + ii)}
												</div>
												<input
													value={item.imageUrl ?? ''}
													oninput={(e) => editItem(i, ii, 'imageUrl', e.currentTarget.value)}
													placeholder={$t('app.items.imageUrl')}
													class="{inputClass} flex-1"
													style="min-width: 10rem"
												/>
												<input
													value={item.heading ?? ''}
													oninput={(e) => editItem(i, ii, 'heading', e.currentTarget.value)}
													placeholder={$t('app.items.caption')}
													class="{inputClass} w-44"
												/>
											{:else}
												<input
													value={item.heading ?? ''}
													oninput={(e) => editItem(i, ii, 'heading', e.currentTarget.value)}
													placeholder={$t('app.items.question')}
													class="{inputClass} flex-1"
													style="min-width: 10rem"
												/>
												<textarea
													value={item.body ?? ''}
													oninput={(e) => editItem(i, ii, 'body', e.currentTarget.value)}
													placeholder={$t('app.items.answer')}
													rows="2"
													class="{inputClass} flex-1"
													style="min-width: 10rem"
												></textarea>
											{/if}
											{#if uploadingItem === i + ':' + ii}
												<span class="font-mono text-metadata-sm text-accent">{$t('app.products.uploading')}</span>
											{/if}
											<button
												type="button"
												onclick={() => removeItem(i, ii)}
												class="font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted hover:text-error"
											>
												{$t('app.items.remove')}
											</button>
										</div>
									{/each}
									<button type="button" onclick={() => addItem(i)} class="btn-ghost self-start">
										+ Add {im.label.toLowerCase()}
									</button>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</div>

			<!-- Live preview: same markup the storefront renders, fed from the unsaved edits above. Opens in
			     a full-size drawer (below) instead of a cramped inline box. -->
			<div>
				<div class="mb-1 flex items-baseline gap-3">
					<h3 class="font-sans text-body-lg">{$t('app.settings.previewTitle')}</h3>
					<span class="font-mono text-metadata-sm text-text-muted">{$t('app.settings.previewHint')}</span>
				</div>
				<button type="button" onclick={() => (previewOpen = true)} class="btn-ghost">
					{$t('app.settings.previewOpen')}
				</button>
			</div>

			<!-- Standalone pages (just Terms — about/faq/contact live on the platform's own marketing site
			     now, not per-store): gets its own storefront route and footer link, but only once it has
			     content. Leave the body blank to keep it unpublished. -->
			<div>
				<div class="mb-1 flex items-baseline gap-3">
					<h3 class="font-sans text-body-lg">{$t('app.settings.pagesTitle')}</h3>
					<span class="font-mono text-metadata-sm text-text-muted">{$t('app.settings.pagesHint')}</span>
				</div>
				<p class="mb-4 font-mono text-metadata-sm text-text-muted">
					{$t('app.settings.pagesDesc')}
				</p>
				<div class="flex flex-col gap-px brutal-border bg-border">
					{#each PAGE_TYPES as type (type)}
						<div class="flex flex-col gap-3 bg-bg p-4 {canPages ? '' : 'opacity-60'}">
							<div class="flex items-baseline justify-between gap-3">
								<span class="font-sans text-body-md flex items-center gap-1">
									{PAGE_META[type].defaultTitle} {#if !canPages}<Lock size={11} /> {$t('app.pro')}{/if}
									<span
										class="ml-2 brutal-border px-2 py-0.5 font-mono text-metadata-sm uppercase {termsPublished
											? 'text-accent'
											: 'text-text-muted'}"
									>
										{termsPublished ? $t('app.settings.pagePublished') : $t('app.settings.pageUnpublished')}
									</span>
								</span>
								<span class={labelClass}>{PAGE_META[type].hint}</span>
							</div>
							<label class="flex flex-col gap-1">
								<span class={labelClass}>{$t('app.settings.pageTitleField')}</span>
								<input
									bind:value={setPages[type].title}
									placeholder={PAGE_META[type].defaultTitle}
									disabled={!canPages}
									title={canPages ? '' : $t('app.settings.pagesLocked')}
									class="{inputClass} disabled:cursor-not-allowed disabled:opacity-40"
								/>
							</label>
							<label class="flex flex-col gap-1">
								<span class={labelClass}>{$t('app.settings.pageSlugField')}</span>
								<input
									bind:value={setPages[type].slug}
									placeholder={$t('app.settings.pageSlugPh')}
									disabled={!canPages}
									title={canPages ? '' : $t('app.settings.pagesLocked')}
									class="{inputClass} disabled:cursor-not-allowed disabled:opacity-40"
								/>
							</label>
							<label class="flex flex-col gap-1">
								<span class={labelClass}>{$t('app.settings.pageContentField')}</span>
								<textarea
									bind:value={setPages[type].body}
									rows="4"
									disabled={!canPages}
									title={canPages ? '' : $t('app.settings.pagesLocked')}
									class="{inputClass} disabled:cursor-not-allowed disabled:opacity-40"
								></textarea>
							</label>
						</div>
					{/each}
				</div>
			</div>

			<button type="submit" class="btn-primary self-start">{$t('app.settings.save')}</button>
		</form>
	</section>
{/if}

{#if previewOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-2 sm:p-6"
		role="presentation"
		onclick={() => (previewOpen = false)}
		onkeydown={(e) => e.key === 'Escape' && (previewOpen = false)}
	>
		<div
			class="flex h-[95vh] w-[95vw] flex-col brutal-border bg-bg"
			role="dialog"
			tabindex="-1"
			aria-modal="true"
			aria-label={$t('app.settings.previewTitle')}
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<div class="flex items-center justify-between border-b border-border px-4 py-3">
				<span class="font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">
					{$t('app.settings.previewTitle')} — {$t('app.settings.previewHint')}
				</span>
				<button
					type="button"
					onclick={() => (previewOpen = false)}
					aria-label={$t('app.settings.previewClose')}
					class="p-2 text-text-muted hover:text-text"
				>
					<X size={18} strokeWidth={1.5} />
				</button>
			</div>
			<div
				class="flex-1 overflow-y-auto"
				style="{previewAccent ? `--accent: ${previewAccent};` : ''}{previewTitleColor
					? `--title-color: ${previewTitleColor};`
					: ''}"
			>
				<StorefrontSections
					sections={previewSections}
					featured={previewFeatured}
					base={previewBase}
					catalogHref={previewBase + '/catalogo'}
					storeName={setName}
				/>
			</div>
		</div>
	</div>
{/if}

{#if confirmDeleteTarget}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
		role="presentation"
		onclick={() => (confirmDeleteTarget = null)}
		onkeydown={(e) => e.key === 'Escape' && (confirmDeleteTarget = null)}
	>
		<div
			class="w-full max-w-sm brutal-border bg-bg p-6"
			role="alertdialog"
			tabindex="-1"
			aria-modal="true"
			aria-labelledby="delete-store-heading"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<h2 id="delete-store-heading" class="font-sans uppercase leading-none tracking-tight text-headline-md text-text">{$t('app.stores.delete')}</h2>
			<p class="mt-3 font-sans text-body-md text-text-muted">
				{$t('app.stores.deleteConfirm', { name: confirmDeleteTarget.displayName })}
			</p>
			<div class="mt-6 flex justify-end gap-3">
				<button type="button" class="btn-ghost" onclick={() => (confirmDeleteTarget = null)}>
					{$t('app.stores.cancel')}
				</button>
				<button type="button" class="btn-ghost text-error" onclick={confirmDeleteStore}>
					{$t('app.stores.delete')}
				</button>
			</div>
		</div>
	</div>
{/if}
