import { useState, useEffect, on, emit } from "gotsx";
import { create, update } from "host:users";     // typed actions: UsersModule.Create / Update (validated in Go)
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
    try {
      if (isEdit) await update(id, name, email, role, dept, status);
      else await create(name, email, role, dept);
      setOpen(false);
      emit("admin:changed", {});
      emit("admin:toast", { msg: isEdit ? "Changes saved" : "User created", kind: "ok" });
    } catch (e) {
      setErrs(e.status === 422 ? e.fields : { _: e.message });   // field messages from gotsx.Invalid, else the error
    } finally {
      setBusy(false);
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
              <h2 class="text-base font-semibold tracking-tight">{id !== "" ? "Edit user" : "New user"}</h2>
              <p class="mt-1 text-sm text-muted-foreground">{id !== "" ? "Save your changes; the server validates again." : "Fill in the basics; the email must be unique."}</p>
            </div>
            <button class="btn btn-ghost btn-icon-sm -mr-2 -mt-1 text-muted-foreground" aria-label="Close" onClick={() => setOpen(false)}><Icon name="x" /></button>
          </div>
          {errs._ !== undefined && <div class="alert alert-error mb-4 flex items-center gap-2"><Icon name="alert" />{errs._}</div>}
          <div class="space-y-4">
            <div class="space-y-2">
              <label class="label">Name</label>
              <input class={field(errs.name !== undefined)} value={name} onInput={(e) => setName(e.target.value)} />
              {errs.name !== undefined && <p class="text-xs text-destructive">{errs.name}</p>}
            </div>
            <div class="space-y-2">
              <label class="label">Email</label>
              <input class={field(errs.email !== undefined)} value={email} onInput={(e) => setEmail(e.target.value)} />
              {errs.email !== undefined && <p class="text-xs text-destructive">{errs.email}</p>}
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="label">Role</label>
                <select class={select(errs.role !== undefined)} value={role} onChange={(e) => setRole(e.target.value)}>
                  <option value="admin">Admin</option>
                  <option value="editor">Editor</option>
                  <option value="viewer">Viewer</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="label">Department</label>
                <input class={field(errs.dept !== undefined)} value={dept} onInput={(e) => setDept(e.target.value)} />
                {errs.dept !== undefined && <p class="text-xs text-destructive">{errs.dept}</p>}
              </div>
            </div>
            {id !== "" && (
              <div class="space-y-2">
                <label class="label">Status</label>
                <select class={select(false)} value={status} onChange={(e) => setStatus(e.target.value)}>
                  <option value="active">Active</option>
                  <option value="disabled">Disabled</option>
                </select>
              </div>
            )}
          </div>
          <div class="mt-6 flex justify-end gap-2">
            <button class="btn btn-outline" onClick={() => setOpen(false)}>Cancel</button>
            <button class="btn btn-primary" disabled={busy} onClick={save}>
              {busy && <Icon name="spinner" cls="icon animate-spin" />}
              Save
            </button>
          </div>
        </div>
      </div>
    ) : <div></div>
  );
}
