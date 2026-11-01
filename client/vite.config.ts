import { readFileSync } from "node:fs";
import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { vanillaExtractPlugin } from "@vanilla-extract/vite-plugin";

const devApiTarget =
  process.env.KOITO_DEV_API_TARGET || "http://127.0.0.1:4110";

const isDocker = process.env.BUILD_TARGET === "docker";

export default defineConfig({
  define: {
    "import.meta.env.VITE_KOITO_VERSION": JSON.stringify(
      process.env.VITE_KOITO_VERSION ||
        readFileSync(
          new URL("../internal/version/VERSION", import.meta.url),
          "utf8",
        ).trim(),
    ),
  },
  plugins: [tailwindcss(), reactRouter(), vanillaExtractPlugin()],
  server: {
    host: "localhost",
    port: 5173,
    strictPort: true,
    proxy: {
      "/apis": {
        target: devApiTarget,
        changeOrigin: false,
      },
      "/image": {
        target: devApiTarget,
        changeOrigin: false,
      },
    },
  },
  resolve: {
    tsconfigPaths: true,
    alias: {
      ...(isDocker ? { "react-dom/server": "react-dom/server.node" } : {}),
    },
  },
});
