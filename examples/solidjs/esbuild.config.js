import { build, context } from "esbuild";
import { solidPlugin } from "esbuild-plugin-solid";
import { copyFile, mkdir, writeFile, rm } from "fs/promises";

const isDev = process.env.ENVIRONMENT === "development";

const shared = {
  entryPoints: ["web/pages/*"],
  bundle: true,
  metafile: true,
  sourcemap: true,
  minify: !isDev,
  treeShaking: true,
  conditions: [process.env.ENVIRONMENT],
  logLevel: "info",
  define: {
    "process.env.ENVIRONMENT": JSON.stringify(process.env.ENVIRONMENT),
  },
};

const createConfig = ({
  server,
  outdir,
  generate,
  format,
  sourcemap = true,
  splitting = false,
}) => ({
  ...shared,
  outdir,
  format,
  platform: server ? "node" : "browser",
  target: server ? ["node20"] : ["chrome120", "firefox120", "safari17"],
  splitting,
  sourcemap,
  entryNames: server ? "[name]" : isDev ? "[name]" : "[name]-[hash]",
  chunkNames: server ? "chunks/[name]" : "chunks/[name]-[hash]",
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
    await rm("dist", { recursive: true, force: true });
    await mkdir("dist", { recursive: true });

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
          sourcemap: isDev,
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
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
};

// TODO: not needed for now
const watch = async () => {
  try {
    const serverCtx = await context(
      createConfig({
        server: true,
        outdir: "dist/server",
        generate: "ssr",
        format: "iife",
      }),
    );

    const clientCtx = await context(
      createConfig({
        server: false,
        outdir: "dist/client",
        generate: "dom",
        format: "esm",
        splitting: true,
        sourcemap: isDev,
      }),
    );

    await mkdir("./dist", { recursive: true });
    await Promise.all([serverCtx.watch(), clientCtx.watch()]);
    await copyFile("web/index.html", "dist/index.html");
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
};

if (isDev) {
  run();
} else {
  run();
}
