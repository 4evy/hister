# Run on a disposable, elevated Windows CI runner.
$ErrorActionPreference = 'Stop'
$exe = Join-Path $env:ProgramFiles 'Nogon Service Test\nogon.exe'
$data = Join-Path $env:ProgramData 'Nogon'
if ((Get-Service Nogon -ErrorAction SilentlyContinue) -or (Test-Path $data) -or (Test-Path $exe)) {
    throw 'Windows service test requires a fresh runner'
}
New-Item -ItemType Directory (Split-Path $exe) | Out-Null
go build -o $exe .
if ($LASTEXITCODE) { throw 'Go build failed' }

function Invoke-Nogon([int]$code, [string[]]$arguments) {
    $output = & $exe @arguments
    if ($LASTEXITCODE -ne $code) { throw "nogon $arguments returned $LASTEXITCODE, expected $code" }
    $output
}
Set-Location (Split-Path $exe)
"server:`n  address: 127.0.0.1:14599" | Set-Content -Encoding utf8 'test config.yml'
Invoke-Nogon 4 @('service', 'status')
# Refuse an existing machine directory that inherits shared permissions.
New-Item -ItemType Directory $data | Out-Null
Invoke-Nogon 1 @('service', 'install', '--config', 'test config.yml')
Invoke-Nogon 4 @('service', 'status')
if ((Get-Acl $data).AreAccessRulesProtected) { throw 'Rewrote existing permissions' }
Remove-Item -LiteralPath $data
try {
    Invoke-Nogon 0 @('service', 'install', '--config', 'test config.yml')
    $status = Invoke-Nogon 0 @('service', 'status')
    if ($status -notcontains 'running' -or ($status -join ' ') -notlike '*NT SERVICE\Nogon*') { throw "Unexpected status: $status" }
    $null = Invoke-WebRequest 'http://127.0.0.1:14599/'
    $key = Get-FileHash (Join-Path $data '.secret_key')
    $acl = Get-Acl $data
    if (!$acl.AreAccessRulesProtected) { throw 'Machine data inherits permissions' }
    if ($acl.Access.IdentityReference -contains 'BUILTIN\Users') { throw 'Machine data is readable by Users' }
    Invoke-Nogon 1 @('service', 'install')
    Invoke-Nogon 0 @('service', 'stop')
    Invoke-Nogon 3 @('service', 'status')
    sc.exe config Nogon obj= 'NT AUTHORITY\LocalService'
    if ($LASTEXITCODE) { throw 'Could not change the service account' }
    Invoke-Nogon 0 @('service', 'start')
    Invoke-Nogon 0 @('service', 'restart')
    $status = Invoke-Nogon 0 @('service', 'status')
    if (($status -join ' ') -notlike '*LocalService*') { throw "Lost the configured account: $status" }
    $null = Invoke-WebRequest 'http://127.0.0.1:14599/'
} finally {
    Invoke-Nogon 0 @('service', 'uninstall')
}
Invoke-Nogon 4 @('service', 'status')
if ((Get-FileHash $key.Path).Hash -ne $key.Hash) { throw 'The service secret changed' }
Invoke-Nogon 0 @('service', 'install', '--config', 'test config.yml')
Invoke-Nogon 0 @('service', 'uninstall')
if ((Get-FileHash $key.Path).Hash -ne $key.Hash) { throw 'Reinstall replaced existing data' }

# Preserve explicit storage even when the service initially lacks access.
$existing = Join-Path (Split-Path $exe) "User's data"
New-Item -ItemType Directory $existing | Out-Null
$env:NOGON_DATA_DIR = $existing
try {
    Invoke-Nogon 1 @('service', 'install', '--config', 'test config.yml')
    Invoke-Nogon 3 @('service', 'status')
    icacls.exe $existing /grant 'NT SERVICE\Nogon:(OI)(CI)M'
    if ($LASTEXITCODE) { throw 'Could not grant access to the existing data directory' }
    Invoke-Nogon 0 @('service', 'start')
    if (!(Test-Path (Join-Path $existing '.secret_key'))) { throw 'Used the wrong index' }
    Invoke-Nogon 0 @('service', 'stop')
    Move-Item -LiteralPath $existing -Destination "$existing offline"
    Invoke-Nogon 1 @('service', 'start')
    if (Test-Path $existing) { throw 'Recreated missing storage' }
} finally {
    Invoke-Nogon 0 @('service', 'uninstall')
    Remove-Item Env:NOGON_DATA_DIR
}

# A service registered by another installer must survive all Nogon commands.
New-Service -Name Nogon -BinaryPathName "$env:SystemRoot\System32\cmd.exe" -StartupType Disabled | Out-Null
try {
    foreach ($action in 'install', 'status', 'start', 'restart', 'stop', 'uninstall') {
        Invoke-Nogon 1 @('service', $action)
    }
    if ((Get-Service Nogon).Status -ne 'Stopped') { throw 'Modified an external service' }
} finally {
    sc.exe delete Nogon
    if ($LASTEXITCODE) { throw 'Could not remove test fixture' }
}
