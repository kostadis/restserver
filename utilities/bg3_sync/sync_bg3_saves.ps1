# --- CONFIGURATION ---
$ConfigPath = Join-Path $PSScriptRoot "config.json"

if (-Not (Test-Path $ConfigPath)) {
    Write-Host "[ERROR] Configuration file not found! Please run setup_bg3_sync.ps1 first." -ForegroundColor Red
    pause
    exit
}

$ConfigData = Get-Content $ConfigPath -Raw | ConvertFrom-Json
$DeckIP = $ConfigData.DeckIP
$RemotePath = $ConfigData.RemotePath
$LocalDest = $ConfigData.LocalDest

# --- CHOICE PROMPT ---
Write-Host "-----------------------------------------------" -ForegroundColor Cyan
Write-Host "  BG3 UNIVERSAL SYNC (TOP 10)                  " -ForegroundColor Cyan
Write-Host "-----------------------------------------------" -ForegroundColor Cyan
Write-Host " [1] PULL: Deck -> Windows                     "
Write-Host " [2] PUSH: Windows -> Deck                     "
$choice = Read-Host "`nChoose an option (1 or 2)"

# --- EXECUTION LOGIC ---
if ($choice -eq "1") {
    # PULL LOGIC
    Write-Host "`nConnecting to Steam Deck at $DeckIP..." -ForegroundColor Gray
    # basename-per-line without xargs, which would word-split names containing spaces
    $RecentSaveNames = ssh deck@$DeckIP "ls -dt ${RemotePath}/*/ | head -n 10 | rev | cut -d/ -f2 | rev"

    foreach ($SaveName in $RecentSaveNames) {
        $LocalPath = Join-Path $LocalDest $SaveName
        if (!(Test-Path -LiteralPath $LocalPath)) {
            Write-Host "[NEW] Pulling: $SaveName" -ForegroundColor Yellow
            # Modern OpenSSH scp uses the SFTP protocol (no remote shell is invoked), so the
            # remote path must be passed LITERALLY with no shell quoting. Spaces are fine as
            # long as the whole "user@host:path" is a single PowerShell argument (it is here).
            scp -rp "deck@${DeckIP}:${RemotePath}/${SaveName}" "$LocalDest"
        } else {
            Write-Host "[OK]  Already exists locally: $SaveName" -ForegroundColor Gray
        }
    }
}
elseif ($choice -eq "2") {
    # PUSH LOGIC
    Write-Host "`nChecking Windows for the 10 latest saves..." -ForegroundColor Gray
    $RecentLocalSaves = Get-ChildItem -Path $LocalDest | Sort-Object LastWriteTime -Descending | Select-Object -First 10

    foreach ($Save in $RecentLocalSaves) {
        $SaveName = $Save.Name
        # Check if the folder exists on the Deck (single-quote the name for the remote shell)
        $RemoteSave = "${RemotePath}/'" + $SaveName.Replace("'", "'\''") + "'"
        $existsOnDeck = ssh deck@$DeckIP "if [ -d $RemoteSave ]; then echo true; fi"

        if ($existsOnDeck -ne "true") {
            Write-Host "[NEW] Pushing: $SaveName" -ForegroundColor Green
            # We push the local folder to the remote symlink
            scp -rp "$($Save.FullName)" "deck@${DeckIP}:${RemotePath}/"
        } else {
            Write-Host "[OK]  Already exists on Deck: $SaveName" -ForegroundColor Gray
        }
    }
}
else {
    Write-Host "Invalid choice. Exiting." -ForegroundColor Red
}

Write-Host "`nSync Complete!" -ForegroundColor Cyan
pause