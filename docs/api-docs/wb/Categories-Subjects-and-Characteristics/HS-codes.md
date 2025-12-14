HS-codes 

https://dev.wildberries.ru/en/openapi/work-with-products/#tag/Categories-Subjects-and-Characteristics/paths/~1content~1v2~1directory~1tnved/get
https://content-api.wildberries.ru/content/v2/directory/tnved

## Method description
The method provides list of HS-codes by category name and filter by HS-code.

Request limit per one seller's account for all methods in the Content category:
Period	Limit	Interval	Burst
1 minute	100 requests	600 milliseconds	5 requests
### Exceptions are the methods:

creating product cards
creating product cards with merge
editing product cards
getting failed product cards with errors

## query Parameters
subjectID
required
integer
Example: subjectID=105
Subject ID

search	
integer
Example: search=6106903000
Search by HS-code. Works only with the subjectID parameter

locale	
string
Example: locale=en
Language for response of the subjectName and name fields:

ru — Russian
en — English
zh — Chinese
Not used in the sandbox

## Response samples 200

{
  "data": [
    {
      "tnved": "6106903000",
      "isKiz": true
    }
  ],
  "error": false,
  "errorText": "",
  "additionalErrors": null
}

