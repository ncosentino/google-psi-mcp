[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('Start', 'Status', 'Stop', 'Restart', 'Update')]
    [string]$Action = 'Status',

    [Parameter(Mandatory)]
    [string]$BinaryPath,

    [string]$ServiceName = 'google-psi-mcp',

    [string]$ListenAddress = '127.0.0.1',

    [ValidateRange(1, 65535)]
    [int]$Port = 8080,

    [string]$HealthPath = '/health',

    [string[]]$ServerArguments = @(),

    [string]$ReplacementBinaryPath
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$isWindowsPlatform = $PSVersionTable.PSVersion.Major -lt 6 -or $IsWindows

$binary = [IO.Path]::GetFullPath($BinaryPath)
$normalizedListenAddress = $ListenAddress.Trim('[', ']')
$parsedListenAddress = $null
$isLoopbackListener =
    [string]::Equals($normalizedListenAddress, 'localhost', [StringComparison]::OrdinalIgnoreCase) -or
    ([Net.IPAddress]::TryParse(
        $normalizedListenAddress,
        [ref]$parsedListenAddress) -and
        [Net.IPAddress]::IsLoopback($parsedListenAddress))
if (-not $isLoopbackListener) {
    throw 'The service manager only supports loopback listeners. Use a platform service supervisor for network deployments.'
}
$serviceDirectory = Join-Path (
    [Environment]::GetFolderPath([Environment+SpecialFolder]::LocalApplicationData)
) (Join-Path 'shared-mcp-services' $ServiceName)
$statePath = Join-Path $serviceDirectory 'service.json'
$runId = [DateTimeOffset]::UtcNow.ToString('yyyyMMddHHmmssfff')
$stdoutPath = Join-Path $serviceDirectory "stdout-$runId.log"
$stderrPath = Join-Path $serviceDirectory "stderr-$runId.log"
$urlHost = if ($ListenAddress.Contains(':') -and -not $ListenAddress.StartsWith('[')) {
    "[$ListenAddress]"
}
else {
    $ListenAddress
}
$healthUrl = "http://${urlHost}:$Port$HealthPath"
$shutdownUrl = "http://${urlHost}:$Port/shutdown"
$mutexPrefix = if ($isWindowsPlatform) { 'Local\' } else { '' }
$mutexName = $mutexPrefix + 'SharedMcpService-' + ($ServiceName -replace '[^A-Za-z0-9_.-]', '-')

function Get-ServiceHealth {
    try {
        $response = Invoke-RestMethod -Uri $healthUrl -TimeoutSec 2
        if ($response.status -eq 'ok' -and $response.service -eq $ServiceName) {
            return $response
        }
    }
    catch {
    }
    return $null
}

function Read-ServiceState {
    if (-not (Test-Path -LiteralPath $statePath)) {
        return $null
    }
    try {
        return Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
    }
    catch {
        Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
        return $null
    }
}

function ConvertTo-PosixShellLiteral([string]$Value) {
    $singleQuote = [string][char]39
    $doubleQuote = [string][char]34
    $escapedQuote = $singleQuote + $doubleQuote + $singleQuote + $doubleQuote + $singleQuote
    return $singleQuote + $Value.Replace($singleQuote, $escapedQuote) + $singleQuote
}

function New-ShutdownToken {
    $bytes = New-Object byte[] 32
    $generator = [Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $generator.GetBytes($bytes)
    }
    finally {
        $generator.Dispose()
    }
    return [Convert]::ToBase64String($bytes).TrimEnd('=').Replace('+', '-').Replace('/', '_')
}

function Write-ServiceState(
    [Diagnostics.Process]$Process,
    $Health,
    [string]$ShutdownToken
) {
    New-Item -ItemType Directory -Path $serviceDirectory -Force | Out-Null
    if (-not $isWindowsPlatform) {
        & chmod 700 $serviceDirectory
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to restrict permissions on $serviceDirectory."
        }
    }
    [pscustomobject]@{
        serviceName = $ServiceName
        pid = $Process.Id
        processStartedAt = $Process.StartTime.ToUniversalTime()
        binaryPath = $binary
        url = "http://${urlHost}:$Port"
        healthUrl = $healthUrl
        shutdownUrl = $shutdownUrl
        shutdownToken = $ShutdownToken
        version = $Health.version
        stdoutPath = $stdoutPath
        stderrPath = $stderrPath
        startedAt = [DateTimeOffset]::UtcNow
    } | ConvertTo-Json | Set-Content -LiteralPath $statePath -Encoding UTF8
    if (-not $isWindowsPlatform) {
        & chmod 600 $statePath
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to restrict permissions on $statePath."
        }
    }
}

function Start-ManagedService {
    $health = Get-ServiceHealth
    if ($null -ne $health) {
        return $health
    }
    if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) {
        throw "Binary not found: $binary"
    }

    New-Item -ItemType Directory -Path $serviceDirectory -Force | Out-Null
    $arguments = @(
        '--transport', 'http',
        '--listen-address', $ListenAddress,
        '--port', $Port
    ) + $ServerArguments
    $shutdownToken = New-ShutdownToken
    $previousShutdownToken = $env:MCP_SHUTDOWN_TOKEN
    $hadShutdownToken = Test-Path Env:MCP_SHUTDOWN_TOKEN

    try {
        $env:MCP_SHUTDOWN_TOKEN = $shutdownToken
        if ($isWindowsPlatform) {
            $process = Start-Process `
                -FilePath $binary `
                -ArgumentList $arguments `
                -WorkingDirectory (Split-Path -Parent $binary) `
                -RedirectStandardOutput $stdoutPath `
                -RedirectStandardError $stderrPath `
                -WindowStyle Hidden `
                -PassThru
        }
        else {
            $commandArguments = @($binary) + $arguments |
                ForEach-Object { ConvertTo-PosixShellLiteral ([string]$_) }
            $command = 'exec nohup ' +
                ($commandArguments -join ' ') +
                ' >' + (ConvertTo-PosixShellLiteral $stdoutPath) +
                ' 2>' + (ConvertTo-PosixShellLiteral $stderrPath) +
                ' </dev/null'
            $startInfo = [Diagnostics.ProcessStartInfo]::new()
            $startInfo.FileName = '/bin/sh'
            $startInfo.WorkingDirectory = Split-Path -Parent $binary
            $startInfo.UseShellExecute = $false
            $startInfo.ArgumentList.Add('-c')
            $startInfo.ArgumentList.Add($command)
            $process = [Diagnostics.Process]::Start($startInfo)
            if ($null -eq $process) {
                throw 'Failed to launch the service process.'
            }
        }
    }
    finally {
        if ($hadShutdownToken) {
            $env:MCP_SHUTDOWN_TOKEN = $previousShutdownToken
        }
        else {
            Remove-Item Env:MCP_SHUTDOWN_TOKEN -ErrorAction SilentlyContinue
        }
    }

    for ($attempt = 0; $attempt -lt 40; $attempt++) {
        Start-Sleep -Milliseconds 500
        $health = Get-ServiceHealth
        if ($null -ne $health) {
            $process.Refresh()
            if (-not $process.HasExited) {
                try {
                    Write-ServiceState $process $health $shutdownToken
                }
                catch {
                    Stop-Process -Id $process.Id -ErrorAction SilentlyContinue
                    $process.WaitForExit(5000) | Out-Null
                    Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
                    throw
                }
            }
            return $health
        }
        if ($process.HasExited) {
            break
        }
    }

    $health = Get-ServiceHealth
    if ($null -ne $health) {
        return $health
    }
    if (-not $process.HasExited) {
        Stop-Process -Id $process.Id
        $process.WaitForExit(5000) | Out-Null
    }
    Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
    throw "Service failed to start. See $stderrPath"
}

function Stop-ManagedService {
    $state = Read-ServiceState
    if ($null -eq $state) {
        if ($null -ne (Get-ServiceHealth)) {
            throw 'The service is running but has no trusted state file; refusing to stop an unidentified process.'
        }
        return
    }

    $process = Get-Process -Id ([int]$state.pid) -ErrorAction SilentlyContinue
    if ($null -eq $process) {
        Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
        return
    }

    $recordedBinary = [IO.Path]::GetFullPath([string]$state.binaryPath)
    $runningBinary = [IO.Path]::GetFullPath($process.Path)
    $pathComparison = if ($isWindowsPlatform) {
        [StringComparison]::OrdinalIgnoreCase
    }
    else {
        [StringComparison]::Ordinal
    }
    if (-not [string]::Equals(
        $recordedBinary,
        $runningBinary,
        $pathComparison)) {
        throw "PID $($process.Id) does not match the recorded executable; refusing to stop it."
    }
    $recordedStart = ([DateTime]$state.processStartedAt).ToUniversalTime()
    $runningStart = $process.StartTime.ToUniversalTime()
    if ([Math]::Abs(($recordedStart - $runningStart).TotalSeconds) -gt 1) {
        throw "PID $($process.Id) does not match the recorded process start time; refusing to stop it."
    }

    $shutdownTokenProperty = $state.PSObject.Properties['shutdownToken']
    $shutdownUrlProperty = $state.PSObject.Properties['shutdownUrl']
    if ($null -ne $shutdownTokenProperty -and $null -ne $shutdownUrlProperty) {
        try {
            Invoke-RestMethod `
                -Uri ([string]$shutdownUrlProperty.Value) `
                -Method Post `
                -Headers @{ Authorization = "Bearer $($shutdownTokenProperty.Value)" } `
                -TimeoutSec 5 | Out-Null
        }
        catch {
        }
        for ($attempt = 0; $attempt -lt 40; $attempt++) {
            $process.Refresh()
            if ($process.HasExited) {
                Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
                return
            }
            Start-Sleep -Milliseconds 250
        }
    }

    $process.Refresh()
    if ($process.HasExited) {
        Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
        return
    }
    Stop-Process -Id $process.Id
    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        if ($process.HasExited) {
            break
        }
        Start-Sleep -Milliseconds 250
        $process.Refresh()
    }
    if (-not $process.HasExited) {
        throw "Service process $($process.Id) did not stop."
    }
    Remove-Item -LiteralPath $statePath -Force -ErrorAction SilentlyContinue
}

$mutex = [Threading.Mutex]::new($false, $mutexName)
$acquired = $false
try {
    try {
        $acquired = $mutex.WaitOne([TimeSpan]::FromSeconds(30))
    }
    catch [Threading.AbandonedMutexException] {
        $acquired = $true
    }
    if (-not $acquired) {
        throw "Timed out waiting to manage $ServiceName."
    }

    switch ($Action) {
        'Start' {
            Start-ManagedService
        }
        'Status' {
            $health = Get-ServiceHealth
            [pscustomobject]@{
                running = $null -ne $health
                service = $ServiceName
                url = "http://${urlHost}:$Port"
                version = if ($null -eq $health) { $null } else { $health.version }
            }
        }
        'Stop' {
            Stop-ManagedService
        }
        'Restart' {
            Stop-ManagedService
            Start-ManagedService
        }
        'Update' {
            if ([string]::IsNullOrWhiteSpace($ReplacementBinaryPath)) {
                throw 'Update requires -ReplacementBinaryPath.'
            }
            $replacement = [IO.Path]::GetFullPath($ReplacementBinaryPath)
            if (-not (Test-Path -LiteralPath $replacement -PathType Leaf)) {
                throw "Replacement binary not found: $replacement"
            }
            Stop-ManagedService
            Copy-Item -LiteralPath $replacement -Destination $binary -Force
            Start-ManagedService
        }
    }
}
finally {
    if ($acquired) {
        $mutex.ReleaseMutex()
    }
    $mutex.Dispose()
}
