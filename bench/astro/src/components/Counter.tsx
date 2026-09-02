import { useState } from "preact/hooks";
export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  return <button onClick={() => setN(n + 1)}>count: {n}</button>;
}
