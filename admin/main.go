// gotsx admin: a back-office demo — sign-in on the framework session, protected pages, a users table with CRUD
// through typed actions, server-side validation, modal, toasts (flash + client events).
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"

	"github.com/childrentime/gotsx/admin/gen"
	gotsx "github.com/childrentime/gotsx/runtime"
)

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "listen address")
	dev := flag.Bool("dev", false, "development mode")
	flag.Parse()

	err := gotsx.Serve(gotsx.Options{
		Addr:      *addr,
		Dev:       *dev,
		Routes:    gen.Routes,
		ClientDir: gotsx.FindDir("gen/client"),
		ClientFS:  gen.ClientFS,
		PublicDir: gotsx.FindDir("public"),
		PublicFS:  mustSub(publicEmbed, "public"),
		NotFound:  gen.NotFound,  // pages/_404.server.tsx
		ErrorPage: gen.ErrorPage, // pages/_error.server.tsx
		// Typed actions (host.Registry[...].Actions): login/logout and the users CRUD. Each action checks the
		// session itself (host.requireUser / requireEditor); pages redirect to /login when session.user is empty.
		HostActions: gen.HostActions,
		// Signed session cookie: profile of the signed-in account + flash messages. Empty → random per start.
		SessionSecret: os.Getenv("SESSION_SECRET"),
	})
	log.Fatal(err)
}

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
