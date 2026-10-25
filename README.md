# micompress (Go)

> Compress **images, PDFs, video, audio, documents & archives** from the command line.
> Powered by [**micompress.com**](https://micompress.com/?utm_source=go&utm_medium=cli&utm_campaign=extend).

Zero dependencies (standard library only).

```bash
go install github.com/micompress/micompress-go/cmd/micompress@latest
micompress photo.jpg
```

## Usage

```bash
micompress -p extreme photo.png            # flags before files
micompress -o out/ report.pdf slides.pptx
```

Options: `-o/--output`, `-p/--preset` (`lossless|recommended|extreme`), `--api-key` (personal API key, see below), `--token` (legacy login token), `--suffix`, `--api`, `-q/--quiet`.

## API key

Without a key the CLI works anonymously, with small file-size limits. For larger files and a
monthly allowance, pass your personal API key (`mc_…`) — sign up for free and copy it from
[micompress.com/account](https://micompress.com/account).

```bash
micompress --api-key mc_xxx photo.jpg
# or
export MICOMPRESS_API_KEY=mc_xxx
micompress photo.jpg
```

`--api-key` takes precedence over `MICOMPRESS_API_KEY`. The key is sent as an `X-API-Key` header on
the compress request only.

| Plan | Compressions per month |
| --- | --- |
| Anonymous (no key) | works, small file-size limits |
| Free account | 50 via the API, files up to 100 MB |
| Plus ($9.99/mo) | 1,000 (shared with website usage) |
| Pro ($19.99/mo) | 5,000 (shared with website usage) |

When the allowance is used up the CLI prints the server's message (`API_QUOTA_REACHED` for the free
key, `USAGE_LIMIT_REACHED` for paid plans). See [micompress.com/plans](https://micompress.com/plans).

The older `--token <jwt>` (login token) option is still supported.

## Publish

Just tag the module — [pkg.go.dev](https://pkg.go.dev) indexes it automatically:

```bash
git tag v0.1.2 && git push --tags   # keep in sync with `version` in cmd/micompress/main.go
```

## License

MIT © [MiCompress](https://micompress.com)
