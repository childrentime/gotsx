/** 加载骨架卡 */
export default function SkeletonCard() {
  return (
    <div class="overflow-hidden rounded-xl2 border border-ink-100 bg-white shadow-card">
      <div class="skel aspect-square rounded-none"></div>
      <div class="space-y-2 p-2.5">
        <div class="skel h-3 w-full"></div>
        <div class="skel h-3 w-3/5"></div>
        <div class="skel h-4 w-2/5"></div>
      </div>
    </div>
  );
}
