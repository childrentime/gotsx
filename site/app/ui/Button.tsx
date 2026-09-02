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
  if (size === "sm") return "btn-sm";
  if (size === "lg") return "btn-lg";
  return "";
}

function variantCls(variant: string): string {
  if (variant === "secondary") return "btn-secondary";
  if (variant === "outline") return "btn-outline";
  if (variant === "ghost") return "btn-ghost";
  return "btn-primary";
}

/** 组件库的 Button: 用方言写的, 服务端编成 Go, 客户端编成 DOM 指令。样式来自设计系统的 .btn-* */
export default function Button({ variant = "primary", size = "md", href = "", disabled = false, className = "", onClick, children }: ButtonProps) {
  const cls = `btn ${variantCls(variant)} ${sizeCls(size)} ${className}`;
  return href !== "" ? (
    <a href={href} class={cls}>{children}</a>
  ) : (
    <button class={cls} disabled={disabled} onClick={onClick}>{children}</button>
  );
}
