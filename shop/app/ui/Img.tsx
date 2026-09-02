/** Image: lazy, async decoding, fixed width/height (no CLS), alt text (SEO / a11y) */
export default function Img({ src, alt, className = "", eager = false }: { src: string; alt: string; className?: string; eager?: boolean }) {
  return <img src={src} alt={alt} width={400} height={400} loading={eager ? "eager" : "lazy"} decoding="async" class={className} />;
}
