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

Options: `-o/--output`, `-p/--preset` (`lossless|recommended|extreme`), `--token`, `--api`, `-q/--quiet`.

## Publish

Just tag the module — [pkg.go.dev](https://pkg.go.dev) indexes it automatically:

```bash
git tag v0.1.0 && git push --tags
```

## License

MIT © [MiCompress](https://micompress.com)
