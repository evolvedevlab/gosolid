import { build as esbuild } from "esbuild";
import { solidPlugin } from "esbuild-plugin-solid";
import { copyFile } from "fs/promises";

const run = () => {
  esbuild({
    entryPoints: ["web/pages/*"],
    bundle: true,
    outdir: "dist/server",
    metafile: true,
    sourcemap: true,
    plugins: [
      solidPlugin({
        solid: {
          generate: "ssr",
          hydratable: true,
        },
      }),
    ],
    // Required configuration for proper development/production handling
    define: {
      "process.env.ENVIRONMENT": JSON.stringify(process.env.ENVIRONMENT),
      __SERVER__: "true",
    },
    conditions: [process.env.ENVIRONMENT],
    platform: "node",
    format: "iife",
  })
    .then((result) => {})
    .catch((err) => {
      console.log("server build error: ", err);
      process.exit(1);
    });

  esbuild({
    entryPoints: ["web/pages/*"],
    bundle: true,
    outdir: "dist/client",
    metafile: true,
    sourcemap: true,
    // splitting: true,
    plugins: [
      solidPlugin({
        solid: {
          generate: "dom",
          hydratable: true,
        },
      }),
    ],
    // Required configuration for proper development/production handling
    define: {
      "process.env.ENVIRONMENT": JSON.stringify(process.env.ENVIRONMENT),
      __SERVER__: "false",
    },
    conditions: [process.env.ENVIRONMENT],
    platform: "browser",
    format: "esm",
  })
    .then((result) => {})
    .catch((err) => {
      console.log("client build error: ", err);
      process.exit(1);
    });

  copyFile("web/index.html", "dist/index.html");
};

Promise.all([run()]);
