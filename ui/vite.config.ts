import GoCrudVitePlugin from "@allape/gocrud-react/vite-plugin";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), GoCrudVitePlugin(), viteSingleFile()],
});
