import type { PageProps, Meta } from "gotsx";
import Shell from "../../components/Shell.server";
import UsersTable from "../../islands/UsersTable.client";
import UserModal from "../../islands/UserModal.client";

export function meta(): Meta {
  return { title: "Users", description: "Search, create, edit and remove users; every write is a typed action validated in Go." };
}

export default function UsersPage({ session }: PageProps) {
  if (session.user === "") redirect("/login");
  return (
    <Shell title="Users" active="users" name={session.name} role={session.role}>
      <UsersTable canEdit={session.role === "admin" || session.role === "editor"} />
      <UserModal />
    </Shell>
  );
}
