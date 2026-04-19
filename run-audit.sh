#!/usr/bin/env bash
set -euo pipefail
go run main.go audit --url https://www.cars24.com/buy-used-cars --max-pages 5000 --max-depth 5 --max-redirects 5 --max-url-length 10000 --max-links-per-page 10000 --max-page-size-kb 50000 --concurrency 20 --sitemap-mode off
