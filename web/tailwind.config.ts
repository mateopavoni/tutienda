import type { Config } from 'tailwindcss';

// Tokens come from the Stitch "Refined Brutalism" design system (docs/design/design-tokens.md).
// Semantic colours are CSS custom properties (RGB triplets) so light/dark swap by toggling .dark,
// with no component hardcoding a colour. Zero radius everywhere enforces the hard-edge language.
export default {
	darkMode: 'class',
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				bg: 'rgb(var(--bg) / <alpha-value>)',
				surface: 'rgb(var(--surface) / <alpha-value>)',
				'surface-variant': 'rgb(var(--surface-variant) / <alpha-value>)',
				border: 'rgb(var(--border) / <alpha-value>)',
				text: 'rgb(var(--text) / <alpha-value>)',
				'text-muted': 'rgb(var(--text-muted) / <alpha-value>)',
				outline: 'rgb(var(--outline) / <alpha-value>)',
				primary: 'rgb(var(--primary) / <alpha-value>)',
				'on-primary': 'rgb(var(--on-primary) / <alpha-value>)',
				accent: 'rgb(var(--accent) / <alpha-value>)',
				'on-accent': 'rgb(var(--on-accent) / <alpha-value>)',
				error: 'rgb(var(--error) / <alpha-value>)',
				inverse: 'rgb(var(--inverse) / <alpha-value>)',
				'on-inverse': 'rgb(var(--on-inverse) / <alpha-value>)'
			},
			fontFamily: {
				sans: ['Inter', 'system-ui', 'sans-serif'],
				mono: ['"JetBrains Mono"', 'ui-monospace', 'monospace']
			},
			fontSize: {
				'display-xl': ['clamp(3.5rem, 12vw, 7.5rem)', { lineHeight: '0.9', letterSpacing: '-0.04em', fontWeight: '800' }],
				'headline-lg': ['clamp(2.5rem, 6vw, 4rem)', { lineHeight: '1', letterSpacing: '-0.02em', fontWeight: '700' }],
				'headline-md': ['2rem', { lineHeight: '1.125', letterSpacing: '-0.01em', fontWeight: '600' }],
				'body-lg': ['1.125rem', { lineHeight: '1.55', fontWeight: '400' }],
				'body-md': ['1rem', { lineHeight: '1.5', fontWeight: '400' }],
				'metadata-sm': ['0.75rem', { lineHeight: '1.33', letterSpacing: '0.05em', fontWeight: '500' }],
				'label-caps': ['0.6875rem', { lineHeight: '1.1', letterSpacing: '0.1em', fontWeight: '700' }]
			},
			borderRadius: {
				none: '0',
				DEFAULT: '0',
				sm: '0',
				md: '0',
				lg: '0',
				xl: '0',
				full: '0'
			},
			maxWidth: {
				container: '1440px'
			}
		}
	},
	plugins: []
} satisfies Config;
