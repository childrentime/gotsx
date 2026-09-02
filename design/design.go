// Package design ships the gotsx UI design system: shared tokens and component classes used by every demo
// and by `gotsx new`. gotsx.css is the Tailwind v4 layer (import it from app/tailwind.css); plain.css is the
// same system as hand-written CSS for apps that do not use Tailwind.
package design

import _ "embed"

//go:embed gotsx.css
var TailwindCSS string

//go:embed plain.css
var PlainCSS string
