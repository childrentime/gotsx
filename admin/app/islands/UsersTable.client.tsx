import { useState, useEffect, on, emit } from "gotsx";
import type { User } from "host:users";
import Badge from "../ui/Badge";
import Avatar from "../ui/Avatar";
import Icon from "../ui/Icon";

const perPage = 6;

export default function UsersTable({ canEdit }: { canEdit: boolean }) {
  const [all, setAll] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [active, setActive] = useState(0);
  const [admins, setAdmins] = useState(0);
  const [q, setQ] = useState("");
  const [role, setRole] = useState("all");
  const [sort, setSort] = useState("name");
  const [page, setPage] = useState(1);
  const [deleting, setDeleting] = useState("");

  const load = async () => {
    setLoading(true);
    const r = await fetch("/api/users");
    const d = await r.json();
    setAll(d.items as User[]);
    setActive(d.active);
    setAdmins(d.admins);
    setLoading(false);
  };
  useEffect(() => {
    load();
    on("admin:changed", () => load());
  }, []);

  // 客户端搜索 + 角色过滤 + 排序
  const filtered = all.filter((u) => {
    const k = q.toLowerCase();
    const hit = k === "" || u.name.toLowerCase().includes(k) || u.email.toLowerCase().includes(k) || u.dept.includes(q);
    return hit && (role === "all" || u.role === role);
  });
  const sorted = filtered.sort((a, b) => {
    if (sort === "name") return a.name < b.name ? -1 : 1;
    if (sort === "role") return a.role < b.role ? -1 : 1;
    return a.dept < b.dept ? -1 : 1;
  });
  const pages = Math.max(1, Math.ceil(sorted.length / perPage));
  const cur = Math.min(page, pages);
  const shown = sorted.slice((cur - 1) * perPage, cur * perPage);

  const roleLabel = (r: string) => (r === "admin" ? "管理员" : r === "editor" ? "编辑" : "只读");
  const del = async (u: User) => {
    if (!window.confirm(`确定删除用户「${u.name}」?`)) return;
    setDeleting(u.id);
    const r = await fetch("/users/delete", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ id: u.id }) });
    const d = await r.json();
    setDeleting("");
    if (d.ok) {
      emit("admin:toast", { msg: "已删除用户", kind: "ok" });
      load();
    } else {
      emit("admin:toast", { msg: "删除失败", kind: "err" });
    }
  };

  return (
    <div>
      <div class="mb-4 grid gap-4 sm:grid-cols-3">
        <div class="card p-5"><div class="text-sm text-muted-foreground">用户总数</div><div class="mt-2 text-2xl font-semibold tracking-tight">{all.length}</div></div>
        <div class="card p-5"><div class="text-sm text-muted-foreground">活跃</div><div class="mt-2 text-2xl font-semibold tracking-tight">{active}</div></div>
        <div class="card p-5"><div class="text-sm text-muted-foreground">管理员</div><div class="mt-2 text-2xl font-semibold tracking-tight">{admins}</div></div>
      </div>

      <div class="card">
        <div class="flex flex-wrap items-center gap-2 border-b border-border p-4">
          <div class="relative min-w-[200px] flex-1">
            <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"><Icon name="search" /></span>
            <input class="input pl-9" placeholder="搜索姓名 / 邮箱 / 部门…" value={q} onInput={(e) => { setQ(e.target.value); setPage(1); }} />
          </div>
          <select class="select" value={role} onChange={(e) => { setRole(e.target.value); setPage(1); }}>
            <option value="all">全部角色</option>
            <option value="admin">管理员</option>
            <option value="editor">编辑</option>
            <option value="viewer">只读</option>
          </select>
          <select class="select" value={sort} onChange={(e) => setSort(e.target.value)}>
            <option value="name">按姓名</option>
            <option value="role">按角色</option>
            <option value="dept">按部门</option>
          </select>
          {canEdit && <button class="btn btn-primary" onClick={() => emit("admin:new", {})}><Icon name="plus" />新建用户</button>}
        </div>

        {loading ? (
          <div class="divide-y divide-border">
            {[0, 1, 2, 3, 4, 5].map(() => (
              <div class="flex items-center gap-3 px-4 py-3"><div class="skeleton h-8 w-8 rounded-full"></div><div class="flex-1 space-y-2"><div class="skeleton h-3 w-1/3"></div><div class="skeleton h-2.5 w-1/4"></div></div><div class="skeleton h-5 w-16"></div></div>
            ))}
          </div>
        ) : shown.length === 0 ? (
          <div class="flex flex-col items-center py-20 text-muted-foreground"><Icon name="search" cls="h-6 w-6" /><p class="mt-3 text-sm">没有匹配的用户</p></div>
        ) : (
          <div class="overflow-x-auto">
            <table class="table min-w-[640px]">
              <thead>
                <tr><th>用户</th><th>角色</th><th>部门</th><th>状态</th><th>最近活跃</th><th></th></tr>
              </thead>
              <tbody>
                {shown.map((u) => (
                  <tr key={u.id}>
                    <td>
                      <div class="flex items-center gap-3">
                        <Avatar name={u.name} />
                        <div><div class="font-medium">{u.name}</div><div class="text-xs text-muted-foreground">{u.email}</div></div>
                      </div>
                    </td>
                    <td><Badge tone={u.role}>{roleLabel(u.role)}</Badge></td>
                    <td class="whitespace-nowrap text-muted-foreground">{u.dept}</td>
                    <td><Badge tone={u.status}>{u.status === "active" ? "启用" : "禁用"}</Badge></td>
                    <td class="whitespace-nowrap text-xs text-muted-foreground">{u.lastSeen}</td>
                    <td class="text-right">
                      {canEdit && (
                        <div class="flex justify-end gap-1">
                          <button class="btn btn-ghost btn-icon-sm text-muted-foreground" title="编辑" aria-label="编辑" onClick={() => emit("admin:edit", u)}><Icon name="pencil" /></button>
                          <button class="btn btn-ghost btn-icon-sm text-muted-foreground hover:text-destructive" title="删除" aria-label="删除" disabled={deleting === u.id} onClick={() => del(u)}><Icon name="trash" /></button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!loading && sorted.length > perPage && (
          <div class="flex items-center justify-between border-t border-border px-4 py-3 text-sm">
            <span class="text-muted-foreground">共 {sorted.length} 条 · 第 {cur} / {pages} 页</span>
            <div class="flex gap-2">
              <button class="btn btn-outline btn-sm" disabled={cur <= 1} onClick={() => setPage(cur - 1)}><Icon name="left" />上一页</button>
              <button class="btn btn-outline btn-sm" disabled={cur >= pages} onClick={() => setPage(cur + 1)}>下一页<Icon name="right" /></button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
