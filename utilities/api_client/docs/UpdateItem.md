# UpdateItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** |  |
**description** | **str** |  | [optional]
**priority** | **int** |  |

## Example

```python
from todo_api_client.models.update_item import UpdateItem

# TODO update the JSON string below
json = "{}"
# create an instance of UpdateItem from a JSON string
update_item_instance = UpdateItem.from_json(json)
# print the JSON string representation of the object
print(UpdateItem.to_json())

# convert the object into a dict
update_item_dict = update_item_instance.to_dict()
# create an instance of UpdateItem from a dict
update_item_from_dict = UpdateItem.from_dict(update_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
