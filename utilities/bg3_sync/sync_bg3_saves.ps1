# --- CONFIGURATION ---
# Please replace these variables with your own IP and paths before running!
$DeckIP = "YOUR_DECK_IP"
$RemotePath = "~/bg3saves"
$LocalDest = "C:\Users\YOUR_USERNAME\AppData\Local\Larian Studios\Baldur's Gate 3\PlayerProfiles\Public\Savegames\Story"

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
    $RecentSaveNames = ssh deck@$DeckIP "ls -dt ${RemotePath}/*/ | head -n 10 | xargs -n 1 basename"

    foreach ($SaveName in $RecentSaveNames) {
        $LocalPath = Join-Path $LocalDest $SaveName
        if (!(Test-Path -LiteralPath $LocalPath)) {
            Write-Host "[NEW] Pulling: $SaveName" -ForegroundColor Yellow
            scp -rp "deck@${DeckIP}:`"${RemotePath}/${SaveName}`"" "$LocalDest"
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
        # Check if the folder exists on the Deck
        $existsOnDeck = ssh deck@$DeckIP "if [ -d `"${RemotePath}/${SaveName}`" ]; then echo 'true'; fi"

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