export default function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div class="rounded-xl border border-zinc-200 bg-white px-5 py-4 dark:border-zinc-800 dark:bg-zinc-900">
      <div class="font-mono text-2xl font-bold text-brand-600 dark:text-brand-300">{value}</div>
      <div class="mt-1 text-xs text-zinc-500 dark:text-zinc-400">{label}</div>
    </div>
  );
}
