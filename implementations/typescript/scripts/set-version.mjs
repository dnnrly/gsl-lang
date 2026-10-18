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
for (const pkg of ["core", "gql"]) {
  const file = `packages/${pkg}/package.json`;
  const manifest = JSON.parse(readFileSync(file, "utf8"));
  manifest.version = version;
  if (manifest.dependencies && manifest.dependencies["@gsl-lang/core"]) {
    manifest.dependencies["@gsl-lang/core"] = `^${version}`;
  }
  writeFileSync(file, JSON.stringify(manifest, null, 2) + "\n");
}
console.log(`Stamped packages to ${version}`);
