import { preview } from "vite";

const server = await preview({
  root: new URL("..", import.meta.url).pathname,
  preview: { host: "127.0.0.1", port: 4173, strictPort: true },
});

try {
  const origin = server.resolvedUrls?.local[0];
  if (!origin) throw new Error("Preview server did not provide a local URL.");
  const response = await fetch(
    new URL("/vehicles/fixture:vehicle:2901", origin),
  );
  if (!response.ok)
    throw new Error(`Preview route returned ${response.status}.`);
  const html = await response.text();
  if (!html.includes('id="root"'))
    throw new Error("Preview did not return the SPA document.");
} finally {
  await new Promise((resolve, reject) =>
    server.httpServer.close((error) => (error ? reject(error) : resolve())),
  );
}
