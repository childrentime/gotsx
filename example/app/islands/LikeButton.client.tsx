import { useState } from "gotsx";
import { like } from "host:data";

/** Islands reach Go through typed actions: `like` is DataModule.Like, listed in the host Registry's Actions.
 *  The call compiles to a same-origin POST; the return type comes from the Go signature. */
export default function LikeButton({ id, likes }: { id: string; likes: number }) {
  const [n, setN] = useState(likes);
  const [busy, setBusy] = useState(false);
  async function onLike() {
    setBusy(true);
    try {
      setN(await like(id));
    } finally {
      setBusy(false);
    }
  }
  return (
    <button class="btn btn-outline btn-sm" disabled={busy} onClick={onLike}>
      <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" /></svg>
      {n}
    </button>
  );
}
