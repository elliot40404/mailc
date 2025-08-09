<div align="center">

## mailc – Type‑safe HTML Email Template Compiler for Go

Generate strongly‑typed Go functions from plain HTML email templates. Render subject and body with Go templates, avoid runtime surprises, and keep templates easy to edit.

</div>

---

## Features

- **Type‑safe data models** from annotations in `.html`
- **Single variables supported**: `{{var}}` inferred as `string` if no type hint
- **Optional subject**: functions return `{Subject, HTML}, error`; empty Subject when not provided
- **Normalized identifiers**: `{{User.Name}}` or `{{ .User.Name}}` both work
- **Per‑template types** to avoid collisions across templates
- **Conditional imports**: `text/template` only when subject exists; `time` when `time.Time` used
- **No runtime file I/O**: templates compile to Go code in your repo

---

## Install

```bash
go install github.com/elliot40404/mailc/cmd/mailc@latest
```

Or use the included Just recipes (Windows/macOS/Linux) to build and run via `./bin/mailc`.

---

## How it works (Overview)

Place HTML templates in a directory. Annotate your data model using special HTML comments:

- Subject: `<!-- $Subject: Welcome {{User.Name}} -->`
- Struct: `<!-- @type User -->`
- Field: `<!-- @type User.Name string -->`
- Top‑level variable: `<!-- @type inviteLink string -->`

During generation, mailc:

- Parses subject and body
- Detects `@type` declarations for structs and fields
- Infers undeclared simple variables like `{{username}}` as `string`
- Normalizes top‑level references to `{{ .Field}}`
- Emits a function `NameEmail(*NameEmailData) (NameEmailResult, error)` in `package emails`

---

## Quickstart

1) In your app, create templates under `./emails/` (or `./templates/`). Example `./emails/order_confirmation.html`:

```html
<!-- $Subject: Welcome {{User.Name}} – Order #{{Order.ID}} placed {{Order.CreatedAt}} -->

<!-- @type Order -->
<!-- @type Order.ID int -->
<!-- @type Order.Name string -->
<!-- @type Order.Qty int -->
<!-- @type Order.CreatedAt time.Time -->

<!-- @type User -->
<!-- @type User.Name string -->

<html>
  <body>
    <h1>Welcome, {{User.Name}}!</h1>
    <p>Your order <strong>#{{Order.ID}}</strong> for {{Order.Name}} (x{{Order.Qty}}) was placed at {{Order.CreatedAt}}.</p>
  </body>
</html>
```

2) Generate code into your project’s `internal/emails`:

```bash
mailc generate -input ./emails -output ./internal/emails
```

3) Use the generated package in your code:

```go
import (
  "log"
  "time"
  emails "your/module/internal/emails"
)

func send() error {
  data := &emails.OrderConfirmationEmailData{
    User:  emails.OrderConfirmationEmailUser{Name: "Jane"},
    Order: emails.OrderConfirmationEmailOrder{ID: 42, Name: "Widget", Qty: 3, CreatedAt: time.Now()},
  }
  res, err := emails.OrderConfirmationEmail(data)
  if err != nil { return err }
  // res.Subject and res.HTML now contain your email subject/body
  return nil
}
```

---

## Examples

This repository also ships example templates under `examples/templates/`. You can compile them to `examples/generated/` with:

```bash
just gen-examples
```

Suggested template names (any filename is supported; names are safely converted to exported Go identifiers):

- `welcome_personalized.html` – uses inferred variables like `{{username}}`, `{{firstName}}`
- `account_invite_link.html` – uses a typed top‑level variable `<!-- @type inviteLink string -->`
- `order_confirmation.html` – demonstrates multiple structs and fields
- `welcome_no_subject.html` – no subject block; result `Subject` will be empty

---

## Generated API

For each `name.html`, mailc generates in `package emails`:

- `type NameEmailData struct { ... }` – root input data
- `type NameEmailResult struct { Subject string; HTML string }` – output
- Struct types per template, e.g. `NameEmailUser`, `NameEmailOrder`
- `func NameEmail(data *NameEmailData) (NameEmailResult, error)` – renders subject and HTML

Constant names are unique per file, e.g. `nameEmailHTMLTemplate` and `nameEmailSubjectTemplate`.

---

## Template syntax and annotations

- **Subject (optional)**: `<!-- $Subject: ... -->`
  - If omitted, `Result.Subject` is empty and `text/template` is not imported
- **Top‑level variables**:
  - With hint: `<!-- @type apiKey string -->` → field `APIKey string`
  - Without hint: `{{username}}` or `{{firstName}}` → inferred as `string`
  - Inference recognizes names matching `[A-Za-z][A-Za-z0-9_]*`
- **Structs and fields**:
  - `<!-- @type User -->`, `<!-- @type User.Name string -->`
  - Use Go types (primitives or qualified like `time.Time`)
- **Normalization**:
  - `{{User.Name}}` or `{{ .User.Name}}` both work
  - Top‑level references are normalized to `{{ .Field}}`

---

## File naming guidelines

- Any filename is supported; mailc converts filenames into exported identifiers safely
- Examples:
  - `order_confirmation.html` → `OrderConfirmationEmail`/`OrderConfirmationEmailData`
  - `account-invite-link.html` → `AccountInviteLinkEmail`/...
- Prefer readable names; underscores or hyphens are fine

---

## Watching for changes (live compile)

`mailc` is a one-shot code generator. To recompile on file changes, wrap it with a watcher such as:

- `watchexec --exts html -- mailc generate -input ./emails -output ./internal/emails`
- `watchman-make -p 'emails/*.html' -r 'mailc generate -input ./emails -output ./internal/emails'`
- `nodemon --ext html --exec "mailc generate -input ./emails -output ./internal/emails"`

Add this to your dev workflow so changes to templates automatically regenerate code.

---

## Demo app (from this repo’s examples)

We include a tiny demo that renders one of the example templates and shows how to send using `net/smtp`.

Steps:

```bash
just gen-examples           # builds local mailc and generates examples
go run ./examples/cmd/demo  # renders and attempts to send (requires SMTP env vars)
```

Env vars used by the demo:

- `SMTP_HOST` (e.g. smtp.gmail.com)
- `SMTP_PORT` (e.g. 587)
- `SMTP_USER`
- `SMTP_PASS`
- `SMTP_FROM`
- `SMTP_TO`

Note: Use an app password or a test SMTP server for local development.

---

## CLI

```text
mailc - Type-safe email templates

Usage:
  mailc [command] [flags]

Commands:
  generate   Parse HTML templates and generate Go code
  help       Show help
  version    Show current mailc version

Flags (for generate):
  -input     Directory containing HTML email templates (default: ./emails)
  -output    Directory to write generated Go code (default: ./internal/emails)
```

Just recipes:

- `just build` – build the CLI to `./bin/mailc`
- `just gen-examples` – generate `examples/templates` → `examples/generated`
- `just test` / `just testv` – run tests
- `just lint-fix` – auto-fix lint issues

---

## Do’s and Don’ts

- Do use `@type` to declare structs and fields you reference
- Do rely on simple variable inference for `{{var}}` when you want `string`
- Do use Go types in hints (e.g. `int`, `string`, `time.Time`)
- Don’t expect inference for complex expressions (e.g. `{{ printf ... }}`) or dotted paths
- Don’t put secrets in templates; mailc compiles templates into your binary

---

## Troubleshooting

- Missing variable/field at render time
  - Re‑run generation after template changes
  - Ensure your variable names match `[A-Za-z][A-Za-z0-9_]*`

- Unused import `text/template`
  - Happens only if a subject was present but later removed; re‑generate

- Importing the generated package
  - Generated code defaults to `package emails`. Import it by the location you generated into

---

## Contributing

- Run `just lint-fix` and `just test` before pushing
- Example templates: add under `examples/templates/` and run `just gen-examples`
- Install git hooks locally so you don’t push broken code:

```bash
git config core.hooksPath .githooks
chmod +x .githooks/*
```

---

## License

MIT License. See `LICENSE` for details.