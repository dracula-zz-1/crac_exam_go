import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import wails from "@wailsio/runtime/plugins/vite";
import { resolve } from "path";

export default defineConfig({
  plugins: [vue(), wails("./bindings")],
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
      "@bindings": resolve(__dirname, "bindings"),
    },
  },
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  build: {
    target: "es2020",
    chunkSizeWarningLimit: 900,
  },
});
