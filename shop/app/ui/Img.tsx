/** 图片组件: 懒加载、异步解码、固定宽高(防布局抖动 CLS)、alt(SEO/无障碍) */
export default function Img({ src, alt, className = "", eager = false }: { src: string; alt: string; className?: string; eager?: boolean }) {
  return <img src={src} alt={alt} width={400} height={400} loading={eager ? "eager" : "lazy"} decoding="async" class={className} />;
}
