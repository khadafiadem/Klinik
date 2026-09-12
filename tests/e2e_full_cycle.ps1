# ===================================================================
#  E2E SIKLUS LENGKAP KLINIK
#  Ambil nomor antrian (kiosk) -> registrasi -> antrean -> pemeriksaan
#  -> rekam medis (diagnosa, tindakan, nyeri) -> resep -> apotek
#  -> tagihan -> selesai periksa (gate bayar) -> pembayaran -> laporan
#
#  Cara pakai:
#    1. Jalankan server lokal:  go run ./cmd/server   (atau .exe yang dibuild)
#    2. PowerShell: .\tests\e2e_full_cycle.ps1
#    3. Opsional: .\tests\e2e_full_cycle.ps1 -BaseUrl "http://localhost:8080" -SkipCleanup
#  Catatan: skrip MEMBUAT data uji (pasien, resep, invoice) lalu menghapusnya
#  otomatis di akhir (kecuali -SkipCleanup). Jalankan dari folder repo.
# ===================================================================

[CmdletBinding()]
param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Username = "admin",
    [string]$Password = "admin123",
    [switch]$SkipCleanup
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot  = Split-Path -Parent $scriptDir

$rand   = Get-Random -Minimum 100000 -Maximum 999999
$work   = Join-Path $env:TEMP "e2e_cycle_$rand"
New-Item -ItemType Directory -Path $work -Force | Out-Null
$cookie = Join-Path $work "cookies.txt"

$PatientName = "E2E Siklus $rand"
$reflect    = "E2E Siklus $rand"
$MasterDoc  = "E2E Dokter Siklus"
$MasterDiag = "E2E Diagnosa Siklus"
$MasterRx   = "E2E Tindakan Siklus"
$MasterMed  = "E2E Obat Paracetamol"
$today      = Get-Date -Format "yyyy-MM-dd"

$passCount = 0
$failCount = 0

function esc($s) {
    return [uri]::EscapeDataString([string]$s)
}

function Assert-Step([string]$name, [bool]$ok, [string]$detail = "") {
    if ($ok) {
        $script:passCount++
        Write-Host "    [PASS] $name$detail" -ForegroundColor Green
    } else {
        $script:failCount++
        Write-Host "    [FAIL] $name$detail" -ForegroundColor Red
    }
}
function Step([string]$name) {
    Write-Host ""
    Write-Host "== $name ==" -ForegroundColor Cyan
}

function Get-LocationFromPost([string]$path, [string]$data) {
    $hd = Join-Path $work ("h_" + [guid]::NewGuid().ToString("N") + ".txt")
    & curl.exe -s -c $cookie -b $cookie -o NUL -D $hd -d $data "$BaseUrl$path" 2>$null
    $loc = ""
    if (Test-Path $hd) {
        $line = Select-String -Path $hd -Pattern '^Location:\s*(.+)$'
        if ($line) { $loc = $line.Matches[0].Groups[1].Value.Trim() }
        Remove-Item $hd -Force
    }
    return $loc
}

function Get-Page([string]$path) {
    return (& curl.exe -s -b $cookie "$BaseUrl$path" 2>$null) -join "`n"
}

function Get-OptionId([string]$html, [string]$selectName, [int]$preferId) {
    $m = [regex]::Match($html, '(?s)<select[^>]*name="' + [regex]::Escape($selectName) + '"[^>]*>(.*?)</select>')
    if (-not $m.Success) { return 0 }
    foreach ($mm in [regex]::Matches($m.Groups[1].Value, '<option value="(\d+)"')) {
        $id = [int]$mm.Groups[1].Value
        if ($id -eq $preferId) { return $id }
    }
    foreach ($mm in [regex]::Matches($m.Groups[1].Value, '<option value="(\d+)"')) {
        return [int]$mm.Groups[1].Value
    }
    return 0
}

function Get-Row([string]$html, [string]$name) {
    foreach ($r in [regex]::Matches($html, '(?s)<tr>(.*?)</tr>')) {
        if ($r.Groups[1].Value -like "*$name*") { return $r.Groups[1].Value }
    }
    return ""
}

function Find-RowStatus([string]$html, [string]$name, [string]$expect) {
    foreach ($r in [regex]::Matches($html, '(?s)<tr>(.*?)</tr>')) {
        if ($r.Groups[1].Value -like "*$name*") {
            if ($expect -eq "DIPANGGIL" -and $r.Groups[1].Value -match '>DIPANGGIL<') { return $true }
            if ($expect -eq "SEDANG_DIPERIKSA" -and $r.Groups[1].Value -match '>SEDANG_DIPERIKSA<') { return $true }
            if ($expect -eq "MENUNGGU_BAYAR" -and $r.Groups[1].Value -match 'Menunggu Pembayaran') { return $true }
            if ($expect -eq "SELESAI" -and $r.Groups[1].Value -match '>SELESAI<') { return $true }
        }
    }
    return $false
}

function Get-RegIDFromRegistrations([string]$html, [string]$name) {
    foreach ($r in [regex]::Matches($html, '(?s)<tr>(.*?)</tr>')) {
        if ($r.Groups[1].Value -like "*$name*") {
            $mm = [regex]::Match($r.Groups[1].Value, '/registrations/(\d+)/cancel')
            if ($mm.Success) { return [int]$mm.Groups[1].Value }
        }
    }
    return 0
}

function Get-QueueIdForAction([string]$html, [string]$name, [string]$action) {
    $row = Get-Row $html $name
    if ($row -eq "") { return 0 }
    $mm = [regex]::Match($row, '/queues/(\d+)/' + [regex]::Escape($action))
    if ($mm.Success) { return [int]$mm.Groups[1].Value }
    return 0
}

function Get-QueueRegID([string]$html, [string]$name) {
    $row = Get-Row $html $name
    if ($row -eq "") { return 0 }
    $mm = [regex]::Match($row, 'registration_id=(\d+)')
    if ($mm.Success) { return [int]$mm.Groups[1].Value }
    return 0
}

function Get-QueueStatus([string]$html, [string]$name) {
    $row = Get-Row $html $name
    if ($row -eq "") { return "" }
    if ($row -match 'Menunggu Pembayaran') { return "MENUNGGU_BAYAR" }
    if ($row -match '>SELESAI<') { return "SELESAI" }
    if ($row -match '>DIPANGGIL<') { return "DIPANGGIL" }
    if ($row -match '>SEDANG_DIPERIKSA<') { return "SEDANG_DIPERIKSA" }
    if ($row -match 'MENUNGGU') { return "MENUNGGU" }
    return ""
}

function Get-IdFromLocation([string]$loc) {
    $mm = [regex]::Match($loc, '/(\d+)\s*$')
    if ($mm.Success) { return [int]$mm.Groups[1].Value }
    return 0
}

function Get-MonitorQueues() {
    $json = & curl.exe -s -b $cookie "$BaseUrl/api/queue/monitor" 2>$null
    try {
        return (($json | ConvertFrom-Json).queues)
    } catch {
        return @()
    }
}

function Get-StatusByQID([int]$qid) {
    for ($i = 0; $i -lt 6; $i++) {
        foreach ($qq in (Get-MonitorQueues)) {
            if ([int]$qq.id -eq $qid) { return [string]$qq.status }
        }
        Start-Sleep -Milliseconds 300
    }
    return ""
}

Write-Host "======================================================"
Write-Host "  E2E SIKLUS LENGKAP (BaseUrl=$BaseUrl, data: $PatientName)"
Write-Host "======================================================"

# ---- Login ----
Step "1. Login"
$code = & curl.exe -s -c $cookie -b $cookie -d "username=$Username&password=$Password" -o NUL -w "%{http_code}" "$BaseUrl/login" 2>$null
$dash = Get-Page "/dashboard"
Assert-Step "Login $Username" (($code -eq "303") -and ($dash -match 'Dashboard'))

# ---- Seed master data via tool (dokter/diagnosa/tindakan/obat E2E) ----
Step "2. Seed master data (dokter, diagnosa, tindakan, obat)"
Push-Location $repoRoot
& go build -o "$work\e2e_seed.exe" ./tests/e2e_seed 2>$null
if ($LASTEXITCODE -ne 0) {
    Pop-Location
    Assert-Step "Build tool seed" $false " (go build gagal)"
} else {
    $seedRaw = & "$work\e2e_seed.exe" seed 2>$null | Select-Object -Last 1
    Pop-Location
    $master = $seedRaw | ConvertFrom-Json
    Assert-Step "Seed master OK" ($null -ne $master -and [int]$master.doctor_id -gt 0) " (doctor=$($master.doctor_id), diag=$($master.diagnosis_id), treat=$($master.treatment_id), obat=$($master.medicine_id))"
}

# ---- Ambil nomor antrian kiosk ----
Step "3. Ambil nomor antrian (kiosk)"
$kioskJson = & curl.exe -s -X POST "$BaseUrl/api/queue/take" 2>$null
$kiosk = $kioskJson | ConvertFrom-Json
$kioskID = 0
$kioskNumber = ""
if ($kiosk.success) {
    $kioskID = [int]$kiosk.queue.id
    $kioskNumber = [string]$kiosk.queue.queue_number
}
Assert-Step "Nomor antrian diambil" ($kiosk.success -and $kioskID -gt 0) " ($kioskNumber, kioskID=$kioskID)"

# ---- Buat pasien ----
Step "4. Buat pasien baru"
$data = "full_name=$(& esc $PatientName)&nik=32$rand&gender=LAKI_LAKI&date_of_birth=1990-01-15&blood_type=O&phone=0812$rand&email=e2e_$rand@klinik.test&address=Jl. Uji Coba No. 7&city=Jakarta&province=DKI Jakarta&postal_code=12345&insurance_name=UMUM"
$loc = Get-LocationFromPost "/patients/save" $data
$page = Get-Page "/patients?search=$(& esc $PatientName)"
$patientID = Get-Row $page $PatientName | ForEach-Object { $mm = [regex]::Match($_, '/patients/(\d+)/edit'); if ($mm.Success) { [int]$mm.Groups[1].Value } }
Assert-Step "Pasien terbuat" ($patientID -gt 0) " (PatientID=$patientID)"

# ---- Registrasi + tautkan nomor kiosk ----
Step "5. Registrasi pasien + tautkan nomor kiosk"
$page = Get-Page "/registrations/new"
$doctorID = Get-OptionId $page "doctor_id" ([int]$master.doctor_id)
Assert-Step "Dokter tersedia" ($doctorID -gt 0) " (DoctorID=$doctorID)"
$data = "patient_id=$patientID&doctor_id=$doctorID&registration_date=$today&registration_type=UMUM&complaint=$(& esc "E2E siklus lengkap - $PatientName")&notes=E2E&kiosk_id=$kioskID"
$loc = Get-LocationFromPost "/registrations/save" $data
$page = Get-Page "/queues"
$qid   = Get-QueueIdForAction $page $PatientName "call"
$regPage = Get-Page "/registrations"
$regID = Get-RegIDFromRegistrations $regPage $PatientName
Assert-Step "Registrasi terdata & antrian dibuat" ($qid -gt 0 -and $regID -gt 0) " (RegID=$regID, QueueID=$qid)"

# ---- Panggil & mulai periksa ----
Step "6. Panggil & mulai periksa"
& curl.exe -s -b $cookie -o NUL "$BaseUrl/queues/$qid/call" 2>$null
Assert-Step "Antrian DIPANGGIL" ((Get-StatusByQID $qid) -eq "DIPANGGIL") " (qid=$qid -> $(Get-StatusByQID $qid))"
& curl.exe -s -b $cookie -o NUL "$BaseUrl/queues/$qid/start" 2>$null
Assert-Step "Antrian mulai diperiksa" ((Get-StatusByQID $qid) -eq "SEDANG_DIPERIKSA") " (qid=$qid -> $(Get-StatusByQID $qid))"

# ---- Rekam medis ----
Step "7. Buat rekam medis (pemeriksaan)"
$data = "patient_id=$patientID&doctor_id=$doctorID&registration_id=$regID&examination_date=$today&chief_complaint=$(& esc "Demam dan batuk - $PatientName")&temperature=37.5&heart_rate=80&blood_pressure=120/80&respiratory_rate=20&weight=60&height=170&anamnesis=$(& esc "E2E anamnesis")&physical_examination=$(& esc "E2E pemeriksaan fisik")&notes=$(& esc "E2E catatan rm")"
$loc = Get-LocationFromPost "/medical-records/save" $data
$mrID = Get-IdFromLocation $loc
Assert-Step "Rekam medis dibuat" ($mrID -gt 0) " (MR=$mrID)"

# ---- Diagnosa + tindakan + nyeri ----
Step "8. Diagnosa, tindakan, dan assesmen nyeri"
$page = Get-Page "/medical-records/$mrID"
$diagID = Get-OptionId $page "diagnosis_id" ([int]$master.diagnosis_id)
$treatID = Get-OptionId $page "treatment_id" ([int]$master.treatment_id)
Get-LocationFromPost "/medical-records/diagnosis/add" "medical_record_id=$mrID&diagnosis_id=$diagID&diagnosis_type=UTAMA" | Out-Null
Get-LocationFromPost "/medical-records/treatment/add" "medical_record_id=$mrID&treatment_id=$treatID&cost=50000" | Out-Null
Get-LocationFromPost "/medical-records/pain/save" "medical_record_id=$mrID&location=Kepala&quality=TAJAM&intensity=5&onset_duration=2 hari&aggravating_factors=Aktivitas&relieving_factors=Istirahat&notes=E2E" | Out-Null
$page = Get-Page "/medical-records/$mrID"
Assert-Step "Diagnosa tersimpan" ($page -match [regex]::Escape($MasterDiag))
Assert-Step "Tindakan tersimpan" ($page -match [regex]::Escape($MasterRx))
Assert-Step "Nyeri terisi (5/10)" ($page -match '5\s*/\s*10')

# ---- Resep + apotek ----
Step "9. Resep, item obat, dan proses apotek"
$loc = Get-LocationFromPost "/medical-records/prescription/create" "medical_record_id=$mrID"
$rxID = Get-IdFromLocation $loc
Assert-Step "Resep dibuat" ($rxID -gt 0) " (RX=$rxID)"
$page = Get-Page "/prescriptions/$rxID"
$medID = Get-OptionId $page "medicine_id" ([int]$master.medicine_id)
Assert-Step "Obat tersedia di resep" ($medID -gt 0) " (MedicineID=$medID)"
Get-LocationFromPost "/prescriptions/item/add" "prescription_id=$rxID&medicine_id=$medID&quantity=2&dosage=1x1&frequency=3x sehari&duration=5 hari&instructions=Sesudah makan" | Out-Null
& curl.exe -s -b $cookie -o NUL "$BaseUrl/prescriptions/$rxID/process" 2>$null
& curl.exe -s -b $cookie -o NUL "$BaseUrl/prescriptions/$rxID/complete" 2>$null
$page = Get-Page "/prescriptions/$rxID"
Assert-Step "Resep diproses apotek (COMPLETED)" ($page -match 'COMPLETED')
Assert-Step "Item obat muncul di resep" ($page -match [regex]::Escape($MasterMed))

# ---- Tagihan ----
Step "10. Buat tagihan (jasa + obat)"
$loc = Get-LocationFromPost "/invoices/save" "patient_id=$patientID&registration_id=$regID&medical_record_id=$mrID&invoice_date=$today&notes=E2E"
$invID = Get-IdFromLocation $loc
Assert-Step "Invoice dibuat" ($invID -gt 0) " (INV=$invID)"
Get-LocationFromPost "/invoices/item/add" "invoice_id=$invID&description=Biaya Konsultasi Dokter&quantity=1&unit_price=100000&item_type=JASA" | Out-Null
Get-LocationFromPost "/invoices/item/add" "invoice_id=$invID&description=Obat Paracetamol 500mg&quantity=2&unit_price=5000&item_type=OBAT" | Out-Null
$page = Get-Page "/invoices/$invID"
Assert-Step "Item tagihan tercatat" ($page -match 'Biaya Konsultasi') $null

# ---- Gate pembayaran: selesai periksa ----
Step "11. Selesai periksa -> gate pembayaran"
& curl.exe -s -b $cookie -o NUL "$BaseUrl/queues/$qid/complete" 2>$null
Assert-Step "Antrian MENUNGGU_BAYAR (belum lunas)" ((Get-StatusByQID $qid) -eq "MENUNGGU_BAYAR") " (qid=$qid -> $(Get-StatusByQID $qid))"

# ---- Pembayaran ----
Step "12. Pembayaran (CASH Rp 110.000)"
$loc = Get-LocationFromPost "/payments/save" "invoice_id=$invID&payment_date=$today&amount=110000&payment_method=CASH&reference_number=REF-$rand&notes=E2E"
Assert-Step "Antrian otomatis SELESAI setelah bayar" ((Get-StatusByQID $qid) -eq "SELESAI") " (qid=$qid -> $(Get-StatusByQID $qid))"
$page = Get-Page "/invoices/$invID"
Assert-Step "Invoice lunas" ($page -match 'SUDAH_BAYAR')

# ---- Laporan ----
Step "13. Verifikasi laporan jasa medis"
$page = Get-Page "/reports/medical-services?from=$today&to=$today"
Assert-Step "Laporan jasa medis memuat dokter E2E" ($page -match [regex]::Escape($MasterDoc))

# ---- Hasil ----
Write-Host ""
Write-Host "======================================================"
Write-Host " HASIL: $passCount PASS, $failCount FAIL"
Write-Host "======================================================"
$failed = ($failCount -gt 0)
Write-Host " Data uji: $PatientName (PatientID=$patientID, RegID=$regID, MR=$mrID, RX=$rxID, INV=$invID)" -ForegroundColor Yellow

# ---- Cleanup otomatis ----
if (-not $SkipCleanup) {
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $ids = [ordered]@{
            patient_id       = [int]$patientID
            registration_id  = [int]$regID
            medical_record_id = [int]$mrID
            prescription_id  = [int]$rxID
            invoice_id       = [int]$invID
            kiosk_id         = [int]$kioskID
            doctor_id        = [int]$master.doctor_id
            diagnosis_id     = [int]$master.diagnosis_id
            treatment_id     = [int]$master.treatment_id
            medicine_id      = [int]$master.medicine_id
        }
        $idsFile = Join-Path $work "ids.json"
        [System.IO.File]::WriteAllText($idsFile, ($ids | ConvertTo-Json), (New-Object System.Text.UTF8Encoding($false)))
        Write-Host ""
        Write-Host "Membersihkan data uji..." -ForegroundColor Yellow
        if (Test-Path "$work\e2e_seed.exe") {
            & "$work\e2e_seed.exe" cleanup $idsFile 2>$null
        } else {
            Push-Location $repoRoot
            & go run ./tests/e2e_seed cleanup $idsFile 2>$null
            Pop-Location
        }
    } else {
        Write-Host "  (Go tidak ditemukan - data uji tidak otomatis dibersihkan)" -ForegroundColor Yellow
    }
} else {
    Write-Host " (-SkipCleanup - data uji TIDAK dihapus)" -ForegroundColor Yellow
}

exit $(if ($failed) { 1 } else { 0 })