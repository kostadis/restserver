# --- CONFIGURATION SETUP ---
Write-Host "-----------------------------------------------" -ForegroundColor Cyan
Write-Host "  BG3 SYNC CONFIGURATION SETUP                 " -ForegroundColor Cyan
Write-Host "-----------------------------------------------" -ForegroundColor Cyan

$ConfigPath = Join-Path $PSScriptRoot "config.json"

$DeckIP = Read-Host "Enter your Steam Deck IP Address"
$LocalDest = Read-Host "Enter the path to your local Windows BG3 saves folder"
# Typical path: C:\Users\YOUR_USERNAME\AppData\Local\Larian Studios\Baldur's Gate 3\PlayerProfiles\Public\Savegames\Story

$RemotePath = "~/bg3saves"
$DeckSavePath = "/home/deck/.steam/steam/steamapps/compatdata/1086940/pfx/drive_c/users/steamuser/AppData/Local/Larian Studios/Baldur's Gate 3/PlayerProfiles/Public/Savegames/Story"

$ConfigData = @{
    DeckIP = $DeckIP
    RemotePath = $RemotePath
    LocalDest = $LocalDest
}

$ConfigData | ConvertTo-Json | Out-File -FilePath $ConfigPath -Encoding utf8
Write-Host "[OK] Configuration saved to $ConfigPath" -ForegroundColor Green


# --- SSH & SYMLINK SETUP ---
Write-Host "`n-----------------------------------------------" -ForegroundColor Cyan
Write-Host "  BG3 SYNC SETUP                               " -ForegroundColor Cyan
Write-Host "-----------------------------------------------" -ForegroundColor Cyan

# 1. SSH KEY SETUP
Write-Host "`n[1] Checking local SSH keys..." -ForegroundColor Gray
$SshDir = Join-Path $env:USERPROFILE ".ssh"
$KeyPath = Join-Path $SshDir "id_rsa"

if (-Not (Test-Path $KeyPath)) {
    Write-Host "Generating new SSH key pair..." -ForegroundColor Yellow
    if (-Not (Test-Path $SshDir)) { New-Item -ItemType Directory -Path $SshDir | Out-Null }
    # Generate key without passphrase for automated syncing
    ssh-keygen -t rsa -b 4096 -f $KeyPath -N '""'
} else {
    Write-Host "[OK] SSH key already exists." -ForegroundColor Green
}

Write-Host "`nCopying SSH key to Steam Deck ($DeckIP)..." -ForegroundColor Gray
Write-Host "You may be prompted for your Steam Deck 'deck' user password." -ForegroundColor Yellow
# Using ssh to create .ssh directory and append key to authorized_keys
$PubKey = Get-Content "$KeyPath.pub"
ssh deck@$DeckIP "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo `"$PubKey`" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"

if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] SSH key successfully copied!" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to copy SSH key. Please check your IP and password." -ForegroundColor Red
    pause
    exit
}

# 2. SYMLINK SETUP
Write-Host "`n[2] Setting up remote symlink..." -ForegroundColor Gray
# Note: Path spaces require careful quoting on the remote side
$CreateSymlinkCmd = "if [ ! -L $RemotePath ]; then ln -s `"$DeckSavePath`" $RemotePath; echo 'Symlink created'; else echo 'Symlink already exists'; fi"

$SymlinkResult = ssh deck@$DeckIP $CreateSymlinkCmd
Write-Host "[INFO] $SymlinkResult" -ForegroundColor Yellow

if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Symlink setup complete!" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to set up symlink." -ForegroundColor Red
}

Write-Host "`nSetup Complete! You can now run sync_bg3_saves.ps1" -ForegroundColor Cyan
pause