## About
`bannedfunc` checks for banned functions and reports them.

## Usage
`bannedfunc` is already integrated into `golangci-lint`, and this is the recommended way to use it.

To enable the linter, add the following lines to [`.golangci.yml`](./.golangci.example.yml):

```yaml
linters-settings:
  bannedfunc:
    (fmt).Println: 'Disable fmt.Println'
    (fmt).Printf: 'Disable fmt.Printf'
    (github.com/stretchr/testify/require).*: 'Disable github.com/stretchr/testify/require.*'

linters:
  enable:
    - bannedfunc
```
