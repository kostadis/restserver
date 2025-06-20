#!/usr/bin/env python3

import argparse
import json
import sys
import os

# Add the generated client library to sys.path
# Correctly determine the path to 'utilities/api_client' relative to 'utilities/todo_cli.py'
# __file__ is 'utilities/todo_cli.py'
# os.path.dirname(__file__) is 'utilities'
# os.path.join(os.path.dirname(__file__), 'api_client') is 'utilities/api_client'
# For sys.path, we need the parent of 'utilities' if api_client is 'utilities/api_client'
# No, if the generated client's root package 'todo_api_client' is directly inside 'utilities/api_client',
# then 'utilities/api_client' is the correct path to add.

CLIENT_LIB_PATH = os.path.abspath(os.path.join(os.path.dirname(__file__), 'api_client'))
sys.path.insert(0, CLIENT_LIB_PATH)

from todo_api_client.api.default_api import DefaultApi
from todo_api_client.models.item import Item
from todo_api_client.models.new_item import NewItem
from todo_api_client.models.update_item import UpdateItem
# from todo_api_client.models.error import Error as ApiError # Error model might not be directly used for exceptions
from todo_api_client.exceptions import ApiException, NotFoundException, ServiceException, ForbiddenException, UnauthorizedException, BadRequestException
import todo_api_client


CONFIG_FILE_PATH = os.path.join(os.path.dirname(__file__), 'config.json')

def load_config():
    try:
        with open(CONFIG_FILE_PATH, 'r') as f:
            config = json.load(f)
        return config["BASE_URL"]
    # Error handling will be added in the next step as per instructions
    except FileNotFoundError:
        print(f"Error: Configuration file '{CONFIG_FILE_PATH}' not found.")
        sys.exit(1)
    except json.JSONDecodeError:
        print(f"Error: Could not decode JSON from '{CONFIG_FILE_PATH}'.")
        sys.exit(1)
    except KeyError:
        print(f"Error: 'BASE_URL' not found in '{CONFIG_FILE_PATH}'.")
        sys.exit(1)

BASE_URL = load_config()

# --- API Client Setup ---
configuration = todo_api_client.Configuration(host=BASE_URL)
# Optional: Add authentication, proxies, SSL settings etc. here if needed
# configuration.api_key['bearerAuth'] = 'YOUR_TOKEN' # Example for bearer token

api_client_instance = todo_api_client.ApiClient(configuration)
api = DefaultApi(api_client_instance)
# --- End API Client Setup ---

def _print_api_exception_details(e: ApiException):
    """Helper to print details from an ApiException."""
    print(f"API Error: {e.status} {e.reason}")
    if e.body:
        try:
            error_details = json.loads(e.body)
            # The actual structure of error_details depends on your API's error response
            if isinstance(error_details, dict):
                 print(f"Details: {error_details.get('error', error_details.get('message', e.body))}")
            else: # if the error body is a simple string (e.g. "Item not found")
                print(f"Details: {e.body}")
        except json.JSONDecodeError:
            print(f"Details (raw): {e.body}")

def handle_list_items(args):
    """Handles listing items. If an ID is provided, lists that specific item."""
    if args.id is not None:
        try:
            item_obj = api.get_item_by_id(id=args.id) # Corresponds to operationId: getItemById
            print(f"Todo Item (ID: {item_obj.id}):")
            print(f"  ID: {item_obj.id}")
            print(f"  Name: {item_obj.name}")
            description = 'N/A'
            if hasattr(item_obj, 'description') and item_obj.description is not None:
                description = item_obj.description
            print(f"  Description: {description}")
            print(f"  Priority: {item_obj.priority}")
        except NotFoundException as e:
            print(f"Error: Item with ID {args.id} not found.")
            _print_api_exception_details(e)
            sys.exit(1)
        except ApiException as e:
            _print_api_exception_details(e)
            sys.exit(1)
        except Exception as e:
            print(f"An unexpected error occurred: {e}")
            sys.exit(1)
    else:
        try:
            items = api.get_items() # Corresponds to operationId: getItems
            if items:
                print("Todo Items:")
                for item_obj in items: # item_obj is now an instance of the Item model
                    print(f"  ID: {item_obj.id}")
                    print(f"  Name: {item_obj.name}")
                    # Handle optional description:
                    description = 'N/A'
                    if hasattr(item_obj, 'description') and item_obj.description is not None:
                        description = item_obj.description
                    print(f"  Description: {description}")
                    print(f"  Priority: {item_obj.priority}")
                    print("-" * 20)
            else:
                print("No items found.")
        except ApiException as e:
            _print_api_exception_details(e)
            sys.exit(1)
        except Exception as e:
            print(f"An unexpected error occurred: {e}")
            sys.exit(1)


def handle_create_item(args):
    """Handles creating a new item."""
    # Create an instance of the NewItem model
    new_item_payload = NewItem(name=args.name, priority=args.priority)
    if args.description:
        new_item_payload.description = args.description

    try:
        # Call the create_item method (corresponds to operationId: createItem)
        created_item = api.create_item(new_item=new_item_payload)
        print("Successfully created item:")
        print(f"  ID: {created_item.id}")
        print(f"  Name: {created_item.name}")
        description = 'N/A'
        if hasattr(created_item, 'description') and created_item.description is not None:
            description = created_item.description
        print(f"  Description: {description}")
        print(f"  Priority: {created_item.priority}")
    except BadRequestException as e: # Example of specific exception handling
        print(f"Error creating item (Invalid Request):")
        _print_api_exception_details(e)
        sys.exit(1)
    except ApiException as e:
        _print_api_exception_details(e)
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred: {e}")
        sys.exit(1)

def handle_delete_item(args):
    """Handles deleting an item by its ID."""
    item_id = args.id
    try:
        # Call the delete_item_by_id method (corresponds to operationId: deleteItemById)
        api.delete_item_by_id(id=item_id)
        print(f"Successfully deleted item with ID: {item_id}")
    except NotFoundException as e:
        print(f"Error: Item with ID {item_id} not found.")
        _print_api_exception_details(e)
        sys.exit(1)
    except ApiException as e:
        _print_api_exception_details(e)
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred: {e}")
        sys.exit(1)

def handle_update_item(args):
    """Handles updating an existing item by its ID."""
    item_id = args.id

    try:
        # First, try to get the existing item
        existing_item = api.get_item_by_id(id=item_id)
    except NotFoundException:
        print(f"Error: Item with ID {item_id} not found. Cannot update.")
        sys.exit(1)
    except ApiException as e:
        print(f"Error fetching item with ID {item_id}:")
        _print_api_exception_details(e)
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred while fetching item {item_id}: {e}")
        sys.exit(1)

    # Determine values to update, using existing values as defaults
    # Name and Priority are required by UpdateItem, so they must have values.
    name_to_update = args.name if args.name is not None else existing_item.name
    priority_to_update = args.priority if args.priority is not None else existing_item.priority

    # Description can be explicitly set to None (or an empty string if the API treats it so)
    # If args.description is not provided (is None), we keep the existing description.
    # If args.description is provided (e.g., --description "" or --description "new desc"), we use it.
    description_to_update = existing_item.description # Default to existing
    if args.description is not None: # User provided --description argument
        description_to_update = args.description


    # At least one field must be provided for an update
    if args.name is None and args.description is None and args.priority is None:
        print("No update fields provided. To update an item, provide at least one of --name, --description, or --priority.")
        # Depending on API behavior, an empty description might be "" or null.
        # The UpdateItem model should clarify if description is nullable or just string.
        # Assuming UpdateItem's description can take None if the API supports it,
        # or an empty string if it should be cleared.
        # For this implementation, if --description is not passed, existing is used.
        # If --description "" is passed, description_to_update becomes "".
        # If API needs explicit null for description, that's a deeper model/API contract detail.
        print("If you meant to clear the description, use --description \"\" (an empty string).")
        sys.exit(0) # Not an error, but user intent is unclear or no change.


    update_payload = UpdateItem(
        name=name_to_update,
        priority=priority_to_update,
        description=description_to_update
    )

    try:
        updated_item = api.update_item_by_id(id=item_id, update_item=update_payload)
        print("Successfully updated item:")
        print(f"  ID: {updated_item.id}")
        print(f"  Name: {updated_item.name}")
        description = 'N/A'
        if hasattr(updated_item, 'description') and updated_item.description is not None:
            description = updated_item.description
        print(f"  Description: {description}")
        print(f"  Priority: {updated_item.priority}")
    except NotFoundException: # Should ideally not happen if the initial GET succeeded, but good practice
        print(f"Error: Item with ID {item_id} disappeared before update.")
        sys.exit(1)
    except BadRequestException as e:
        print(f"Error updating item {item_id} (Invalid Request):")
        _print_api_exception_details(e)
        sys.exit(1)
    except ApiException as e:
        print(f"Error updating item {item_id}:")
        _print_api_exception_details(e)
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred while updating item {item_id}: {e}")
        sys.exit(1)

def main():
    """Main function to parse arguments and call appropriate handlers."""
    parser = argparse.ArgumentParser(description="A simple CLI tool to interact with a TODO API.")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # List command
    list_parser = subparsers.add_parser("list", help="List all items")
    list_parser.add_argument("id", type=int, nargs="?", help="Optional ID of the item to list")

    # Create command
    create_parser = subparsers.add_parser("create", help="Create a new item")
    create_parser.add_argument("--name", required=True, help="Name of the item")
    create_parser.add_argument("--priority", required=True, type=int, help="Priority of the item")
    create_parser.add_argument("--description", help="Description of the item")

    # Delete command
    delete_parser = subparsers.add_parser("delete", help="Delete an item by ID")
    # The generated client expects path parameters as keyword arguments.
    # The 'id' here matches the parameter name in the OpenAPI spec for deleteItemById.
    delete_parser.add_argument("id", type=int, help="ID of the item to delete")

    # Update command
    update_parser = subparsers.add_parser("update", help="Update an existing item by ID")
    update_parser.add_argument("id", type=int, help="ID of the item to update")
    update_parser.add_argument("--name", type=str, help="New name for the item")
    update_parser.add_argument("--description", type=str, help="New description for the item")
    update_parser.add_argument("--priority", type=int, help="New priority for the item")

    args = parser.parse_args()

    if args.command == "list":
        handle_list_items(args)
    elif args.command == "create":
        handle_create_item(args)
    elif args.command == "delete":
        handle_delete_item(args)
    elif args.command == "update":
        handle_update_item(args)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
