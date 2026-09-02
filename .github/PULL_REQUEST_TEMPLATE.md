**What this changes**

**Checklist**
- [ ] `make test` passes (`make gen` first — `gen/` is gitignored)
- [ ] A language change lands in the checker **and** both backends, with a `lang_test.go` case per backend, and is added to the syntax table in `site/app/content/site.server.tsx`
- [ ] Demo UI changes stay inside the tokens / component classes of `design/README.md`
- [ ] User-visible changes have a `CHANGELOG.md` entry (and a `STABILITY.md` tier if they add API)
