[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repositoryRoot = Split-Path -Parent $PSScriptRoot
$env:CONTENT_API_BASE_URL = 'http://localhost:8080/api/v1'
$env:NEXT_PUBLIC_API_BASE_URL = 'http://localhost:8080/api/v1'
$env:NEXT_PUBLIC_SITE_URL = 'http://localhost:3000'

Push-Location (Join-Path $repositoryRoot 'web')
try {
    pnpm dev
    if ($LASTEXITCODE -ne 0) { throw 'Next.js development server failed.' }
}
finally {
    Pop-Location
}
