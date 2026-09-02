import type { Node } from "gotsx";

/** 16px 线性图标(lucide 风格), 两端共用。stroke 用 currentColor, 颜色跟随文字 */
function glyph(name: string): Node {
  switch (name) {
    case "search":
      return <><circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" /></>;
    case "cart":
      return <><circle cx="8" cy="21" r="1" /><circle cx="19" cy="21" r="1" /><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12" /></>;
    case "package":
      return <><path d="m7.5 4.27 9 5.15" /><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z" /><path d="m3.3 7 8.7 5 8.7-5" /><path d="M12 22V12" /></>;
    case "truck":
      return <><path d="M14 18V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v11a1 1 0 0 0 1 1h2" /><path d="M15 18H9" /><path d="M19 18h2a1 1 0 0 0 1-1v-3.65a1 1 0 0 0-.22-.62l-3.48-4.35A1 1 0 0 0 17.52 8H14" /><circle cx="17" cy="18" r="2" /><circle cx="7" cy="18" r="2" /></>;
    case "undo":
      return <><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" /><path d="M3 3v5h5" /></>;
    case "shield":
      return <><path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" /><path d="m9 12 2 2 4-4" /></>;
    case "zap":
      return <path d="M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z" />;
    case "heart":
      return <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" />;
    case "check":
      return <path d="M20 6 9 17l-5-5" />;
    case "check-circle":
      return <><circle cx="12" cy="12" r="10" /><path d="m9 12 2 2 4-4" /></>;
    case "map-pin":
      return <><path d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0" /><circle cx="12" cy="10" r="3" /></>;
    case "globe":
      return <><circle cx="12" cy="12" r="10" /><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" /><path d="M2 12h20" /></>;
    case "chevron-right":
      return <path d="m9 18 6-6-6-6" />;
    case "arrow-right":
      return <><path d="M5 12h14" /><path d="m12 5 7 7-7 7" /></>;
    case "minus":
      return <path d="M5 12h14" />;
    case "plus":
      return <><path d="M5 12h14" /><path d="M12 5v14" /></>;
    case "x":
      return <><path d="M18 6 6 18" /><path d="m6 6 12 12" /></>;
    case "trash":
      return <><path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" /><path d="M10 11v6" /><path d="M14 11v6" /></>;
    case "alert":
      return <><circle cx="12" cy="12" r="10" /><path d="M12 8v4" /><path d="M12 16h.01" /></>;
    case "loader":
      return <path d="M21 12a9 9 0 1 1-6.219-8.56" />;
    case "sparkles":
      return <><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z" /></>;
    default:
      return <circle cx="12" cy="12" r="9" />;
  }
}

export default function Icon({ name, className = "", fill = "none" }: { name: string; className?: string; fill?: string }) {
  return (
    <svg class={className !== "" ? `icon ${className}` : "icon"} viewBox="0 0 24 24" fill={fill} stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">{glyph(name)}</svg>
  );
}
