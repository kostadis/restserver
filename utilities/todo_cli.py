#!/usr/bin/env python3

import argparse
import json
import sys
import requests

BASE_URL = "http://localhost:8080"

def handle_list_items():
    """Handles listing all items."""
    try:
        response = requests.get(f"{BASE_URL}/items")
        response.raise_for_status()  # Raises an HTTPError for bad responses (4XX or 5XX)

        if not response.content:
            print("No items found or empty response from server.")
            return

        try:
            items = response.json()
        except json.JSONDecodeError:
            print("Error: Could not decode JSON response from server.")
            sys.exit(1)

        if items:
            print("Todo Items:")
            for item in items:
                print(f"  ID: {item.get('id')}")
                print(f"  Name: {item.get('name')}")
                print(f"  Description: {item.get('description', 'N/A')}")
                print(f"  Priority: {item.get('priority')}")
                print("-" * 20)
        else:
            print("No items found.")

    except requests.exceptions.RequestException as e:
        print(f"Error connecting to the server: {e}")
        sys.exit(1)
    except Exception as e: # Catch other potential errors, like JSONDecodeError if server sends malformed JSON on error
        print(f"An unexpected error occurred: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Status Code: {e.response.status_code}")
            print(f"Response Text: {e.response.text}")
        sys.exit(1)


def handle_create_item(args):
    """Handles creating a new item."""
    payload = {
        "name": args.name,
        "priority": args.priority
    }
    if args.description:
        payload["description"] = args.description

    try:
        response = requests.post(f"{BASE_URL}/items", json=payload)

        if response.status_code == 201: # Created
            try:
                item = response.json()
                print("Successfully created item:")
                print(f"  ID: {item.get('id')}")
                print(f"  Name: {item.get('name')}")
                print(f"  Description: {item.get('description', 'N/A')}")
                print(f"  Priority: {item.get('priority')}")
            except json.JSONDecodeError:
                print("Error: Could not decode JSON response from server after creating item.")
                sys.exit(1)
        else:
            print(f"Error creating item. Status Code: {response.status_code}")
            try:
                error_data = response.json()
                print(f"Server error: {error_data.get('error', 'Unknown error')}")
            except json.JSONDecodeError:
                print(f"Server response: {response.text}")
            sys.exit(1)

    except requests.exceptions.RequestException as e:
        print(f"Error connecting to the server: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred: {e}")
        sys.exit(1)

def handle_delete_item(args):
    """Handles deleting an item by its ID."""
    item_id = args.id
    try:
        response = requests.delete(f"{BASE_URL}/items/{item_id}")

        if response.status_code == 204:  # No Content
            print(f"Successfully deleted item with ID: {item_id}")
        elif response.status_code == 404: # Not Found
            print(f"Error: Item with ID {item_id} not found.")
            sys.exit(1)
        else:
            # Attempt to get more details from response for other errors
            print(f"Error deleting item with ID {item_id}. Status Code: {response.status_code}")
            try:
                error_data = response.json()
                print(f"Server error: {error_data.get('error', 'Unknown error')}")
            except json.JSONDecodeError:
                if response.text:
                    print(f"Server response: {response.text}")
                else:
                    print("No additional error information from server.")
            sys.exit(1)

    except requests.exceptions.RequestException as e:
        print(f"Error connecting to the server: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred: {e}")
        sys.exit(1)

def main():
    """Main function to parse arguments and call appropriate handlers."""
    parser = argparse.ArgumentParser(description="A simple CLI tool to interact with a TODO API.")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # List command
    list_parser = subparsers.add_parser("list", help="List all items")

    # Create command
    create_parser = subparsers.add_parser("create", help="Create a new item")
    create_parser.add_argument("--name", required=True, help="Name of the item")
    create_parser.add_argument("--priority", required=True, type=int, help="Priority of the item")
    create_parser.add_argument("--description", help="Description of the item")

    # Delete command
    delete_parser = subparsers.add_parser("delete", help="Delete an item by ID")
    delete_parser.add_argument("id", type=int, help="ID of the item to delete")

    args = parser.parse_args()

    if args.command == "list":
        handle_list_items()
    elif args.command == "create":
        handle_create_item(args)
    elif args.command == "delete":
        handle_delete_item(args)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
