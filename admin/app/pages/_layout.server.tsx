import type { LayoutProps } from "gotsx";
import Toasts from "../islands/Toasts.client";

// pages/_layout.server.tsx wraps every page: the document, <title> / description from the page's meta,
// and the session's flash messages rendered as toasts. The signed-in chrome (sidebar, header) is components/Shell.
export default function Layout({ meta, flash, children }: LayoutProps) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{meta.title ? meta.title + " · gotsx admin" : "gotsx admin"}</title>
        {meta.description && <meta name="description" content={meta.description} />}
        <link rel="stylesheet" href="/public/tailwind.css" />
      </head>
      <body class="min-h-screen bg-background text-foreground">
        <div id="gotsx-bar"></div>
        {children}
        <Toasts initial={flash} />
      </body>
    </html>
  );
}
