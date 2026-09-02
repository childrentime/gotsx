import { useState, useEffect, on, emit } from "gotsx";

interface Errs { name?: string; email?: string; role?: string; dept?: string; _?: string; }
interface UserForm { id?: string; name?: string; email?: string; role?: string; dept?: string; status?: string; }

export default function UserModal() {
  const [open, setOpen] = useState(false);
  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("viewer");
  const [dept, setDept] = useState("");
  const [status, setStatus] = useState("active");
  const [errs, setErrs] = useState<Errs>({});
  const [busy, setBusy] = useState(false);

  const reset = (u: UserForm) => {
    setId(u.id ?? "");
    setName(u.name ?? "");
    setEmail(u.email ?? "");
    setRole(u.role ?? "viewer");
    setDept(u.dept ?? "");
    setStatus(u.status ?? "active");
    setErrs({});
    setOpen(true);
  };
  useEffect(() => {
    on("admin:new", () => reset({}));
    on("admin:edit", (u: any) => reset(u));
  }, []);

  const save = async () => {
    setBusy(true);
    setErrs({});
    const isEdit = id !== "";
    const url = isEdit ? "/users/update" : "/users/create";
    const r = await fetch(url, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ id, name, email, role, dept, status }) });
    const d = await r.json();
    setBusy(false);
    if (d.ok) {
      setOpen(false);
      emit("admin:changed", {});
      emit("admin:toast", { msg: isEdit ? "已保存修改" : "已创建用户", kind: "ok" });
    } else {
      setErrs(d.errors);
    }
  };

  const field = (bad: boolean) => bad
    ? "h-10 w-full rounded-lg border-2 border-rose-400 bg-white px-3 text-sm outline-none"
    : "h-10 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm outline-none focus:border-brand-500";

  return (
    open ? (
      <div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/40 p-4" onClick={() => setOpen(false)}>
        <div class="modal-in w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl" onClick={(e) => e.stopPropagation()}>
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-lg font-bold">{id !== "" ? "编辑用户" : "新建用户"}</h2>
            <button class="text-slate-400 hover:text-slate-700" onClick={() => setOpen(false)}>✕</button>
          </div>
          {errs._ !== undefined && <div class="mb-3 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-600">{errs._}</div>}
          <div class="space-y-3">
            <div>
              <label class="mb-1 block text-xs font-medium text-slate-500">姓名</label>
              <input class={field(errs.name !== undefined)} value={name} onInput={(e) => setName(e.target.value)} />
              {errs.name !== undefined && <p class="mt-1 text-xs text-rose-500">{errs.name}</p>}
            </div>
            <div>
              <label class="mb-1 block text-xs font-medium text-slate-500">邮箱</label>
              <input class={field(errs.email !== undefined)} value={email} onInput={(e) => setEmail(e.target.value)} />
              {errs.email !== undefined && <p class="mt-1 text-xs text-rose-500">{errs.email}</p>}
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="mb-1 block text-xs font-medium text-slate-500">角色</label>
                <select class={field(errs.role !== undefined)} value={role} onChange={(e) => setRole(e.target.value)}>
                  <option value="admin">管理员</option>
                  <option value="editor">编辑</option>
                  <option value="viewer">只读</option>
                </select>
              </div>
              <div>
                <label class="mb-1 block text-xs font-medium text-slate-500">部门</label>
                <input class={field(errs.dept !== undefined)} value={dept} onInput={(e) => setDept(e.target.value)} />
                {errs.dept !== undefined && <p class="mt-1 text-xs text-rose-500">{errs.dept}</p>}
              </div>
            </div>
            {id !== "" && (
              <div>
                <label class="mb-1 block text-xs font-medium text-slate-500">状态</label>
                <select class={field(false)} value={status} onChange={(e) => setStatus(e.target.value)}>
                  <option value="active">启用</option>
                  <option value="disabled">禁用</option>
                </select>
              </div>
            )}
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <button class="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50" onClick={() => setOpen(false)}>取消</button>
            <button class="inline-flex items-center gap-2 rounded-lg bg-brand-500 px-5 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50" disabled={busy} onClick={save}>
              {busy && <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"></span>}
              保存
            </button>
          </div>
        </div>
      </div>
    ) : <div></div>
  );
}
