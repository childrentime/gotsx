import type { PageProps } from "gotsx";
import Shell from "../../components/Shell.server";
import UsersTable from "../../islands/UsersTable.client";
import UserModal from "../../islands/UserModal.client";
import Toasts from "../../islands/Toasts.client";

export default function UsersPage({ cookies }: PageProps) {
  return (
    <Shell title="用户管理" active="users" name={cookies._name ?? ""} role={cookies._role ?? ""}>
      <UsersTable canEdit={cookies._role === "admin" || cookies._role === "editor"} />
      <UserModal />
      <Toasts />
    </Shell>
  );
}
