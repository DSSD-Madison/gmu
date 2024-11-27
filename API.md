# API Documentation


## Filters API

### Overview

The Filters API enables clients to retrieve structured filter data categorized by Authors, File Type, and Region. This endpoint is ideal for search functionalities where filter options are dynamically generated.

### Base URL

```bash
http://localhost:8080/api
```

### Endpoints

1. GET /filters

#### Description

Retrieves a list of categorized filters. Each category contains multiple filter options with associated counts, representing the number of items matching that filter.

#### Request

##### HTTP Method

GET

##### Endpoint
```http
/filters
```
##### Headers

No custom headers required.

#### Response

##### HTTP Status Codes

- 200 OK: Request succeeded, and the response body contains the filters data.
- 500 Internal Server Error: Server encountered an error while processing the request.

##### Response Body Example
```json
{
  "Filters": [
    {
      "Category": "Authors",
      "Options": [
        {
          "Label": "Search for Common Ground (SFCG)",
          "Count": 35
        },
        {
          "Label": "The United States Agency for International Development (USAID)",
          "Count": 32
        },
        {
          "Label": "Mercy Corps",
          "Count": 8
        }
      ]
    },
    {
      "Category": "File Type",
      "Options": [
        {
          "Label": "PDF",
          "Count": 391
        },
        {
          "Label": "MS_WORD",
          "Count": 71
        }
      ]
    },
    {
      "Category": "Region",
      "Options": [
        {
          "Label": "Global",
          "Count": 391
        },
        {
          "Label": "Nepal",
          "Count": 71
        }
      ]
    }
  ]
}
```

#### Data Model

##### FilterCategory Object

Field | Type | Description
| ----------- | ----------- | ----------- |
Category | string | The name of the filter category.
Options | Array<FilterOption> | A list of filter options in the category.

##### FilterOption Object

Field | Type | Description
| ----------- | ----------- | ----------- |
Label | string | Name or description of the filter option.
Count | int | Number of items associated with this option.

#### Error Handling

##### 500 Internal Server Error

If the server fails to process the request, it will return:
```json
{
  "message": "Internal Server Error"
}
```

#### Examples

##### Request Example: Using curl
```bash
curl -X GET http://localhost:8080/api/filters
```
##### Request Example: Using Fetch API
```javascript
fetch('http://localhost:8080/api/filters')
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));
```
##### Request Example: Using Axios

import axios from 'axios';
```javascript
axios.get('http://localhost:8080/api/filters')
  .then(response => console.log(response.data))
  .catch(error => console.error('Error:', error));
```
#### Usage Considerations

- Dynamic Updates: The endpoint can be extended to retrieve data from a database or external API instead of hardcoded values.
- Caching: Consider implementing caching mechanisms for this endpoint if the filter data does not change frequently, improving performance for repeated requests.
- Security: Ensure proper input validation and error handling to safeguard against injection attacks and other vulnerabilities.
