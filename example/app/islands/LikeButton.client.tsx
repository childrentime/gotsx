import { useState } from "gotsx";

/** 岛回到 Go 的通道是 HTTP: /actions/like 是 Go handler。async 函数只进 JS 后端 */
export default function LikeButton({ id, likes }: { id: string; likes: number }) {
  const [n, setN] = useState(likes);
  const [busy, setBusy] = useState(false);
  async function like() {
    setBusy(true);
    try {
      const r = await fetch(`/actions/like?id=${encodeURIComponent(id)}`, { method: "POST" });
      const data = await r.json();
      setN(data.likes);
    } finally {
      setBusy(false);
    }
  }
  return (
    <button class="btn" disabled={busy} onClick={like}>♥ {n}</button>
  );
}
