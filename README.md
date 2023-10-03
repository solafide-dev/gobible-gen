# GoBible Generator

This repo contains versions of the Bible in the [Go Bible](https://github.com/solafide-dev/gobible) format, as well as a generator for creating new versions.

Versions contained in this repo are public domain. 

If you use the tools to generate another bible, please be aware that the bible text of various translations are copyrighted and not freely distributable with your project.

Additionally, this tool relies on [BibleGateway.com](https://www.biblegateway.com/) to fetch the text of the bible. It uses DOM scraping, and thus could break at any time with changes to the site. If you find that a version is no longer generating, please open an issue.

### Todo

- [ ] Support formatting
- [ ] Support footnotes
- [ ] Detect Language

## Pregenerated Bibles

Bibles in the public domain are already included, and be be found in the [/generated](https://github.com/solafide-dev/gobible-gen/tree/master/generated) directory.

Feel free to just download these directly if you don't want to generate your own.

## Usage

If you need a version not already included, you can generate it yourself.

```bash
$ go run main.go --version=ESV
```

This will drop a file named `ESV.json` in the `generated` directory.

## Available Versions

This tool uses BibleGateway.com to fetch the text of the bible. You can view their full list of available versions at [https://www.biblegateway.com/versions/](https://www.biblegateway.com/versions/)