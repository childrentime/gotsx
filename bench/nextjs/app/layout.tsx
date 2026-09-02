import { readFileSync } from "node:fs";
import { join } from "node:path";
const css = readFileSync(join(process.cwd(), "shared.css"), "utf8"); // next start runs in the app directory
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head><meta name="viewport" content="width=device-width, initial-scale=1" /><title>Bench · next</title><style dangerouslySetInnerHTML={{ __html: css }} /></head>
      <body>{children}</body>
    </html>
  );
}
