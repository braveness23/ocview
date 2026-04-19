import type { BunPlugin } from "bun";

const stubDevtools: BunPlugin = {
  name: "stub-react-devtools-core",
  setup(build) {
    build.onResolve({ filter: /^react-devtools-core$/ }, () => ({
      path: "react-devtools-core-stub",
      namespace: "stub",
    }));
    build.onLoad({ filter: /.*/, namespace: "stub" }, () => ({
      contents: "export default { connectToDevTools: () => {} };",
      loader: "js",
    }));
  },
};

const result = await Bun.build({
  entrypoints: ["src/index.tsx"],
  outdir: "dist",
  naming: "index.js",
  target: "bun",
  plugins: [stubDevtools],
});

if (!result.success) {
  for (const log of result.logs) {
    console.error(log);
  }
  process.exit(1);
}

// Fix shebang so the binary works when bun is not on the system PATH
const { readFileSync, writeFileSync } = await import('fs');
const outFile = 'dist/index.js';
const built = readFileSync(outFile, 'utf-8');
const fixed = built.replace(/^#!.*/, `#!/usr/bin/env -S ${process.execPath}`);
writeFileSync(outFile, fixed);
