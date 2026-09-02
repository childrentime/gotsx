import { useState, useEffect, on, emit } from "gotsx";
import type { User } from "host:users";
import Badge from "../ui/Badge";
import Avatar from "../ui/Avatar";

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
        <div class="rounded-xl border border-slate-200 bg-white p-4"><div class="text-xs text-slate-400">用户总数</div><div class="mt-1 text-2xl font-black">{all.length}</div></div>
        <div class="rounded-xl border border-slate-200 bg-white p-4"><div class="text-xs text-slate-400">活跃</div><div class="mt-1 text-2xl font-black text-emerald-600">{active}</div></div>
        <div class="rounded-xl border border-slate-200 bg-white p-4"><div class="text-xs text-slate-400">管理员</div><div class="mt-1 text-2xl font-black text-violet-600">{admins}</div></div>
      </div>

      <div class="rounded-xl border border-slate-200 bg-white">
        <div class="flex flex-wrap items-center gap-3 border-b border-slate-100 p-4">
          <div class="flex h-10 flex-1 items-center rounded-lg border border-slate-300 px-3 focus-within:border-brand-500">
            <span class="mr-2 text-slate-400">🔍</span>
            <input class="h-full min-w-0 flex-1 bg-transparent text-sm outline-none" placeholder="搜索姓名 / 邮箱 / 部门…" value={q} onInput={(e) => { setQ(e.target.value); setPage(1); }} />
          </div>
          <select class="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm" value={role} onChange={(e) => { setRole(e.target.value); setPage(1); }}>
            <option value="all">全部角色</option>
            <option value="admin">管理员</option>
            <option value="editor">编辑</option>
            <option value="viewer">只读</option>
          </select>
          <select class="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm" value={sort} onChange={(e) => setSort(e.target.value)}>
            <option value="name">按姓名</option>
            <option value="role">按角色</option>
            <option value="dept">按部门</option>
          </select>
          {canEdit && <button class="h-10 rounded-lg bg-brand-500 px-4 text-sm font-bold text-white hover:bg-brand-600" onClick={() => emit("admin:new", {})}>+ 新建用户</button>}
        </div>

        {loading ? (
          <div class="divide-y divide-slate-100">
            {[0, 1, 2, 3, 4, 5].map(() => (
              <div class="flex items-center gap-3 p-4"><div class="skel h-8 w-8 rounded-full"></div><div class="flex-1 space-y-2"><div class="skel h-3 w-1/3"></div><div class="skel h-2.5 w-1/4"></div></div><div class="skel h-5 w-16"></div></div>
            ))}
          </div>
        ) : shown.length === 0 ? (
          <div class="py-20 text-center text-slate-400"><div class="text-4xl">🔍</div><p class="mt-3 text-sm">没有匹配的用户</p></div>
        ) : (
          <div class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-400">
                <tr><th class="px-4 py-3">用户</th><th class="px-4 py-3">角色</th><th class="px-4 py-3">部门</th><th class="px-4 py-3">状态</th><th class="px-4 py-3">最近活跃</th><th class="px-4 py-3"></th></tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                {shown.map((u) => (
                  <tr class="hover:bg-slate-50">
                    <td class="px-4 py-3">
                      <div class="flex items-center gap-3">
                        <Avatar name={u.name} />
                        <div><div class="font-medium">{u.name}</div><div class="text-xs text-slate-400">{u.email}</div></div>
                      </div>
                    </td>
                    <td class="px-4 py-3"><Badge tone={u.role}>{roleLabel(u.role)}</Badge></td>
                    <td class="px-4 py-3 text-slate-600">{u.dept}</td>
                    <td class="px-4 py-3"><Badge tone={u.status}>{u.status === "active" ? "启用" : "禁用"}</Badge></td>
                    <td class="px-4 py-3 text-xs text-slate-400">{u.lastSeen}</td>
                    <td class="px-4 py-3 text-right">
                      {canEdit && (
                        <div class="flex justify-end gap-1">
                          <button class="rounded-md px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 hover:text-brand-600" onClick={() => emit("admin:edit", u)}>编辑</button>
                          <button class="rounded-md px-2 py-1 text-xs text-slate-500 hover:bg-rose-50 hover:text-rose-600 disabled:opacity-40" disabled={deleting === u.id} onClick={() => del(u)}>删除</button>
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
          <div class="flex items-center justify-between border-t border-slate-100 p-4 text-sm">
            <span class="text-slate-400">共 {sorted.length} 条 · 第 {cur} / {pages} 页</span>
            <div class="flex gap-2">
              <button class="rounded-lg border border-slate-300 px-3 py-1.5 text-slate-600 hover:bg-slate-50 disabled:opacity-40" disabled={cur <= 1} onClick={() => setPage(cur - 1)}>上一页</button>
              <button class="rounded-lg border border-slate-300 px-3 py-1.5 text-slate-600 hover:bg-slate-50 disabled:opacity-40" disabled={cur >= pages} onClick={() => setPage(cur + 1)}>下一页</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
