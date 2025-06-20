# todo_api_client.DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**create_item**](DefaultApi.md#create_item) | **POST** /items | Create a new item
[**delete_item_by_id**](DefaultApi.md#delete_item_by_id) | **DELETE** /items/{id} | Delete an item by its ID
[**get_item_by_id**](DefaultApi.md#get_item_by_id) | **GET** /items/{id} | Get an item by its ID
[**get_items**](DefaultApi.md#get_items) | **GET** /items | List all items
[**update_item_by_id**](DefaultApi.md#update_item_by_id) | **PUT** /items/{id} | Update an existing item


# **create_item**
> Item create_item(new_item)

Create a new item

### Example


```python
import todo_api_client
from todo_api_client.models.item import Item
from todo_api_client.models.new_item import NewItem
from todo_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = todo_api_client.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with todo_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = todo_api_client.DefaultApi(api_client)
    new_item = todo_api_client.NewItem() # NewItem | Item to create

    try:
        # Create a new item
        api_response = api_instance.create_item(new_item)
        print("The response of DefaultApi->create_item:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DefaultApi->create_item: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **new_item** | [**NewItem**](NewItem.md)| Item to create |

### Return type

[**Item**](Item.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Item created successfully |  -  |
**400** | Invalid request payload |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_item_by_id**
> delete_item_by_id(id)

Delete an item by its ID

### Example


```python
import todo_api_client
from todo_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = todo_api_client.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with todo_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = todo_api_client.DefaultApi(api_client)
    id = 56 # int | ID of the item to delete

    try:
        # Delete an item by its ID
        api_instance.delete_item_by_id(id)
    except Exception as e:
        print("Exception when calling DefaultApi->delete_item_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| ID of the item to delete |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**204** | Item deleted successfully |  -  |
**400** | Invalid ID supplied |  -  |
**404** | Item not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_item_by_id**
> Item get_item_by_id(id)

Get an item by its ID

### Example


```python
import todo_api_client
from todo_api_client.models.item import Item
from todo_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = todo_api_client.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with todo_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = todo_api_client.DefaultApi(api_client)
    id = 56 # int | ID of the item to retrieve

    try:
        # Get an item by its ID
        api_response = api_instance.get_item_by_id(id)
        print("The response of DefaultApi->get_item_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DefaultApi->get_item_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| ID of the item to retrieve |

### Return type

[**Item**](Item.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Successful operation |  -  |
**400** | Invalid ID supplied |  -  |
**404** | Item not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_items**
> List[Item] get_items()

List all items

### Example


```python
import todo_api_client
from todo_api_client.models.item import Item
from todo_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = todo_api_client.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with todo_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = todo_api_client.DefaultApi(api_client)

    try:
        # List all items
        api_response = api_instance.get_items()
        print("The response of DefaultApi->get_items:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DefaultApi->get_items: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[Item]**](Item.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of items |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_item_by_id**
> Item update_item_by_id(id, update_item)

Update an existing item

### Example


```python
import todo_api_client
from todo_api_client.models.item import Item
from todo_api_client.models.update_item import UpdateItem
from todo_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = todo_api_client.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with todo_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = todo_api_client.DefaultApi(api_client)
    id = 56 # int | ID of the item to update
    update_item = todo_api_client.UpdateItem() # UpdateItem | Item data to update

    try:
        # Update an existing item
        api_response = api_instance.update_item_by_id(id, update_item)
        print("The response of DefaultApi->update_item_by_id:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling DefaultApi->update_item_by_id: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **int**| ID of the item to update |
 **update_item** | [**UpdateItem**](UpdateItem.md)| Item data to update |

### Return type

[**Item**](Item.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Item updated successfully |  -  |
**400** | Invalid request payload or input |  -  |
**404** | Item not found |  -  |
**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
