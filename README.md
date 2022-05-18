# ScheduleScraper

## Usage

```sh
go run main.go --help
```

## PCC

```sh
go run . --runschools pcc
```

## ECC

- Uses [Puppeteer](https://pptr.dev) to scrape initial scheduling data. Remaining data is simply fetched. Use [Yarn](https://yarnpkg.com) to install dependencies

- Transforms the scheduling data with [BeautifulSoup](https://www.crummy.com/software/BeautifulSoup). Use [Poetry](https://python-poetry.org) to install dependencies
