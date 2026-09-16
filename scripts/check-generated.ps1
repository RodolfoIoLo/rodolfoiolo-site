[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repositoryRoot = Split-Path -Parent $PSScriptRoot

& (Join-Path $PSScriptRoot 'generate.ps1')

Push-Location $repositoryRoot
try {
    git diff --exit-code -- 'server/generated/oapi/api.gen.go' 'web/types/openapi.d.ts'
    if ($LASTEXITCODE -ne 0) {
        throw 'Generated contract files are stale. Run scripts/generate.ps1 and commit the result.'
    }
}
finally {
    Pop-Location
}
