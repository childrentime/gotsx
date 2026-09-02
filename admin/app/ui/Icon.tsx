import type { Node } from "gotsx";

/** 16px 线性图标(lucide 风格路径), currentColor; 两端共用 */
function paths(name: string): Node {
  if (name === "dashboard") return <><rect width="7" height="9" x="3" y="3" rx="1" /><rect width="7" height="5" x="14" y="3" rx="1" /><rect width="7" height="9" x="14" y="12" rx="1" /><rect width="7" height="5" x="3" y="16" rx="1" /></>;
  if (name === "users") return <><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></>;
  if (name === "search") return <><circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" /></>;
  if (name === "plus") return <><path d="M5 12h14" /><path d="M12 5v14" /></>;
  if (name === "pencil") return <><path d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z" /><path d="m15 5 4 4" /></>;
  if (name === "trash") return <><path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" /></>;
  if (name === "x") return <><path d="M18 6 6 18" /><path d="m6 6 12 12" /></>;
  if (name === "check") return <><circle cx="12" cy="12" r="10" /><path d="m9 12 2 2 4-4" /></>;
  if (name === "alert") return <><circle cx="12" cy="12" r="10" /><path d="M12 8v4" /><path d="M12 16h.01" /></>;
  if (name === "logout") return <><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="m16 17 5-5-5-5" /><path d="M21 12H9" /></>;
  if (name === "left") return <path d="m15 18-6-6 6-6" />;
  if (name === "right") return <path d="m9 18 6-6-6-6" />;
  if (name === "up") return <><path d="M7 7h10v10" /><path d="M7 17 17 7" /></>;
  if (name === "down") return <><path d="m7 7 10 10" /><path d="M17 7v10" /></>;
  if (name === "shield") return <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" />;
  if (name === "activity") return <path d="M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" />;
  if (name === "spinner") return <path d="M21 12a9 9 0 1 1-6.219-8.56" />;
  return <circle cx="12" cy="12" r="1" />;
}

export default function Icon({ name, cls = "icon" }: { name: string; cls?: string }) {
  return (
    <svg class={cls} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      {paths(name)}
    </svg>
  );
}
