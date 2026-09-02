import { defineConfig } from "astro/config";
import node from "@astrojs/node";
import preact from "@astrojs/preact";
export default defineConfig({ output: "server", adapter: node({ mode: "standalone" }), integrations: [preact()], compressHTML: false });
