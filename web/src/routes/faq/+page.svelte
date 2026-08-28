<script lang="ts">
	// Platform FAQ (marketing). Same accordion pattern as the storefront's FAQ section
	// (StorefrontSections.svelte) — native <details>/<summary>, no JS needed for the toggle.
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { ChevronDown } from 'lucide-svelte';
	import { fly } from 'svelte/transition';

	const questions = ['q1', 'q2', 'q3', 'q4', 'q5', 'q6', 'q7', 'q8'];

	// Static ES copy for the FAQPage schema — JSON-LD must match visible content;
	// mirroring the live i18n store here would need SSR-time locale resolution we don't have yet.
	const faqSchema = {
		'@context': 'https://schema.org',
		'@type': 'FAQPage',
		mainEntity: [
			{
				'@type': 'Question',
				name: '¿De verdad hay un plan gratis?',
				acceptedAnswer: { '@type': 'Answer', text: 'Sí — una tienda, storefront completo y el motor de no-sobreventa incluido, sin tarjeta de crédito. Pro y Scale suman logo propio, caché prioritaria, analytics, drops programados y límites más altos.' }
			},
			{
				'@type': 'Question',
				name: '¿Puedo usar mi propio dominio?',
				acceptedAnswer: { '@type': 'Answer', text: 'Cada tienda arranca con una URL de TuTienda en /store/tu-slug. Los dominios propios están en nuestro roadmap.' }
			},
			{
				'@type': 'Question',
				name: '¿Qué evita realmente la sobreventa?',
				acceptedAnswer: { '@type': 'Answer', text: 'Una operación atómica en la base de datos en cada compra: solo tiene éxito si todavía hay stock, acotada a tu propia tienda. Sin locks distribuidos, sin condición de carrera.' }
			},
			{
				'@type': 'Question',
				name: '¿Puedo migrar mi catálogo actual?',
				acceptedAnswer: { '@type': 'Answer', text: 'Podés cargar productos y stock directo desde el panel del comerciante apenas te registrás, uno por uno o en bloque con una importación CSV.' }
			},
			{
				'@type': 'Question',
				name: '¿Qué pasa con mis datos si cierro la cuenta?',
				acceptedAnswer: { '@type': 'Answer', text: 'Borrar una tienda elimina sus productos y stock de forma inmediata y permanente. Es una acción de un clic desde tu panel.' }
			},
			{
				'@type': 'Question',
				name: '¿Se quedan con un porcentaje de mis ventas?',
				acceptedAnswer: { '@type': 'Answer', text: 'No. Los planes son una tarifa mensual fija según funciones y límites, nunca un porcentaje de lo que vendés.' }
			},
			{
				'@type': 'Question',
				name: '¿Puedo hacer un drop limitado/programado?',
				acceptedAnswer: { '@type': 'Answer', text: 'Sí, en Pro y Scale — programás un horario de lanzamiento, el storefront muestra la cuenta regresiva, y la compra queda bloqueada del lado del servidor hasta que se libera.' }
			},
			{
				'@type': 'Question',
				name: '¿Los datos de mis clientes están seguros?',
				acceptedAnswer: { '@type': 'Answer', text: 'Los datos de cada tienda están aislados por tenant y nunca se comparten entre tiendas.' }
			}
		]
	};
</script>

<svelte:head>
	<title>{BRAND} — {$t('marketing.faq.title')}</title>
	{@html `<script type="application/ld+json">${JSON.stringify(faqSchema)}</` + `script>`}
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<section class="mx-auto w-full max-w-container px-4 pb-12 pt-20 md:px-12 md:pt-32">
			<div in:fly={{ y: 16, duration: 400 }}>
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.faq.kicker')}</p>
				<h1 class="mt-4 max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg md:text-display-xl">
					{$t('marketing.faq.title')}
				</h1>
				<p class="mt-6 max-w-2xl font-sans text-body-lg text-text-muted">{$t('marketing.faq.subtitle')}</p>
			</div>
		</section>

		<section class="border-t border-border">
			<div class="mx-auto max-w-3xl px-4 py-12 md:px-12">
				<div class="flex flex-col divide-y divide-border border-b border-border">
					{#each questions as key, i (key)}
						<div in:fly={{ y: 12, duration: 300, delay: i * 40 }}>
							<details class="group py-6">
								<summary class="flex cursor-pointer list-none items-center justify-between gap-4 font-sans text-body-lg text-text">
									{$t('marketing.faq.' + key + 'Title', { brand: BRAND })}
									<ChevronDown size={18} class="shrink-0 text-text-muted transition-transform group-open:rotate-180" />
								</summary>
								<p class="mt-4 font-sans text-body-md leading-relaxed text-text-muted">
									{$t('marketing.faq.' + key + 'Body', { brand: BRAND })}
								</p>
							</details>
						</div>
					{/each}
				</div>
			</div>
		</section>
	</main>

	<SiteFooter />
</div>
