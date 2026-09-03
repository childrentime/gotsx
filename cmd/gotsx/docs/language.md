# Language reference (gotsx {{version}})

The dialect is a static language that borrows TSX syntax; the subset is defined by the type system. Every
expression must have a static type Go can represent. If a construct is not in this table, assume it does
not compile and write it another way (see the "no" rows for the suggested rewrite).

Semantic conventions that differ from JavaScript/TypeScript:

- `number` is float64 (integer host fields are converted at the boundary). Integer division: use `Math.floor`.
- An **optional primitive** (`x?: string`, `T | undefined` for primitives) is represented by its zero value: absent means `""` / `0` / `false`. `??` and `||` therefore behave the same for primitives, and `x === undefined` is `x === ""` for a string.
- An **absent object** (`x?: T` for an object type) is falsy and `=== undefined`; a `Record<string, T>` read of a missing key gives the zero value of `T` on both backends (`Object.hasOwn(m, k)` tests presence, `delete m[k]` removes).
- Strings are handled by rune (character) on the Go side: `s.length`, `s[i]`, `slice` count characters, not UTF-16 units.
- `throw` is a Go `panic` recovered by the request layer (500, or 404 when the error wraps `gotsx.ErrNotFound`); `try/catch` runs the `catch` body only in the browser.
- Code that runs on the server (`.server.tsx`, and the server render of an island) may not contain `async`, `await`, `fetch`, DOM access, timers or `Promise`. Inside an island they are allowed in event handlers and effects.
- Regex literals use the RE2 subset (no lookahead/lookbehind/backreferences). `Date.now()`, `Date.parse()` and `isoDate()` are the date API.
- Object spread and rest are client-only. Module-level bindings must be `const`.
- Objects are compared structurally by the checker, but the Go backend still needs the same Go type: passing a host `Model` where an interface `Titled { title: string }` is expected passes `gotsx check` and fails at `go build` (with the `.tsx` line). Declare the parameter with the host type.
- Arrays are homogeneous: the element type is the annotation or the first element, and every element must fit it (`["a", 1]` is an error — tuples are not in the subset; use an object per row).
- Types are checked at every call, prop, typed declaration and declared return: `f(1)` for `f(a: string)` is a compile error. `any` disables the check for that value; an optional primitive (`string | undefined`) may be passed where `string` is expected (absence is the zero value); objects are compared structurally.

## Syntax table

Category keys: module (imports/exports/files), stmt (statements), expr (expressions), jsx, type, hooks, builtin, array, string.

| category | syntax | ok | note |
|---|---|:-:|---|
| module | `import X from "./a"` | yes | default import; extensions .tsx / .server.tsx / .client.tsx may be omitted |
| module | `import { a, b as c } from "./a"` | yes | named imports |
| module | `import type { T } from "./a"` | yes | type import; client code may only import type from host:* (except actions) |
| module | `import { useState, useEffect, useMemo } from "gotsx"` | yes | hooks |
| module | `import type { Node, PageProps, LayoutProps, Meta, Flash } from "gotsx"` | yes | framework types |
| module | `import { x } from "host:name"` | yes | host module in a server component: compiles to a direct Go call |
| module | `import { toggle } from "host:name" (island) → await toggle(id)` | yes | typed action: a Go method listed in Registry[...].Actions; the call is a same-origin POST typed Promise<T> from the Go signature; errors throw with .status / .fields |
| module | `export function meta(props?: PageProps): Meta` | yes | page metadata (title, description, canonical, image, noIndex), evaluated once per request and handed to every layout as props.meta |
| module | `props.session / props.flash / props.csrf` | yes | PageProps: signed-session values (read-only), one-shot flash messages, the CSRF token for classic <form method="post"> |
| module | `export default function / export function` | yes | component (capitalized) or plain function |
| module | `export const data: T[] = [...]` | yes | module-level const → Go package var; no let |
| module | `export interface / export type` | yes | type export (all types can be import type'd) |
| module | `pages/a/[id].server.tsx / pages/docs/[...slug].server.tsx` | yes | file routing: params.id; catch-all params.slug = "x/y/z"; more specific routes win |
| module | `pages/**/_layout.server.tsx / _404 / _error` | yes | nested layouts (LayoutProps = PageProps + meta + children, outer layouts wrap inner ones); _404 → gen.NotFound, _error (ErrorProps) → gen.ErrorPage |
| module | `import { Suspense } from "gotsx"` | yes | streaming boundary (server only): <Suspense fallback={…}> ships the fallback with the shell, children render in their own goroutine and stream in |
| module | `import * as ns from` | **no** | namespace import |
| stmt | `const x = ... / let x = ...` | yes | single declaration; type annotated or inferred |
| stmt | `const { a, b = 1 } = obj / const [x, setX] = ...` | yes | destructuring + defaults (primitives, zero-value semantics) |
| stmt | `if / else if / else` | yes | JS truthiness (empty string, 0 are falsy) |
| stmt | `for (const x of xs)` | yes | for-of over arrays |
| stmt | `for (let i = 0; i < n; i++) / while (cond) / break / continue` | yes | classic loops; the Go side uses a real for statement (continue runs the update) |
| stmt | `switch (x) { case a: case b: … break; default: … }` | yes | JS fall-through semantics are preserved (translated to Go fallthrough); switch (true) works |
| stmt | `return / throw` | yes | throw is a panic on the Go side, recovered by the request layer |
| stmt | `try / catch / finally` | yes | full on the client; Go runs only the try and finally bodies |
| stmt | `function f() {} / async function f() {}` | yes | nested functions; async only reaches the JS backend |
| stmt | `do … while / for … in / labeled break` | **no** | rewrite as while, or for-of over Object.keys(obj) |
| stmt | `const a = 1, b = 2` | **no** | multiple declarations in one statement |
| expr | `number / string / template string / true / null / undefined` | yes | number is float64 |
| expr | `[a, ...b] / { a, b: 1, ...c }` | yes | array & object literals (object spread client-only) |
| expr | `a.b / a?.b / a[i] / a?.[i]` | yes | member, optional chaining, index |
| expr | `f(x) / f?.(x) / useState<T>(x)` | yes | call, optional call, explicit type argument |
| expr | `! - + typeof` | yes | unary |
| expr | `+ - * / %   === !==   < > <= >=` | yes | arithmetic, strict equality, comparison |
| expr | `&& \|\| ??` | yes | logical; for primitives ?? ≡ \|\| (zero value = absent); an absent object (find miss, optional field) is falsy and === undefined |
| expr | `a ? b : c` | yes | ternary (node or value) |
| expr | `(a, b) => expr / x => { ... } / async () => ...` | yes | arrow functions; param types inferred from context |
| expr | `x = v / += -= *= /= %= / x++ x-- ++x --x` | yes | assignment to variables, fields, Record keys and array indexes; ++/-- on numbers |
| expr | `x as T / x!` | yes | type assertion, non-null assertion |
| expr | `await x` | yes | client code only |
| expr | `== !=` | **no** | only === / !== allowed |
| expr | `new / class / this / function expression` | **no** | no classes or prototypes |
| expr | `/pattern/gimsu` | yes | regex literal, RE2 subset checked at compile time (no lookaround / backreferences); re.test, s.match/replace/replaceAll/split/search |
| expr | `delete m.key / delete m[key] / Object.hasOwn(m, key)` | yes | Record keys: reads of an absent key give the zero value on both sides; hasOwn tests presence |
| jsx | `<div class="x" id={v} disabled>…</div> / <br />` | yes | elements, string/expression/boolean-shorthand attrs, self-closing |
| jsx | `<></>` | yes | fragment |
| jsx | `onClick={fn} / onInput={(e) => ...}` | yes | events; handler param type any; not generated on the server |
| jsx | `aria-* / data-* / role` | yes | booleans render as "true" / "false" |
| jsx | `{cond && <x/>} / {a ? <x/> : <y/>} / {list.map(...)}` | yes | conditions & lists; reactive only if a signal is read |
| jsx | `{list.map((x) => <li key={x.id}>…</li>)}` | yes | keyed list: the client reuses / moves / disposes DOM per key, so inputs, focus and per-row effects survive reorders; without key the list rebuilds |
| jsx | `<Comp prop={v}>children</Comp>` | yes | component call; children is a Node |
| jsx | `<div {...props} />` | **no** | attribute spread |
| jsx | `passing children to an island` | **no** | island props go through JSON in an HTML attribute — no room for a Node; children of shared & server components are fine |
| jsx | `dangerouslySetInnerHTML` | **no** | no raw-HTML injection hole; render tokens returned by a host module yourself |
| type | `T[] / Array<T> / Record<string, T>` | yes | arrays, maps |
| type | `{ a: string; b?: number; f(x: T): R }` | yes | object types; optional primitive = zero-value semantics |
| type | `interface A extends B, C { … } / type alias` | yes | extends copies the base fields (a same-named field overrides) |
| type | `"a" \| "b" / T \| undefined` | yes | literal union → string; optional |
| type | `(x: T) => R / Promise<T>` | yes | function type; Promise<T> treated as T |
| hooks | `const [x, setX] = useState(init)` | yes | server = initial value; client = signal |
| hooks | `const y = x * 2` | yes | a signal-dependent const is automatically a memo, no useMemo needed |
| hooks | `useMemo(() => ...) / useEffect(() => ...)` | yes | useEffect is client-only, deps tracked automatically |
| builtin | `console.log / JSON.stringify / Math.max min floor ceil round abs sqrt random` | yes | on both sides |
| builtin | `fetch setTimeout document window location history navigator localStorage` | yes | client only, type any |
| builtin | `Object.keys / Object.values` | yes | keys sorted (matching a Go map, keeping hydration stable) |
| builtin | `redirect(url, status?) / notFound()` | yes | server pages only: abort the render and answer with a 3xx / the 404 page (return redirect(…) also works) |
| builtin | `Date.now() / Date.parse(iso) / isoDate(ms)` | yes | milliseconds on both sides; format with fmtDate / isoDate |
| builtin | `Array.from / new Date(...) / Object.assign` | **no** | no constructors; dates are numbers + isoDate |
| array | `length map filter find findIndex some every includes indexOf lastIndexOf join slice concat forEach` | yes | find on a miss yields the zero value, which is falsy and === undefined |
| array | `sort reduce reverse flat at` | yes | sort/reduce match the Go backend (copy, don't mutate) |
| array | `push pop shift unshift splice` | yes | in place, on a variable / field / index (the Go side takes its address); not on a useState array — use setXs([...xs, x]) |
| string | `length toUpperCase toLowerCase trim includes startsWith endsWith split slice replace replaceAll repeat indexOf charAt` | yes | Go side works by rune |
| string | `padStart padEnd trimStart trimEnd lastIndexOf at localeCompare toString / number toFixed toString` | yes | by rune; localeCompare compares by code point on both sides (not locale-aware) |
| string | `matchAll / named groups / sticky regex` | **no** | not in the RE2 subset |

## Islands and reactivity

- `const [n, setN] = useState(0)` — `n` is a signal; every read compiles to `n()` in the browser and to a plain value on the server render.
- A `const` whose initializer reads a signal is automatically a memo; JSX that reads a signal becomes a fine-grained text/attribute binding.
- Component props are **not** reactive (like Solid: read once at creation). Put reactive content in children or a conditional block.
- `useEffect(() => { … }, [deps])` runs in the browser only. `useMemo` is accepted for familiarity; a plain `const` does the same.
- `{cond && <X/>}` / `{cond ? a : b}` and `{list.map(x => <li key={x.id}>…</li>)}` are the conditional and list forms; give island lists a `key` so rows are reused/moved instead of rebuilt (without `key` the whole list re-renders on change).
- Event handlers: `onClick`, `onInput`, `onSubmit`, … (camelCase). `class` and `for` are the attribute names (not `className` / `htmlFor`); `style` is a string.
- Emitting events between islands: `emit("name", payload)` / `on("name", handler)` from `"gotsx"`.
