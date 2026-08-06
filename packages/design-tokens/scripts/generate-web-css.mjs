import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptsDirectory = dirname(fileURLToPath(import.meta.url));
const packageDirectory = resolve(scriptsDirectory, "..");
const outputPath = resolve(packageDirectory, "dist", "web.css");
const { webCss } = await import(resolve(packageDirectory, "dist", "web.js"));

await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, webCss, "utf8");
