export default function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div class="card px-5 py-4">
      <div class="font-mono text-2xl font-semibold tracking-tight text-foreground">{value}</div>
      <div class="mt-1 text-xs leading-5 text-muted-foreground">{label}</div>
    </div>
  );
}
