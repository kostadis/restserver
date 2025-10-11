import argparse
import os

def replace_strings(file_path, substitutions):
    """
    Replaces strings in a file based on the provided replacements.
    """
    try:
        with open(file_path, 'r') as f:
            content = f.read()

        for substitution in substitutions:
            parts = substitution.split(':', 1)
            if len(parts) != 2:
                print(f"Warning: Skipping invalid substitution format: {substitution}")
                continue

            replacement = parts[0]
            targets = parts[1].split(',')

            for target in targets:
                content = content.replace(target, replacement)

        with open(file_path, 'w') as f:
            f.write(content)

        print(f"Successfully replaced strings in {file_path}")

    except FileNotFoundError:
        print(f"Error: File not found at {file_path}")
    except Exception as e:
        print(f"An error occurred: {e}")

def main():
    """
    Main function to parse arguments and call the replacement function.
    """
    parser = argparse.ArgumentParser(description="Replace strings in a file.")
    parser.add_argument("-f", "--file", required=True, help="The input file to process.")
    parser.add_argument("-s", "--substitutions", required=True, action='append', help="The substitutions to make, in the format 'replacement:target1,target2,...'")
    args = parser.parse_args()

    replace_strings(args.file, args.substitutions)

if __name__ == "__main__":
    main()
