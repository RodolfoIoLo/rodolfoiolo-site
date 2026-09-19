[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repositoryRoot = Split-Path -Parent $PSScriptRoot

$env:APP_ENV = 'development'
$env:HTTP_ADDRESS = ':8080'
# 宿主机端口 5433 —— 5432 被本机原生 PostgreSQL 18 占用（见 compose.dev.yml 注释）。
$env:DATABASE_URL = 'postgres://personal_site:personal_site_dev@localhost:5433/personal_site?sslmode=disable'
$env:DATABASE_MAX_CONNS = '10'
$env:DATABASE_MIN_CONNS = '1'
$env:DATABASE_TIMEOUT = '5s'
$env:PUBLIC_SITE_URL = 'http://localhost:3000'
$env:ALLOWED_ORIGINS = 'http://localhost:3000'

Push-Location (Join-Path $repositoryRoot 'server')
try {
  go run ./cmd/api
  if ($LASTEXITCODE -ne 0) { throw 'API process failed.' }
}
finally {
  Pop-Location
}
