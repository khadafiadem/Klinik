# Registrasi Windows Scheduled Task untuk backup harian
#
# Jalankan SEKALI (tidak perlu admin, task berjalan sebagai user saat ini):
#   powershell -ExecutionPolicy Bypass -File scripts\register-backup-task.ps1
#
# Default: setiap hari 22:30. Ubah $time jika perlu.

$ErrorActionPreference = "Stop"

$taskName = "KlinikApp-DatabaseBackup"
$time = "22:30"
$projectRoot = Split-Path -Parent $PSScriptRoot
$scriptPath = Join-Path $PSScriptRoot "backup-database.ps1"

if (-not (Test-Path $scriptPath)) {
    Write-Error "Skrip backup tidak ditemukan: $scriptPath"
    exit 1
}

$action = New-ScheduledTaskAction -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File `"$scriptPath`"" `
    -WorkingDirectory $projectRoot

$trigger = New-ScheduledTaskTrigger -Daily -At $time

# Jalankan hanya saat daya listrik tersedia (laptop tidur = skip)
$settings = New-ScheduledTaskSettingsSet -StartWhenAvailable `
    -DontStopOnIdleEnd -MultipleInstances IgnoreNew `
    -ExecutionTimeLimit (New-TimeSpan -Hours 2)

Register-ScheduledTask -TaskName $taskName `
    -Action $action -Trigger $trigger -Settings $settings `
    -Description "Backup harian database KlinikApp (Neon PostgreSQL)" `
    -Force | Out-Null

Write-Output "Task '$taskName' terdaftar: harian pukul $time"
Write-Output "Cek status : Get-ScheduledTask -TaskName '$taskName'"
Write-Output "Jalankan manual : Start-ScheduledTask -TaskName '$taskName'"
Write-Output "Log hasil : backups\backup.log"
