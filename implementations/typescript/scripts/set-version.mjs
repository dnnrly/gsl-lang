import { readFileSync, writeFileSync } from "node:fs";

const version = process.env.VERSION;
if (!version) {
  console.log("VERSION not set — leaving package.json versions as-is");
  process.exit(0);
}
if (!/^\d+\.\d+\.\d+([-+].*)?$/.test(version)) {
  console.error(`VERSION '${version}' is not valid semver`);
  process.exit(1);
}
// Workspace packages published from this workspace, in dependency order.
const packages = ["core", "gql", "dialect"];

// Internal dependency ranges rewritten alongside every version stamp: each
// package is rewritten for every internal dependency it declares, so a
// package that depends on both @gsl-lang/core and @gsl-lang/gql is fully
// restamped.
const internalDeps = ["@gsl-lang/core", "@gsl-lang/gql"];

for (const pkg of packages) {
  const file = `packages/${pkg}/package.json`;
  const manifest = JSON.parse(readFileSync(file, "utf8"));
  manifest.version = version;
  for (const dep of internalDeps) {
    if (manifest.dependencies && manifest.dependencies[dep]) {
      manifest.dependencies[dep] = `^${version}`;
    }
  }
  writeFileSync(file, JSON.stringify(manifest, null, 2) + "\n");
}
console.log(`Stamped packages to ${version}`);
