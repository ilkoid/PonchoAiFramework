# Products Parent Categories

https://dev.wildberries.ru/en/openapi/work-with-products/#tag/Categories-Subjects-and-Characteristics/paths/~1content~1v2~1object~1parent~1all/get

https://content-api.wildberries.ru/content/v2/object/parent/all

## Method description
Returns the list of all products parent categories

Request limit per one seller's account for all methods in the Content category:
Period	Limit	Interval	Burst
1 minute	100 requests	600 milliseconds	5 requests
Exceptions are the methods:

creating product cards
creating product cards with merge
editing product cards
getting failed product cards with errors
Authorizations:
HeaderApiKey

## query Parameters
locale	
string
Example: locale=en
Language for response of the name field:

ru — Russian
en — English
zh — Chinese
Not used in the sandbox

## Response samples 200:

Content type
application/json

{
"data": [
{
"name": "Электроника",
"id": 479,
"isVisible": true
}
],
"error": false,
"errorText": "",
"additionalErrors": ""
}
