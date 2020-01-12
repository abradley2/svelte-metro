import App from './App.svelte';

const dev = window.location.hostname.includes('localhost')

const server = dev
	? `${window.location.hostname}:9966`
	: window.location.hostname

const API_URL = `${window.location.protocol}//${server}/api`

const WS_URL = `ws://${server}/ws`

const app = new App({
	target: document.body,
	props: {
		API_URL,
		WS_URL
	}
});

export default app;