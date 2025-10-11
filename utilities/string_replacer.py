import argparse
import os

def replace_strings(file_path, substitutions, output_file=None):
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

        output_path = output_file if output_file else file_path
        with open(output_path, 'w') as f:
            f.write(content)

        if output_file:
            print(f"Successfully wrote changes to {output_path}")
        else:
            print(f"Successfully modified {output_path} in place.")

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
    parser.add_argument("-o", "--output", help="The output file to write to. If not specified, the input file will be modified in place.")
    args = parser.parse_args()

    replace_strings(args.file, args.substitutions, args.output)

if __name__ == "__main__":
    main()
