import { build } from "esbuild";
import { solidPlugin } from "esbuild-plugin-solid";
import { copyFile, writeFile } from "fs/promises";

const isDev = process.env.ENVIRONMENT === "development";

const shared = {
  entryPoints: ["web/pages/*"],
  bundle: true,
  metafile: true,
  sourcemap: isDev ? "inline" : false,
  minify: !isDev,
  treeShaking: true,
  conditions: [process.env.ENVIRONMENT],
  logLevel: "info",
  define: {
    "process.env.ENVIRONMENT": JSON.stringify(process.env.ENVIRONMENT),
  },
};

const createConfig = ({ server, outdir, generate, format, splitting = false }) => ({
  ...shared,
  outdir,
  format,
  platform: server ? "node" : "browser",
  target: server ? ["node20"] : ["chrome120", "firefox120", "safari17"],
  splitting,
  entryNames: server ? "[name]" : isDev ? "[name]" : "[name]-[hash]",
  chunkNames: server ? "chunks/[name]" : isDev ? "chunks/[name]" : "chunks/[name]-[hash]",
  assetNames: isDev ? "assets/[name]" : "assets/[name]-[hash]",
  plugins: [
    solidPlugin({
      solid: {
        generate,
        hydratable: true,
      },
    }),
  ],
  define: {
    ...shared.define,
    __SERVER__: server ? "true" : "false",
  },
});
const run = async () => {
  try {
    const [server, client] = await Promise.all([
      build(
        createConfig({
          server: true,
          outdir: "dist/server",
          generate: "ssr",
          format: "iife",
        }),
      ),
      build(
        createConfig({
          server: false,
          outdir: "dist/client",
          generate: "dom",
          format: "esm",
          splitting: true,
        }),
      ),
    ]);

    const data = {
      server: server.metafile,
      client: client.metafile,
    };
    await Promise.all([
      writeFile("dist/metafile.json", JSON.stringify(data, null, 2)),
      copyFile("web/index.html", "dist/index.html"),
    ]);

    console.log(`build complete (${isDev ? "dev" : "prod"})`);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
};

run();
