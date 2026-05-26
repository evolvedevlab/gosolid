import { execSync } from "child_process";
import { build } from "esbuild";
import { solidPlugin } from "esbuild-plugin-solid";
import { copyFile, mkdir, writeFile, rm } from "fs/promises";

// Example json data from Go:
// { goConfig: {...}, "server": {...}, "client": {...} }
const cfg = JSON.parse(
  execSync("go run . -- -gowebi-esbuild-config", {
    encoding: "utf8",
  }),
);

const { bundleDir } = cfg.goConfig;
const run = async () => {
  try {
    // init cleanup
    await rm(bundleDir, { recursive: true, force: true });
    await mkdir(bundleDir, { recursive: true });

    // build
    const [server, client] = await Promise.all([
      build({
        ...cfg.server,
        plugins: [
          solidPlugin({
            solid: {
              generate: "ssr",
              hydratable: true,
            },
          }),
        ],
      }),
      build({
        ...cfg.client,
        plugins: [
          solidPlugin({
            solid: {
              generate: "dom",
              hydratable: true,
            },
          }),
        ],
      }),
    ]);

    // end
    await Promise.all([
      writeFile(bundleDir + "/metafile.server.json", JSON.stringify(server.metafile)),
      writeFile(bundleDir + "/metafile.client.json", JSON.stringify(client.metafile)),
      copyFile("web/index.html", bundleDir + "/index.html"),
    ]);
  } catch (err) {
    console.error(err);
    process.exit(1);
  }
};

run();
