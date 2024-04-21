import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';
import { viteSingleFile } from 'vite-plugin-singlefile';

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [
		svelte(),
		// cssInjectedByJsPlugin(),
		viteSingleFile()
	],
	build: {
		// rollupOptions: {
		//   output: {
		//     format: 'commonjs',
		//     entryFileNames: 'app.js',
		//     manualChunks: undefined,
		//   },
		// },
	}
});
