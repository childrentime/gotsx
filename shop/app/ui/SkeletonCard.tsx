/** Loading skeleton card */
export default function SkeletonCard() {
  return (
    <div class="card overflow-hidden">
      <div class="skeleton aspect-square rounded-none"></div>
      <div class="space-y-2 p-3">
        <div class="skeleton h-3 w-full"></div>
        <div class="skeleton h-3 w-3/5"></div>
        <div class="skeleton h-4 w-2/5"></div>
      </div>
    </div>
  );
}
