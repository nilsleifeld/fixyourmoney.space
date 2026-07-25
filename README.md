# FixYourMoney.space

FixYourMoney.space is a website designed to inspire reflection and encourage people to change their perspective. It invites visitors to explore what money really is and which qualities good money should have.

## Development

```sh
go run . serve
```

The development server rebuilds after source changes and reloads connected browser tabs. Use verbose mode to show every changed file and build phase:

```sh
go run . serve -verbose
```

## Production build

```sh
go run . build
```

The build output is concise, structured plain text suitable for local terminals and CI logs. For detailed diagnostics:

```sh
go run . build -verbose
```
