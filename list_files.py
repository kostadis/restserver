#!/usr/bin/env python3
import os
import datetime
import csv
import argparse

def generate_file_csv(directory_path, output_csv_path):
  """
  Iterates through a directory and its subdirectories, collects detailed file information,
  and writes it to a CSV file.

  Args:
    directory_path: The path to the directory to traverse.
    output_csv_path: The path to the CSV file where the information will be written.
  """
  file_data_list = []
  print(f"Scanning directory: {os.path.abspath(directory_path)}...")
  for root, _, files in os.walk(directory_path):
    for file_name in files:
      # Skip the output CSV file itself if it's in the scanned directory
      if os.path.abspath(os.path.join(root, file_name)) == os.path.abspath(output_csv_path):
          continue

      full_path = os.path.join(root, file_name)
      try:
        file_size = os.path.getsize(full_path)
        modification_time_timestamp = os.path.getmtime(full_path)
        modification_time = datetime.datetime.fromtimestamp(modification_time_timestamp).strftime('%Y-%m-%d %H:%M:%S')

        file_data_list.append([file_name, full_path, file_size, modification_time])
      except FileNotFoundError:
        print(f"Warning: File not found at {full_path}. It might have been deleted during the scan. Skipping.")
      except Exception as e:
        print(f"Warning: Error processing file {full_path}: {e}. Skipping.")

  if not file_data_list:
    print("No files found in the specified directory (or all files were skipped).")
    return

  try:
    with open(output_csv_path, 'w', newline='', encoding='utf-8') as csvfile:
      writer = csv.writer(csvfile)
      # Write the header row
      writer.writerow(["Name", "Path", "Size (Bytes)", "Last Modified"])
      # Write all the collected file data
      writer.writerows(file_data_list)
    print(f"File data successfully written to {os.path.abspath(output_csv_path)}")
  except IOError as e:
    print(f"Error: Could not write to CSV file {output_csv_path}: {e}")
  except Exception as e:
    print(f"Error: An unexpected error occurred while writing the CSV: {e}")


if __name__ == "__main__":
  parser = argparse.ArgumentParser(description="Lists files in a directory and saves details to a CSV file.")
  parser.add_argument("directory_path",
                      help="The path to the directory to scan.")
  parser.add_argument("-o", "--output",
                      default="file_details.csv",
                      help="The name of the output CSV file. Defaults to 'file_details.csv'.")

  args = parser.parse_args()

  # Ensure the directory path exists
  if not os.path.isdir(args.directory_path):
    print(f"Error: Directory not found: {args.directory_path}")
  else:
    generate_file_csv(args.directory_path, args.output)
