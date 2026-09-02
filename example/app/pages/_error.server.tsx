import type { ErrorProps } from "gotsx";

/** pages/_error.server.tsx → gen.ErrorPage; message 在 dev 下是错误详情, 生产下是通用文案 */
export default function Oops({ message }: ErrorProps) {
  return (
    <div class="empty">
      <h1>Something went wrong</h1>
      <p class="muted">{message}</p>
      <a href="/" class="btn btn-outline">Back to models</a>
    </div>
  );
}
