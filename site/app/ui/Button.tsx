import type { Node } from "gotsx";

export interface ButtonProps {
  variant?: string;      // primary | secondary | outline | ghost
  size?: string;         // sm | md | lg
  href?: string;         // 有 href 就渲染成 <a>
  disabled?: boolean;
  className?: string;
  onClick?: () => void;
  children?: Node;
}

function sizeCls(size: string): string {
  if (size === "sm") return "h-8 px-3 text-sm";
  if (size === "lg") return "h-12 px-6 text-base";
  return "h-10 px-4 text-sm";
}

function variantCls(variant: string): string {
  if (variant === "secondary") return "bg-zinc-100 text-zinc-900 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-100 dark:hover:bg-zinc-700";
  if (variant === "outline") return "border border-zinc-300 text-zinc-900 hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-100 dark:hover:bg-zinc-800";
  if (variant === "ghost") return "text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800";
  return "bg-brand-600 text-white hover:bg-brand-700 shadow-sm";
}

/** 组件库的 Button: 用方言写的, 服务端编成 Go, 客户端编成 DOM 指令 */
export default function Button({ variant = "primary", size = "md", href = "", disabled = false, className = "", onClick, children }: ButtonProps) {
  const cls = `inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:pointer-events-none ${sizeCls(size)} ${variantCls(variant)} ${className}`;
  return href !== "" ? (
    <a href={href} class={cls}>{children}</a>
  ) : (
    <button class={cls} disabled={disabled} onClick={onClick}>{children}</button>
  );
}
