/** 内联 SVG 图标(lucide 风格, 16px, stroke 2, currentColor)。共享组件: 页面和岛都能用。 */
function path(name: string): string {
  if (name === "sun") return "M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0M12 2v2M12 20v2m-7.07-15.07 1.41 1.41m11.32 11.32 1.41 1.41M2 12h2m16 0h2M6.34 17.66l-1.41 1.41m14.14-14.14-1.41 1.41";
  if (name === "moon") return "M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z";
  if (name === "check") return "M20 6 9 17l-5-5";
  if (name === "copy") return "M10 8h10a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H10a2 2 0 0 1-2-2V10a2 2 0 0 1 2-2zM4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2";
  if (name === "chevron-down") return "m6 9 6 6 6-6";
  if (name === "chevron-right") return "m9 18 6-6-6-6";
  if (name === "arrow-right") return "M5 12h14m-7-7 7 7-7 7";
  if (name === "info") return "M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0M12 16v-4m0-4h.01";
  if (name === "alert") return "m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3ZM12 9v4m0 4h.01";
  if (name === "bulb") return "M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.7 1.3 1.5 1.5 2.5M9 18h6m-5 4h4";
  if (name === "flame") return "M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z";
  if (name === "cpu") return "M6 6h12v12H6zM9 9h6v6H9zM9 1v5m6-5v5M9 18v5m6-5v5m3-14h5m-5 5h5M1 9h5m-5 5h5";
  if (name === "zap") return "M13 2 3 14h9l-1 8 10-12h-9l1-8z";
  if (name === "compass") return "M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0m-5.76-4.24-2.12 6.36-6.36 2.12 2.12-6.36 6.36-2.12z";
  if (name === "plug") return "M12 22v-5M9 8V2m6 6V2m3 6H6v5a6 6 0 0 0 12 0V8Z";
  if (name === "layers") return "m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83ZM2 12l8.58 3.91a2 2 0 0 0 1.66 0L21 12M2 17l8.58 3.91a2 2 0 0 0 1.66 0L21 17";
  if (name === "shield") return "M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z";
  if (name === "palette") return "M12 22a10 10 0 1 1 0-20c5.5 0 10 3.6 10 8a4 4 0 0 1-4 4h-2a2 2 0 0 0-1.5 3.3c.3.4.5.9.5 1.4a2.3 2.3 0 0 1-3 2.3ZM7.5 10.5h.01M12 7.5h.01m4.49 3h.01";
  if (name === "refresh") return "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8M21 3v5h-5m5 4a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16m0 5v-5h5";
  if (name === "box") return "M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16ZM3.3 7 12 12l8.7-5M12 22V12";
  if (name === "code") return "m16 18 6-6-6-6M8 6l-6 6 6 6";
  if (name === "minus") return "M5 12h14";
  if (name === "plus") return "M5 12h14m-7-7v14";
  return "M12 12h.01";
}

export default function Icon({ name, className = "icon" }: { name: string; className?: string }) {
  return (
    <svg class={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d={path(name)} />
    </svg>
  );
}
