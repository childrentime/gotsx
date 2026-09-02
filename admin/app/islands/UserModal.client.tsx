import { useState, useEffect, on, emit } from "gotsx";
import Icon from "../ui/Icon";

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

  const field = (bad: boolean) => (bad ? "input border-destructive focus-visible:ring-destructive/30" : "input");
  const select = (bad: boolean) => (bad ? "select w-full border-destructive" : "select w-full");

  return (
    open ? (
      <div class="fixed inset-0 z-40 flex items-center justify-center bg-foreground/40 p-4" onClick={() => setOpen(false)}>
        <div class="card pop-in w-full max-w-md p-6 shadow-lg" role="dialog" aria-modal="true" onClick={(e) => e.stopPropagation()}>
          <div class="mb-5 flex items-start justify-between">
            <div>
              <h2 class="text-base font-semibold tracking-tight">{id !== "" ? "编辑用户" : "新建用户"}</h2>
              <p class="mt-1 text-sm text-muted-foreground">{id !== "" ? "修改后保存, 服务端会再次校验。" : "填写基本信息, 邮箱需唯一。"}</p>
            </div>
            <button class="btn btn-ghost btn-icon-sm -mr-2 -mt-1 text-muted-foreground" aria-label="关闭" onClick={() => setOpen(false)}><Icon name="x" /></button>
          </div>
          {errs._ !== undefined && <div class="mb-4 flex items-center gap-2 rounded-md border border-destructive/30 px-3 py-2 text-sm text-destructive"><Icon name="alert" />{errs._}</div>}
          <div class="space-y-4">
            <div class="space-y-2">
              <label class="label">姓名</label>
              <input class={field(errs.name !== undefined)} value={name} onInput={(e) => setName(e.target.value)} />
              {errs.name !== undefined && <p class="text-xs text-destructive">{errs.name}</p>}
            </div>
            <div class="space-y-2">
              <label class="label">邮箱</label>
              <input class={field(errs.email !== undefined)} value={email} onInput={(e) => setEmail(e.target.value)} />
              {errs.email !== undefined && <p class="text-xs text-destructive">{errs.email}</p>}
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="label">角色</label>
                <select class={select(errs.role !== undefined)} value={role} onChange={(e) => setRole(e.target.value)}>
                  <option value="admin">管理员</option>
                  <option value="editor">编辑</option>
                  <option value="viewer">只读</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="label">部门</label>
                <input class={field(errs.dept !== undefined)} value={dept} onInput={(e) => setDept(e.target.value)} />
                {errs.dept !== undefined && <p class="text-xs text-destructive">{errs.dept}</p>}
              </div>
            </div>
            {id !== "" && (
              <div class="space-y-2">
                <label class="label">状态</label>
                <select class={select(false)} value={status} onChange={(e) => setStatus(e.target.value)}>
                  <option value="active">启用</option>
                  <option value="disabled">禁用</option>
                </select>
              </div>
            )}
          </div>
          <div class="mt-6 flex justify-end gap-2">
            <button class="btn btn-outline" onClick={() => setOpen(false)}>取消</button>
            <button class="btn btn-primary" disabled={busy} onClick={save}>
              {busy && <Icon name="spinner" cls="icon animate-spin" />}
              保存
            </button>
          </div>
        </div>
      </div>
    ) : <div></div>
  );
}
