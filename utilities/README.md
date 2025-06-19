# File Listing Script

## Description
This script scans a specified directory (and its subdirectories) and outputs detailed information about each file found into a CSV file. The information includes the file's name, full path, size in bytes, and last modification timestamp.

## Usage

First, ensure the script is executable:
```bash
chmod +x list_files.py
```

Then, run the script with the following command structure:
```bash
./list_files.py <directory_to_scan> [options]
```

**Example:**
To scan a directory named `my_documents` in the current location and save the output to `report.csv`:
```bash
./list_files.py ./my_documents -o report.csv
```

If the output file is not specified, it will default to `file_details.csv` in the directory where the script is run.

## Command-line Arguments

-   **`directory_path`** (Positional Argument)
    -   Description: The path to the directory that you want to scan.
    -   Example: `/path/to/your/photos`, `../project_files`, `.`

-   **`-o <filename>`, `--output <filename>`** (Optional Argument)
    -   Description: Specifies the name for the output CSV file.
    -   Default: `file_details.csv`
    -   Example: `-o scan_results.csv`, `--output data/archive_list.csv`

## Output CSV Format

The script generates a CSV file with the following columns:

-   **`Name`**: The name of the file (e.g., `document.txt`).
-   **`Path`**: The full absolute or relative path to the file (e.g., `/home/user/documents/document.txt` or `./subdir/image.jpg`).
-   **`Size (Bytes)`**: The size of the file in bytes (e.g., `1024`).
-   **`Last Modified`**: The date and time the file was last modified, in `YYYY-MM-DD HH:MM:SS` format (e.g., `2023-10-27 14:35:10`).
