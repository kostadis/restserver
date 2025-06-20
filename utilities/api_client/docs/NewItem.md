# NewItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** |  |
**description** | **str** |  | [optional]
**priority** | **int** |  |

## Example

```python
from todo_api_client.models.new_item import NewItem

# TODO update the JSON string below
json = "{}"
# create an instance of NewItem from a JSON string
new_item_instance = NewItem.from_json(json)
# print the JSON string representation of the object
print(NewItem.to_json())

# convert the object into a dict
new_item_dict = new_item_instance.to_dict()
# create an instance of NewItem from a dict
new_item_from_dict = NewItem.from_dict(new_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
