[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repositoryRoot = Split-Path -Parent $PSScriptRoot

Push-Location (Join-Path $repositoryRoot 'server')
try {
    go generate ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go contract generation failed.' }
}
finally {
    Pop-Location
}

Push-Location $repositoryRoot
try {
    pnpm exec openapi-typescript 'docs/api/openapi.yaml' --output 'web/types/openapi.d.ts'
    if ($LASTEXITCODE -ne 0) { throw 'TypeScript contract generation failed.' }
}
finally {
    Pop-Location
}
