# Backup otomatis database Neon PostgreSQL
#
# Usage:
#   powershell -File scripts\backup-database.ps1            # jalankan sekali
#   Register via: powershell -File scripts\register-backup-task.ps1
#
# Konfigurasi dibaca dari (urutan prioritas):
#   1. Environment variable BACKUP_DATABASE_URL
#   2. File scripts\.backup-env  ->  BACKUP_DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
#
# Output: backups\neon_YYYYMMDD_HHMMSS.dump (format custom, terkompresi)
# Log   : backups\backup.log

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
$backupDir = Join-Path $projectRoot "backups"
$logFile = Join-Path $backupDir "backup.log"
$pgBin = Join-Path $projectRoot "tools\pg18"
$retentionDays = 14

function Write-Log {
    param([string]$Level, [string]$Message)
    $line = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') [$Level] $Message"
    Add-Content -Path $logFile -Value $line
    if ($env:BACKUP_VERBOSE) { Write-Output $line }
}

if (-not (Test-Path $backupDir)) {
    New-Item -ItemType Directory -Path $backupDir | Out-Null
}

# --- Resolve connection string ---
$connUrl = $env:BACKUP_DATABASE_URL
if (-not $connUrl) {
    $envFile = Join-Path $PSScriptRoot ".backup-env"
    if (Test-Path $envFile) {
        Get-Content $envFile | ForEach-Object {
            if ($_ -match '^\s*BACKUP_DATABASE_URL\s*=\s*(.+)\s*$') {
                $connUrl = $Matches[1].Trim()
            }
        }
    }
}
if (-not $connUrl) {
    Write-Log ERROR "BACKUP_DATABASE_URL tidak ditemukan (env var atau scripts\.backup-env)"
    exit 1
}

# Strip channel_binding (tidak didukung libpq lama) & pastikan sslmode=require
$connUrl = $connUrl -replace '&?channel_binding=require', ''
if ($connUrl -notmatch 'sslmode=') {
    $connUrl = "$connUrl&sslmode=require"
}

# --- Locate pg_dump ---
$pgDump = Join-Path $pgBin "pg_dump.exe"
$pgRestore = Join-Path $pgBin "pg_restore.exe"
foreach ($tool in @($pgDump, $pgRestore)) {
    if (-not (Test-Path $tool)) {
        Write-Log ERROR "Tool tidak ditemukan: $tool"
        exit 1
    }
}

# --- Run backup ---
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$dumpFile = Join-Path $backupDir "neon_$timestamp.dump"

Write-Log INFO "Mulai backup -> $(Split-Path -Leaf $dumpFile)"
& $pgDump "--dbname=$connUrl" --format=custom --file="$dumpFile" 2>&1 |
    ForEach-Object { Write-Log WARN "$_" }

if ($LASTEXITCODE -ne 0) {
    Write-Log ERROR "pg_dump gagal (exit $LASTEXITCODE)"
    if (Test-Path $dumpFile) { Remove-Item $dumpFile -Force }
    exit 1
}

# --- Verify integrity (list archive contents) ---
$null = & $pgRestore --list "$dumpFile" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Log ERROR "Verifikasi integritas GAGAL untuk $dumpFile"
    exit 1
}

$sizeMb = [math]::Round((Get-Item $dumpFile).Length / 1MB, 2)
Write-Log INFO "Backup OK: $(Split-Path -Leaf $dumpFile) ($sizeMb MB) - integritas terverifikasi"

# --- Retention: hapus dump lebih tua dari N hari ---
$cutoff = (Get-Date).AddDays(-$retentionDays)
Get-ChildItem $backupDir -Filter "neon_*.dump" | Where-Object { $_.LastWriteTime -lt $cutoff } | ForEach-Object {
    Remove-Item $_.FullName -Force
    Write-Log INFO "Retention: hapus $($_.Name) (> $retentionDays hari)"
}

exit 0
