# Baldur's Gate 3 Universal Sync (Windows & Steam Deck)

This toolset provides a pair of PowerShell scripts designed to help you synchronize your 10 most recent *Baldur's Gate 3* save files between a Windows PC and a Steam Deck over your local network using SSH and SCP.

## Requirements
*   Windows PC with PowerShell.
*   Steam Deck connected to the same local network.
*   SSH enabled on your Steam Deck (you must have set a `passwd` for the `deck` user).
*   OpenSSH client installed on Windows (usually installed by default on Windows 10/11).

## Scripts

### 1. `setup_bg3_sync.ps1`
Run this script once to configure your environment. It handles the initial setup automatically:
1.  **Configuration:** Prompts you for your Steam Deck's IP address and the path to your local Windows BG3 save directory, and saves them to a local `config.json` file.
2.  **SSH Key:** Generates a new SSH key (if you don't already have one) and copies the public key to your Steam Deck. This allows the sync script to connect automatically without prompting for a password every time.
3.  **Symlink:** Creates a convenient symbolic link on your Steam Deck (`~/bg3saves`) that points to the deep, hidden folder where Proton actually stores the BG3 save data.

### 2. `sync_bg3_saves.ps1`
This is the main synchronization script. Run this whenever you want to transfer saves. It presents a simple menu:
*   **[1] PULL: Deck -> Windows:** Copies the 10 most recent save folders from your Steam Deck to your Windows PC. It skips folders that already exist locally.
*   **[2] PUSH: Windows -> Deck:** Copies the 10 most recent save folders from your Windows PC to your Steam Deck. It skips folders that already exist on the Deck.

## Usage

1.  Open PowerShell.
2.  Navigate to the directory containing these scripts:
    ```powershell
    cd utilities/bg3_sync
    ```
3.  Run the setup script (first time only):
    ```powershell
    .\setup_bg3_sync.ps1
    ```
    *Follow the on-screen prompts. You may be asked to enter your Steam Deck's 'deck' user password so the SSH key can be copied.*
4.  Run the sync script:
    ```powershell
    .\sync_bg3_saves.ps1
    ```

## Manual Configuration (`config.json`)

If you have already configured SSH access between your Windows PC and your Steam Deck, and you have already created the necessary symbolic link on the Deck, you can skip the setup script.

To do this, simply create a file named `config.json` in the same directory as the scripts (`utilities/bg3_sync/config.json`) with the following format:

```json
{
  "DeckIP": "192.168.1.100",
  "RemotePath": "~/bg3saves",
  "LocalDest": "C:\\Users\\YOUR_USERNAME\\AppData\\Local\\Larian Studios\\Baldur's Gate 3\\PlayerProfiles\\Public\\Savegames\\Story"
}
```

*   **`DeckIP`**: The local IP address of your Steam Deck.
*   **`RemotePath`**: The path on the Steam Deck where saves are located. If you used the setup script, this is the symlink `~/bg3saves`. If you are setting this up manually, you can use the symlink you created or the direct Proton path (`/home/deck/.steam/steam/steamapps/compatdata/1086940/pfx/drive_c/users/steamuser/AppData/Local/Larian Studios/Baldur's Gate 3/PlayerProfiles/Public/Savegames/Story`).
*   **`LocalDest`**: The path to your BG3 save folder on Windows. Note that backslashes `\` must be escaped as double backslashes `\\` in JSON.